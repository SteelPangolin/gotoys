package main

import (
	"fmt"
)

type Document struct {
	Objects map[Ref]PDFValue
	Attrs   PDFMap
}

// An object reference can be used almost anywhere in a PDF
type PDFValue interface {
	Val() interface{}
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

// PDF stream object. Bytes, but may be compressed or otherwise encoded.
type Stream struct {
	buf   []byte
	attrs PDFMap
}

func (o Stream) String() string {
	return fmt.Sprintf("Stream (%d bytes)", len(o.buf))
}

func (o Stream) Val() interface{} {
	return nil
}

// Read an object number and generation from a given index.
func parseRef(doc *Document, stack []interface{}, index int) (Ref, error) {
	ref := Ref{
		Num: stack[index].(PDFInt).Val().(int64),
		Gen: stack[index+1].(PDFInt).Val().(int64),
		Doc: doc,
	}
	return ref, nil
}

func ctxPush(contextStack *[]int, index int) {
	*contextStack = append(*contextStack, index)
}

func ctxPop(contextStack *[]int) int {
	ctxEnd := len(*contextStack) - 1
	index := (*contextStack)[ctxEnd]
	*contextStack = (*contextStack)[:ctxEnd]
	return index
}

func parse(tokens []Token) (doc Document, err error) {
	doc = Document{
		Objects: map[Ref]PDFValue{},
	}
	stack := []interface{}{}
	defer func() {
		e := recover()
		if e != nil {
			if len(stack) > 0 {
				fmt.Printf("stack = %v\n", stack)
			}
			panic(e)
			// TODO: more useful error object
		}
	}()
	contextStack := []int{}
	for pos := 0; pos < len(tokens); pos++ {
		stack = append(stack, tokens[pos])
		end := len(stack) - 1

		switch last := stack[end].(type) {
		// TODO: drop comments for now
		case *MetaToken:
			stack = stack[:end]

		// wrap primitives
		case *IntToken:
			stack[end] = PDFInt(last.val)
		case *FloatToken:
			stack[end] = PDFFloat(last.val)
		case *SymbolToken:
			stack[end] = PDFString(last.val)
		case *StringToken:
			stack[end] = PDFString(last.val)
		case *HexToken:
			stack[end] = PDFString(last.buf)

		case *StreamToken:
			// leave as is

		case *OperatorToken:
			switch last.op {
			// list
			case "[":
				ctxPush(&contextStack, end)
			case "]":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 2 {
					panic(fmt.Errorf("Not enough items for list"))
				}

				list := PDFList{}
				for i := start + 1; i < end; i++ {
					list = append(list, stack[i].(PDFValue))
				}

				// put the new list on top of the stack
				stack = stack[:start+1]
				stack[start] = list

			// map
			case "<<":
				ctxPush(&contextStack, end)
			case ">>":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 2 {
					panic(fmt.Errorf("Not enough items for map"))
				}
				if (end-start+1)%2 != 0 {
					panic(fmt.Errorf("Map needs an even number of items"))
				}

				m := PDFMap{}
				for i := start + 1; i < end; i += 2 {
					key := stack[i].(PDFValue).Val().(string)
					m[key] = stack[i+1].(PDFValue)
				}

				// put the new map on top of the stack
				stack = stack[:start+1]
				stack[start] = m

			default:
				panic(fmt.Errorf("Unknown operator %s", last.op))
			}

		case *WordToken:
			switch last.val {
			case "stream":
				fallthrough
			case "endstream":
				// ignore these; we already turned the stuff between them into StreamTokens
				stack = stack[:end]

			case "R":
				// object reference: num gen R
				start := end - 2
				if start < 0 {
					panic(fmt.Errorf("Not enough items for object ref"))
				}
				ref, err := parseRef(&doc, stack, start)
				if err != nil {
					return doc, err
				}
				stack = stack[:start+1]
				stack[start] = ref

			case "obj":
				// num gen obj attrs? value endobj
				start := end - 2
				if start < 0 {
					panic(fmt.Errorf("Not enough items for object ref"))
				}
				ctxPush(&contextStack, start)
			case "endobj":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 5 {
					panic(fmt.Errorf("Not enough items for object"))
				}
				if (end - start + 1) > 6 {
					panic(fmt.Errorf("Too many items for object"))
				}

				ref, err := parseRef(&doc, stack, start)
				if err != nil {
					return doc, err
				}

				var obj PDFValue
				if (end - start + 1) > 5 {
					// object is a stream with an attribute map
					obj = Stream{
						attrs: stack[start+3].(PDFMap),
						buf:   stack[start+4].(*StreamToken).buf,
					}
				} else {
					// object is a map, list, scalar, or ref
					obj = stack[start+3].(PDFValue)
				}

				// move object from stack to document
				stack = stack[:start]
				doc.Objects[ref] = obj

			case "xref":
				// look ahead to discover the number of xref entries
				xrefEntryCountIdx := pos + 2
				if xrefEntryCountIdx >= len(tokens) {
					panic(fmt.Errorf("xref table truncated by EOF"))
				}
				var xrefEntryCount int
				if xrefEntryCountToken, ok := tokens[xrefEntryCountIdx].(*IntToken); ok {
					xrefEntryCount = int(xrefEntryCountToken.val)
				} else {
					panic(fmt.Errorf("Wrong type for xref entry count"))
				}
				// skip over the xref table, which is useless if reading the entire file
				stack = stack[:end]
				pos += 2 + xrefEntryCount*3

			case "trailer":
				// trailer attrs startxref xref_offset %%EOF
				ctxPush(&contextStack, end)
			case "startxref":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 3 {
					panic(fmt.Errorf("Not enough items for trailer"))
				}

				// set document attributes from trailer
				doc.Attrs = stack[start+1].(PDFMap)

				// remove trailer from stack
				stack = stack[:start]

				// skip over the xref table offset
				pos++

			default:
				panic(fmt.Errorf("Unknown word %s", last.val))
			}
		}
	}
	return doc, nil
}
