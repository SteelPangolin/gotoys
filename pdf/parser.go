package main

import (
	"fmt"
)

type Document struct {
	Objects map[Ref]interface{}
	Attrs   map[string]interface{}
}

type Ref struct {
	Num, Gen int64
}

func (r Ref) String() string {
	return fmt.Sprintf("Ref %d %d", r.Num, r.Gen)
}

type Object struct {
	Attrs map[string]interface{}
}

// Read an object number and generation from a given index.
func parseRef(stack []interface{}, index int) (Ref, error) {
	ref := Ref{}
	if num, ok := stack[index].(int64); ok {
		ref.Num = num
	} else {
		return ref, fmt.Errorf("Wrong type for object ref number")
	}
	if gen, ok := stack[index+1].(int64); ok {
		ref.Gen = gen
	} else {
		return ref, fmt.Errorf("Wrong type for object ref generation")
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

func parse(tokens []Token) (Document, []interface{}, error) {
	doc := Document{
		Objects: map[Ref]interface{}{},
	}
	stack := []interface{}{}
	contextStack := []int{}
	for pos := 0; pos < len(tokens); pos++ {
		stack = append(stack, tokens[pos])
		end := len(stack) - 1

		switch last := stack[end].(type) {
		case *StreamToken:
			// TODO: drop for now
			stack = stack[:end]
		case *MetaToken:
			stack = stack[:end]

		case *IntToken:
			// unwrap
			stack[end] = last.val
		case *FloatToken:
			stack[end] = last.val
		case *SymbolToken:
			stack[end] = last.val
		case *StringToken:
			stack[end] = last.val
		case *HexToken:
			stack[end] = last.buf

		case *OperatorToken:
			switch last.op {
			case "[":
				ctxPush(&contextStack, end)
			case "]":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 2 {
					return doc, stack, fmt.Errorf("Not enough items for list")
				}

				list := []interface{}{}
				for i := start + 1; i < end; i++ {
					list = append(list, stack[i])
				}

				// put the new list on top of the stack
				stack = stack[:start+1]
				stack[start] = list

			case "<<":
				ctxPush(&contextStack, end)
			case ">>":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 2 {
					return doc, stack, fmt.Errorf("Not enough items for map")
				}
				if (end-start+1)%2 != 0 {
					return doc, stack, fmt.Errorf("Map needs an even number of items")
				}

				m := map[string]interface{}{}
				for i := start + 1; i < end; i += 2 {
					if key, ok := stack[i].(string); ok {
						m[key] = stack[i+1]
					} else {
						return doc, stack, fmt.Errorf("Map key %d must be a string", i)
					}
				}

				// put the new map on top of the stack
				stack = stack[:start+1]
				stack[start] = m

			default:
				return doc, stack, fmt.Errorf("Unknown operator %s", last.op)
			}

		case *WordToken:
			switch last.val {
			case "stream":
				fallthrough
			case "endstream":
				// ignore these
				stack = stack[:end]

			case "R":
				// object reference: num gen R
				if len(stack) < 3 {
					return doc, stack, fmt.Errorf("Not enough items for object ref")
				}
				ref, err := parseRef(stack, end-2)
				if err != nil {
					return doc, stack, err
				}
				stack[end-2] = ref
				stack = stack[:end-1]

			case "obj":
				// num gen obj attrs? stream? endobj
				ctxPush(&contextStack, end-2)
			case "endobj":
				start := ctxPop(&contextStack)
				if (end - start + 1) < 4 {
					return doc, stack, fmt.Errorf("Not enough items for object")
				}

				ref, err := parseRef(stack, start)
				if err != nil {
					return doc, stack, err
				}

				obj := Object{}
				if (end - start + 1) > 4 {
					// object probably has an attribute dict
					if attrs, ok := stack[start+3].(map[string]interface{}); ok {
						obj.Attrs = attrs
					} else {
						return doc, stack, fmt.Errorf("Wrong type for object attrs")
					}
				}

				// move object from stack to document
				stack = stack[:start]
				doc.Objects[ref] = obj

			case "xref":
				// look ahead to discover the number of xref entries
				xrefEntryCountIdx := pos + 2
				if xrefEntryCountIdx >= len(tokens) {
					return doc, stack, fmt.Errorf("xref table truncated by EOF")
				}
				var xrefEntryCount int
				if xrefEntryCountToken, ok := tokens[xrefEntryCountIdx].(*IntToken); ok {
					xrefEntryCount = int(xrefEntryCountToken.val)
				} else {
					return doc, stack, fmt.Errorf("Wrong type for xref entry count")
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
					return doc, stack, fmt.Errorf("Not enough items for trailer")
				}

				// set document attributes from trailer
				if attrs, ok := stack[start+1].(map[string]interface{}); ok {
					doc.Attrs = attrs
				} else {
					return doc, stack, fmt.Errorf("Wrong type for trailer attrs")
				}

				// remove trailer from stack
				stack = stack[:start]

				// skip over the xref table offset
				pos++

			default:
				return doc, stack, fmt.Errorf("Unknown word %s", last.val)
			}
		}
	}
	return doc, stack, nil
}
