package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
)

type Token interface {
	String() string
}

const (
	ModeStart = iota
	ModeMeta
	ModeInt
	ModeFloat
	ModeLT
	ModeGT
	ModeWord
	ModeSymbol
	ModeHex
	ModeStream
	ModeString
)

type OperatorToken struct {
	op string
}

func (t *OperatorToken) String() string {
	return t.op
}

type MetaToken struct {
	buf []byte
}

func (t *MetaToken) String() string {
	return fmt.Sprintf("Meta %q", t.buf)
}

type IntToken struct {
	buf []byte
}

func (t *IntToken) String() string {
	return fmt.Sprintf("Int %q", t.buf)
}

type FloatToken struct {
	buf []byte
}

func (t *FloatToken) String() string {
	return fmt.Sprintf("Float %q", t.buf)
}

type WordToken struct {
	buf []byte
}

func (t *WordToken) String() string {
	return fmt.Sprintf("Word %q", t.buf)
}

type SymbolToken struct {
	buf []byte
}

func (t *SymbolToken) String() string {
	return fmt.Sprintf("Symbol %q", t.buf)
}

type HexToken struct {
	buf []byte
}

func (t *HexToken) String() string {
	return fmt.Sprintf("Hex %q", t.buf)
}

type StreamToken struct {
	buf []byte
}

func (t *StreamToken) String() string {
	return fmt.Sprintf("Stream (%d bytes)", len(t.buf))
}

type StringToken struct {
	buf []byte
}

func (t *StringToken) String() string {
	return fmt.Sprintf("String %q", t.buf)
}

var AttrsStart = &OperatorToken{"<<"}
var AttrsEnd = &OperatorToken{">>"}
var ArrayStart = &OperatorToken{"["}
var ArrayEnd = &OperatorToken{"]"}

func lex(buf []byte) ([]Token, error) {
	mode := ModeStart
	tokenBuf := []byte{}
	tokens := []Token{}
	for pos := 0; pos < len(buf); {
		c := buf[pos]
		switch mode {
		case ModeStart:
			switch {
			case c == ' ' || c == '\t' || c == '\n' || c == '\r':
				pos++
			case c == '%':
				mode = ModeMeta
				pos++
			case c == '-':
				mode = ModeInt
				tokenBuf = append(tokenBuf, c)
				pos++
			case '0' <= c && c <= '9':
				mode = ModeInt
			case 'a' <= c && c <= 'z':
				fallthrough
			case 'A' <= c && c <= 'Z':
				mode = ModeWord
			case c == '<':
				mode = ModeLT
				pos++
			case c == '>':
				mode = ModeGT
				pos++
			case c == '[':
				tokens = append(tokens, ArrayStart)
				pos++
			case c == ']':
				tokens = append(tokens, ArrayEnd)
				pos++
			case c == '/':
				mode = ModeSymbol
				pos++
			case c == '(':
				mode = ModeString
				pos++
			default:
				return tokens, fmt.Errorf("ModeStart: Unexpected %q at pos %d\n", c, pos)
			}
		case ModeMeta:
			switch {
			case c == ' ' || c == '\t' || c == '\n' || c == '\r':
				tokens = append(tokens, &MetaToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
			default:
				tokenBuf = append(tokenBuf, c)
				pos++
			}
		case ModeInt:
			switch {
			case '0' <= c && c <= '9':
				tokenBuf = append(tokenBuf, c)
				pos++
			case c == '.':
				tokenBuf = append(tokenBuf, c)
				mode = ModeFloat
				pos++
			default:
				tokens = append(tokens, &IntToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeFloat:
			switch {
			case '0' <= c && c <= '9':
				tokenBuf = append(tokenBuf, c)
				pos++
			case c == '.':
				return tokens, fmt.Errorf("ModeFloat: Unexpected %q at pos %d\n", c, pos)
			default:
				tokens = append(tokens, &FloatToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeWord:
			switch {
			case 'a' <= c && c <= 'z':
				fallthrough
			case 'A' <= c && c <= 'Z':
				tokenBuf = append(tokenBuf, c)
				pos++
			default:
				word := &WordToken{tokenBuf}
				tokens = append(tokens, word)
				tokenBuf = []byte{}
				if bytes.Equal(word.buf, []byte("stream")) {
					mode = ModeStream
					// skip CRLF before start of stream data
					pos++
					pos++
				} else {
					mode = ModeStart
				}
			}
		case ModeLT:
			switch {
			case c == '<':
				tokens = append(tokens, AttrsStart)
				mode = ModeStart
				pos++
			case 'a' <= c && c <= 'f':
				fallthrough
			case '0' <= c && c <= '9':
				mode = ModeHex
			default:
				return tokens, fmt.Errorf("ModeLT: Unexpected %q at pos %d\n", c, pos)
			}
		case ModeGT:
			switch {
			case c == '>':
				tokens = append(tokens, AttrsEnd)
				mode = ModeStart
				pos++
			default:
				return tokens, fmt.Errorf("ModeGT: Unexpected %q at pos %d\n", c, pos)
			}
		case ModeHex:
			switch {
			case 'a' <= c && c <= 'f':
				fallthrough
			case '0' <= c && c <= '9':
				tokenBuf = append(tokenBuf, c)
				pos++
			case c == '>':
				tokens = append(tokens, &HexToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
				pos++
			default:
				return tokens, fmt.Errorf("ModeHex: Unexpected %q at pos %d\n", c, pos)
			}
		case ModeSymbol:
			switch {
			case 'a' <= c && c <= 'z':
				fallthrough
			case 'A' <= c && c <= 'Z':
				fallthrough
			case '0' <= c && c <= '9':
				fallthrough
			case c == '_' || c == '-' || c == ',':
				tokenBuf = append(tokenBuf, c)
				pos++
			default:
				tokens = append(tokens, &SymbolToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeStream:
			tokenBuf = append(tokenBuf, c)
			pos++
			if c == '\r' {
				endstream := []byte("\rendstream\r")
				tokenBufTail := tokenBuf[len(tokenBuf)-len(endstream):]
				if bytes.Equal(tokenBufTail, endstream) {
					tokenBuf = tokenBuf[:len(tokenBuf)-len(endstream)]
					tokens = append(tokens, &StreamToken{tokenBuf})
					tokens = append(tokens, &WordToken{[]byte("endstream")})
					tokenBuf = []byte{}
					mode = ModeStart
				}
			}
		case ModeString:
			switch {
			case c == ')':
				tokens = append(tokens, &StringToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
				pos++
			default:
				tokenBuf = append(tokenBuf, c)
				pos++
			}
		default:
			return tokens, fmt.Errorf("Mode %d: Unexpected %q at pos %d\n", mode, c, pos)
		}
	}
	return tokens, nil
}

func main() {
	flag.Parse()
	path := flag.Arg(0)

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	tokens, err := lex(buf)
	for _, token := range tokens {
		fmt.Printf("%s\n", token)
	}
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
}
