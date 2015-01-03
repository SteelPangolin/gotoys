package main

import (
	"fmt"
)

type WordHandler interface {
	Connect(p *ParserState)
	Word(word string)
}

type Document struct {
	Objects map[Ref]PDFValue
	// linearized PDFs and incrementally updated PDFs have multiple trailers
	Trailers []PDFMap
}

// An object reference can be used almost anywhere in a PDF
type PDFValue interface {
	Val() interface{}
}

type PDFNull struct{}

func (o PDFNull) Val() interface{} {
	return nil
}

type PDFBool bool

func (o PDFBool) Val() interface{} {
	return bool(o)
}

type PDFInt int64

func (o PDFInt) Val() interface{} {
	return int64(o)
}

type PDFFloat float64

func (o PDFFloat) Val() interface{} {
	return float64(o)
}

type PDFString string

func (o PDFString) Val() interface{} {
	return string(o)
}

type PDFList []PDFValue

func (o PDFList) Val() interface{} {
	return o
}

// TODO: I really hope map keys can't be references
type PDFMap map[string]PDFValue

func (o PDFMap) Val() interface{} {
	return o
}

type Ref struct {
	Doc      *Document
	Num, Gen int64
}

func (r Ref) String() string {
	return fmt.Sprintf("Ref %d %d", r.Num, r.Gen)
}

func (r Ref) Val() interface{} {
	if obj, ok := r.Doc.Objects[r]; ok {
		return obj.Val()
	}
	panic(fmt.Errorf("Missing object reference: %s", r))
	// TODO: PDF standard says dangling refs should be treated as nulls, not errors
}

type ParserState struct {
	tokens       []Token
	pos          int
	stack        []interface{}
	contextStack []int
	// TODO: context stack doesn't check the kind of context (map, list, object)
}

func (p *ParserState) len() int {
	return len(p.stack)
}

func (p *ParserState) push(o interface{}) {
	p.stack = append(p.stack, o)
}

func (p *ParserState) dropFrom(index int) {
	p.stack = p.stack[:index]
}

func (p *ParserState) ctxPush(index int) {
	p.contextStack = append(p.contextStack, index)
}

func (p *ParserState) ctxPop() int {
	ctxEnd := len(p.contextStack) - 1
	index := p.contextStack[ctxEnd]
	p.contextStack = p.contextStack[:ctxEnd]
	return index
}

func parse(tokens []Token, wh WordHandler) {
	p := &ParserState{
		tokens: tokens,
	}
	wh.Connect(p)

	// on panic, print stack and re-panic
	defer func() {
		e := recover()
		if e != nil {
			if len(p.stack) > 0 {
				fmt.Printf("stack = %v\n", p.stack)
				fmt.Printf("contextStack = %v\n", p.contextStack)
			}
			panic(e)
		}
	}()

	for ; p.pos < len(tokens); p.pos++ {
		switch token := tokens[p.pos].(type) {

		case *MetaToken:
			// TODO: drop comments for now

		// wrap primitives
		case *IntToken:
			p.push(PDFInt(token.val))
		case *FloatToken:
			p.push(PDFFloat(token.val))
		case *SymbolToken:
			p.push(PDFString(token.val))
		case *StringToken:
			p.push(PDFString(token.val))
		case *HexToken:
			p.push(PDFString(token.buf))

		case *StreamToken:
			// leave as is
			p.push(token)

		case *OperatorToken:
			switch token.op {
			// list
			case "[":
				p.ctxPush(p.len())
			case "]":
				start := p.ctxPop()
				length := p.len() - start
				if length < 0 {
					panic(fmt.Errorf("Not enough items for list"))
				}

				list := PDFList{}
				for i := start + 1; i < p.len(); i++ {
					list = append(list, p.stack[i].(PDFValue))
				}

				p.dropFrom(start)
				p.push(list)

			// map
			case "<<":
				p.ctxPush(p.len())
			case ">>":
				start := p.ctxPop()
				length := p.len() - start
				if length < 2 {
					panic(fmt.Errorf("Not enough items for map"))
				}
				if length%2 != 0 {
					panic(fmt.Errorf("Map needs an even number of items"))
				}

				m := PDFMap{}
				for i := start; i < p.len(); i += 2 {
					key := p.stack[i].(PDFValue).Val().(string)
					m[key] = p.stack[i+1].(PDFValue)
				}

				p.dropFrom(start)
				p.push(m)

			default:
				panic(fmt.Errorf("Unknown operator %s", token.op))
			}

		case *WordToken:
			switch token.val {
			// more primitives
			case "null":
				p.push(PDFNull{})
			case "false":
				p.push(PDFBool(false))
			case "true":
				p.push(PDFBool(true))

			default:
				wh.Word(token.val)
			}
		}
	}
}
