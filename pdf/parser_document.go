package main

import (
	"fmt"
)

type DocumentParser struct {
	p   *ParserState
	Doc Document
}

func (dp *DocumentParser) Connect(p *ParserState) {
	dp.p = p
	dp.Doc = Document{
		Objects: map[Ref]PDFValue{},
	}
}

// Read an object number and generation from a given index.
func (dp *DocumentParser) parseRef(index int) Ref {
	return Ref{
		Num: dp.p.stack[index].(PDFInt).Val().(int64),
		Gen: dp.p.stack[index+1].(PDFInt).Val().(int64),
		Doc: &dp.Doc,
	}
}

func (dp *DocumentParser) Word(word string) {
	switch word {
	case "stream":
		fallthrough
	case "endstream":
		// ignore these; we already turned the stuff between them into StreamTokens

	case "R":
		// object reference: num gen R
		start := dp.p.len() - 2
		if start < 0 {
			panic(fmt.Errorf("Not enough items for object ref"))
		}
		ref := dp.parseRef(start)
		dp.p.dropFrom(start)
		dp.p.push(ref)

	case "obj":
		// num gen obj attrs? value endobj
		start := dp.p.len() - 2
		if start < 0 {
			panic(fmt.Errorf("Not enough items for object ref"))
		}
		dp.p.ctxPush(start)
	case "endobj":
		start := dp.p.ctxPop()
		length := dp.p.len() - start
		if length < 3 {
			panic(fmt.Errorf("Not enough items for object"))
		}
		if length > 4 {
			panic(fmt.Errorf("Too many items for object"))
		}

		ref := dp.parseRef(start)

		var obj PDFValue
		if length > 3 {
			// object is a stream with an attribute map
			obj = Stream{
				attrs: dp.p.stack[start+2].(PDFMap),
				buf:   dp.p.stack[start+3].(*StreamToken).buf,
			}
		} else {
			// object is a map, list, scalar, or ref
			obj = dp.p.stack[start+2].(PDFValue)
		}

		// add object to document object map
		dp.p.dropFrom(start)
		dp.Doc.Objects[ref] = obj

	case "xref":
		// look ahead to discover the number of xref entries
		xrefEntryCountIdx := dp.p.pos + 2
		if xrefEntryCountIdx >= len(dp.p.tokens) {
			panic(fmt.Errorf("xref table truncated by EOF"))
		}
		var xrefEntryCount int
		if xrefEntryCountToken, ok := dp.p.tokens[xrefEntryCountIdx].(*IntToken); ok {
			xrefEntryCount = int(xrefEntryCountToken.val)
		} else {
			panic(fmt.Errorf("Wrong type for xref entry count"))
		}
		// skip over the xref table, which is useless if reading the entire file
		dp.p.pos += 2 + xrefEntryCount*3

	case "trailer":
		// trailer attrs startxref xref_offset %%EOF
		dp.p.ctxPush(dp.p.len())
	case "startxref":
		start := dp.p.ctxPop()
		length := dp.p.len() - start
		if length < 1 {
			panic(fmt.Errorf("Not enough items for trailer"))
		}
		attrs := dp.p.stack[start].(PDFMap)

		// add to document trailers list
		dp.Doc.Trailers = append(dp.Doc.Trailers, attrs)
		dp.p.dropFrom(start)

		// skip over the xref table offset
		dp.p.pos++

	default:
		panic(fmt.Errorf("Unknown word %s", word))
	}
}
