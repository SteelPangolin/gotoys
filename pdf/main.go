package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

var streamBytes []byte = []byte("stream")
var endstreamBytes []byte = []byte("endstream")

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
				tokens = append(tokens, &IntToken{tokenBuf})
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
				tokens = append(tokens, &FloatToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
			}
		case ModeWord:
			switch {
			case isLetter(c):
				tokenBuf = append(tokenBuf, c)
				pos++
			default:
				word := &WordToken{tokenBuf}
				tokens = append(tokens, word)
				tokenBuf = []byte{}
				if bytes.Equal(word.buf, streamBytes) {
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
			case c == '_' || c == '-' || c == ',':
				tokenBuf = append(tokenBuf, c)
				pos++
			default:
				tokens = append(tokens, &SymbolToken{tokenBuf})
				tokenBuf = []byte{}
				mode = ModeStart
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
					if bytes.HasPrefix(buf[lPos:], endstreamBytes) {
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
			return tokens, fmt.Errorf("Mode %d: Unexpected %q at pos %d", mode, c, pos)
		}
	}
	if mode != ModeStart {
		return tokens, fmt.Errorf("Mode %d: Unfinished business", mode)
	}
	return tokens, nil
}

func saveStream(stream *StreamToken, path string) error {
	// TODO: check the filter type instead of assuming Flate
	buf := bytes.NewBuffer(stream.buf)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
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
		fmt.Printf("lexer error: %s\n", err)
	}

	streamIdx := 1
	for _, token := range tokens {
		if stream, ok := token.(*StreamToken); ok {
			streamPath := fmt.Sprintf("stream_%04d.dat", streamIdx)
			streamIdx++
			err := saveStream(stream, streamPath)
			if err != nil {
				fmt.Printf("error saving stream %s: %v\n", streamPath, err)
			}
		}
	}
}
