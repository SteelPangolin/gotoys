package main

import ()

type Command struct {
	Word     string
	Operands []interface{}
}

type ContentParser struct {
	p        *ParserState
	Commands []Command
}

func (cp *ContentParser) Connect(p *ParserState) {
	cp.p = p
}

func (cp *ContentParser) Word(word string) {
	command := Command{
		Word: word,
	}
	for _, operand := range cp.p.stack {
		command.Operands = append(command.Operands, operand)
	}
	cp.p.dropFrom(0)
	cp.Commands = append(cp.Commands, command)
}
