package grammar

import "strings"

const (
	// AsciiEOL code point
	AsciiEOL = 10
)

type intToken struct {
	tokTyp string
	tokVal int
}

type lexerState struct {
	handler     TokenHandler
	indentStack []int
	indentLevel int
	rootEnv     interface{}
	pendIndent  []intToken
	nextIndent  int
	started     bool
	retVal      int
}

func (state *lexerState) indent(whitespace string) {
	state.pendIndent = nil
	state.nextIndent = 0
	emitInt := func(tokTyp string, tokVal int) {
		tok := intToken{
			tokTyp: tokTyp,
			tokVal: tokVal,
		}
		state.pendIndent = append(state.pendIndent, tok)
	}
	level := len(whitespace) - 1
	// Consider whitespace since last line break
	idxLastEol := strings.LastIndexByte(whitespace, AsciiEOL)
	if idxLastEol != -1 {
		level -= idxLastEol
	}
	if level > state.indentLevel {
		// Open block
		state.indentStack = append(state.indentStack, state.indentLevel)
		state.indentLevel = level
		emitInt("BLK", state.indentLevel)
	} else {
		if level == state.indentLevel {
			// Same block
			emitInt("EOL", state.indentLevel)
		} else {
			// Close block
			if level == 0 && state.indentLevel == 0 {
				return
			}

			idx := len(state.indentStack)
			for level < state.indentLevel && idx > 0 {
				emitInt("EOB", state.indentLevel)
				idx = idx - 1
				state.indentLevel = state.indentStack[idx]
			}
			state.indentStack = state.indentStack[:idx]
			if level == state.indentLevel {
				emitInt("EOL", state.indentLevel)
			} else {
				emitInt("BIE", state.indentLevel)
			}
		}
	}
}
