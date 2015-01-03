package main

import (
	"bytes"
	"fmt"
	"strconv"
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
	ModeStreamStart
	ModeString
	ModeStringEscape
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
	val int64
}

func (t *IntToken) String() string {
	return fmt.Sprintf("Int %d", t.val)
}

type FloatToken struct {
	val float64
	// TODO: retain original, or maybe number of digits, for round trips
}

func (t *FloatToken) String() string {
	return fmt.Sprintf("Float %f", t.val)
}

type WordToken struct {
	val string
}

func (t *WordToken) String() string {
	return fmt.Sprintf("Word %q", t.val)
}

type SymbolToken struct {
	val string
}

func (t *SymbolToken) String() string {
	return fmt.Sprintf("Symbol %q", t.val)
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
	val string
}

func (t *StringToken) String() string {
	return fmt.Sprintf("String %q", t.val)
}

var AttrsStart = &OperatorToken{"<<"}
var AttrsEnd = &OperatorToken{">>"}
var ArrayStart = &OperatorToken{"["}
var ArrayEnd = &OperatorToken{"]"}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isLetter(c byte) bool {
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}

func isHexDigit(c byte) bool {
	return isDigit(c) || ('A' <= c && c <= 'F') || ('a' <= c && c <= 'f')
}

func lex(buf []byte) ([]Token, error) {
	mode := ModeStart
	tokenBuf := []byte{}
	tokens := []Token{}
	for pos := 0; pos < len(buf); {
		c := buf[pos]
		switch mode {
		case ModeStart:
			switch {
			case isWhitespace(c):
				pos++
			case c == '%':
				mode = ModeMeta
				pos++
			case c == '-':
				mode = ModeInt
				tokenBuf = append(tokenBuf, c)
				pos++
			case isDigit(c):
				mode = ModeInt
			case isLetter(c):
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
				return tokens, fmt.Errorf("ModeStart: Unexpected %q at pos %d", c, pos)
			}
		case ModeMeta:
			switch {
			case isWhitespace(c):
				tokens = append(tokens, &MetaToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
			default:
				tokenBuf = append(tokenBuf, c)
				pos++
			}
		case ModeInt:
			switch {
			case isDigit(c):
				tokenBuf = append(tokenBuf, c)
				pos++
			case c == '.':
				tokenBuf = append(tokenBuf, c)
				mode = ModeFloat
				pos++
			default:
				value, err := strconv.ParseInt(string(tokenBuf), 10, 64)
				if err != nil {
					return tokens, err
				}
				tokens = append(tokens, &IntToken{value})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeFloat:
			switch {
			case isDigit(c):
				tokenBuf = append(tokenBuf, c)
				pos++
			case c == '.':
				return tokens, fmt.Errorf("ModeFloat: Unexpected %q at pos %d", c, pos)
			default:
				value, err := strconv.ParseFloat(string(tokenBuf), 64)
				if err != nil {
					return tokens, err
				}
				tokens = append(tokens, &FloatToken{value})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeWord:
			switch {
			case isLetter(c) || c == '*':
				tokenBuf = append(tokenBuf, c)
				pos++
			default:
				word := &WordToken{string(tokenBuf)}
				tokens = append(tokens, word)
				tokenBuf = []byte{}
				if word.val == "stream" {
					mode = ModeStreamStart
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
			case isHexDigit(c):
				mode = ModeHex
			default:
				return tokens, fmt.Errorf("ModeLT: Unexpected %q at pos %d", c, pos)
			}
		case ModeGT:
			switch {
			case c == '>':
				tokens = append(tokens, AttrsEnd)
				mode = ModeStart
				pos++
			default:
				return tokens, fmt.Errorf("ModeGT: Unexpected %q at pos %d", c, pos)
			}
		case ModeHex:
			switch {
			case isHexDigit(c):
				tokenBuf = append(tokenBuf, c)
				pos++
			case c == '>':
				tokens = append(tokens, &HexToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
				pos++
			default:
				return tokens, fmt.Errorf("ModeHex: Unexpected %q at pos %d", c, pos)
			}
		case ModeSymbol:
			switch {
			case isLetter(c):
				fallthrough
			case isDigit(c):
				fallthrough
			case c == '_' || c == '-' || c == ',' || c == '*':
				tokenBuf = append(tokenBuf, c)
				pos++
			default:
				tokens = append(tokens, &SymbolToken{string(tokenBuf)})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeString:
			switch {
			case c == '\\':
				mode = ModeStringEscape
				pos++
			case c == ')':
				tokens = append(tokens, &StringToken{string(tokenBuf)})
				tokenBuf = []byte{}
				mode = ModeStart
				pos++
			default:
				tokenBuf = append(tokenBuf, c)
				pos++
			}
		case ModeStringEscape:
			switch {
			case isWhitespace(c):
				mode = ModeString
				pos++
			default:
				tokenBuf = append(tokenBuf, c)
				mode = ModeString
				pos++
			}
		case ModeStreamStart:
			switch {
			case isWhitespace(c):
				pos++
			default:
				mode = ModeStream
			}
		case ModeStream:
			// TODO: doesn't use object's stream length info
		StreamSwitch:
			switch {
			case isWhitespace(c):
				// look ahead for endstream token
				for lPos := pos + 1; ; lPos++ {
					if lPos >= len(buf) {
						return tokens, fmt.Errorf("ModeStream: EOF while looking for endstream from pos %d", pos)
					}
					if isWhitespace(buf[lPos]) {
						continue
					}
					if bytes.HasPrefix(buf[lPos:], []byte("endstream")) {
						tokens = append(tokens, &StreamToken{tokenBuf})
						tokenBuf = []byte{}
						mode = ModeStart
						break StreamSwitch
					}
					// if we don't see an endstream, c is part of the stream data
					break
				}
				fallthrough
			default:
				tokenBuf = append(tokenBuf, c)
				pos++
			}
		default:
			return tokens, fmt.Errorf("Mode %d: Unexpected %q at pos %d", mode, c, pos)
		}
	}
	if mode != ModeStart {
		return tokens, fmt.Errorf("Mode %d: Unfinished business", mode)
	}
	return tokens, nil
}
