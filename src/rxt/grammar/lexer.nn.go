package main

import (
	"fmt"
	"os"
)
import (
	"bufio"
	"io"
	"strings"
)

type frame struct {
	i            int
	s            string
	line, column int
}
type Lexer struct {
	// The lexer runs in its own goroutine, and communicates via channel 'ch'.
	ch      chan frame
	ch_stop chan bool
	// We record the level of nesting because the action could return, and a
	// subsequent call expects to pick up where it left off. In other words,
	// we're simulating a coroutine.
	// TODO: Support a channel-based variant that compatible with Go's yacc.
	stack []frame
	stale bool

	// The 'l' and 'c' fields were added for
	// https://github.com/wagerlabs/docker/blob/65694e801a7b80930961d70c69cba9f2465459be/buildfile.nex
	// Since then, I introduced the built-in Line() and Column() functions.
	l, c int

	parseResult interface{}

	// The following line makes it easy for scripts to insert fields in the
	// generated code.
	// [NEX_END_OF_LEXER_STRUCT]
}

// NewLexerWithInit creates a new Lexer object, runs the given callback on it,
// then returns it.
func NewLexerWithInit(in io.Reader, initFun func(*Lexer)) *Lexer {
	yylex := new(Lexer)
	if initFun != nil {
		initFun(yylex)
	}
	yylex.ch = make(chan frame)
	yylex.ch_stop = make(chan bool, 1)
	var scan func(in *bufio.Reader, ch chan frame, ch_stop chan bool, family []dfa, line, column int)
	scan = func(in *bufio.Reader, ch chan frame, ch_stop chan bool, family []dfa, line, column int) {
		// Index of DFA and length of highest-precedence match so far.
		matchi, matchn := 0, -1
		var buf []rune
		n := 0
		checkAccept := func(i int, st int) bool {
			// Higher precedence match? DFAs are run in parallel, so matchn is at most len(buf), hence we may omit the length equality check.
			if family[i].acc[st] && (matchn < n || matchi > i) {
				matchi, matchn = i, n
				return true
			}
			return false
		}
		var state [][2]int
		for i := 0; i < len(family); i++ {
			mark := make([]bool, len(family[i].startf))
			// Every DFA starts at state 0.
			st := 0
			for {
				state = append(state, [2]int{i, st})
				mark[st] = true
				// As we're at the start of input, follow all ^ transitions and append to our list of start states.
				st = family[i].startf[st]
				if -1 == st || mark[st] {
					break
				}
				// We only check for a match after at least one transition.
				checkAccept(i, st)
			}
		}
		atEOF := false
		stopped := false
		for {
			if n == len(buf) && !atEOF {
				r, _, err := in.ReadRune()
				switch err {
				case io.EOF:
					atEOF = true
				case nil:
					buf = append(buf, r)
				default:
					panic(err)
				}
			}
			if !atEOF {
				r := buf[n]
				n++
				var nextState [][2]int
				for _, x := range state {
					x[1] = family[x[0]].f[x[1]](r)
					if -1 == x[1] {
						continue
					}
					nextState = append(nextState, x)
					checkAccept(x[0], x[1])
				}
				state = nextState
			} else {
			dollar: // Handle $.
				for _, x := range state {
					mark := make([]bool, len(family[x[0]].endf))
					for {
						mark[x[1]] = true
						x[1] = family[x[0]].endf[x[1]]
						if -1 == x[1] || mark[x[1]] {
							break
						}
						if checkAccept(x[0], x[1]) {
							// Unlike before, we can break off the search. Now that we're at the end, there's no need to maintain the state of each DFA.
							break dollar
						}
					}
				}
				state = nil
			}

			if state == nil {
				lcUpdate := func(r rune) {
					if r == '\n' {
						line++
						column = 0
					} else {
						column++
					}
				}
				// All DFAs stuck. Return last match if it exists, otherwise advance by one rune and restart all DFAs.
				if matchn == -1 {
					if len(buf) == 0 { // This can only happen at the end of input.
						break
					}
					lcUpdate(buf[0])
					buf = buf[1:]
				} else {
					text := string(buf[:matchn])
					buf = buf[matchn:]
					matchn = -1
					for {
						sent := false
						select {
						case ch <- frame{matchi, text, line, column}:
							{
								sent = true
							}
						case stopped = <-ch_stop:
							{
							}
						default:
							{
								// nothing
							}
						}
						if stopped || sent {
							break
						}
					}
					if stopped {
						break
					}
					if len(family[matchi].nest) > 0 {
						scan(bufio.NewReader(strings.NewReader(text)), ch, ch_stop, family[matchi].nest, line, column)
					}
					if atEOF {
						break
					}
					for _, r := range text {
						lcUpdate(r)
					}
				}
				n = 0
				for i := 0; i < len(family); i++ {
					state = append(state, [2]int{i, 0})
				}
			}
		}
		ch <- frame{-1, "", line, column}
	}
	go scan(bufio.NewReader(in), yylex.ch, yylex.ch_stop, dfas, 0, 0)
	return yylex
}

type dfa struct {
	acc          []bool           // Accepting states.
	f            []func(rune) int // Transitions.
	startf, endf []int            // Transitions at start and end of input.
	nest         []dfa
}

var dfas = []dfa{
	// [#].*\n
	{[]bool{false, false, true, false}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 10:
				return -1
			case 35:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 10:
				return 2
			case 35:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 10:
				return 2
			case 35:
				return 3
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 10:
				return 2
			case 35:
				return 3
			}
			return 3
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// [\n][ \t\n]*
	{[]bool{false, true, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return 1
			case 32:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return 2
			case 10:
				return 2
			case 32:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return 2
			case 10:
				return 2
			case 32:
				return 2
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [ \t]+
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 9:
				return 1
			case 32:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return 1
			case 32:
				return 1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// AAA
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// DATASET
	{[]bool{false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 68:
				return 1
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 2
			case 68:
				return -1
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 4
			case 68:
				return -1
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 83:
				return 5
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return 6
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return 7
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// FOR SERVICE
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return 1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return 2
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return 3
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 4
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return 5
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return 6
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return 7
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return 8
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return 9
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return 10
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return 11
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 86:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// FOR STACK
	{[]bool{false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return 1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return 2
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return 3
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 4
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return 5
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 7
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return 8
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return 9
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 70:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// DEFINE AUTH
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return 1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return 2
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return 3
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return 4
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return 5
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return 6
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 7
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 8
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return 9
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return 10
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return 11
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// AS
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return 1
			case 83:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 83:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 83:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// SET
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 83:
				return 1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return 2
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// TO
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 79:
				return -1
			case 84:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 79:
				return 2
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 79:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// GET
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 71:
				return 1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return 2
			case 71:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 71:
				return -1
			case 84:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 71:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// POST
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 79:
				return -1
			case 80:
				return 1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 79:
				return 2
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return 3
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// FROM
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 70:
				return 1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 70:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 70:
				return -1
			case 77:
				return -1
			case 79:
				return 3
			case 82:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 70:
				return -1
			case 77:
				return 4
			case 79:
				return -1
			case 82:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 70:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// HELP
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 72:
				return 1
			case 76:
				return -1
			case 80:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return 2
			case 72:
				return -1
			case 76:
				return -1
			case 80:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 72:
				return -1
			case 76:
				return 3
			case 80:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 72:
				return -1
			case 76:
				return -1
			case 80:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 72:
				return -1
			case 76:
				return -1
			case 80:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// EXTRACT USING
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return 1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return 3
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return 4
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 5
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return 6
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return 7
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 8
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return 9
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return 10
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return 11
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return 12
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return 13
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 67:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// METRIC
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 77:
				return 1
			case 82:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return 2
			case 73:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 84:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 82:
				return 4
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 73:
				return 5
			case 77:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return 6
			case 69:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// NAME
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 77:
				return -1
			case 78:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 2
			case 69:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 77:
				return 3
			case 78:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return 4
			case 77:
				return -1
			case 78:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// TYPE
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 80:
				return -1
			case 84:
				return 1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 89:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 80:
				return 3
			case 84:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return 4
			case 80:
				return -1
			case 84:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 69:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 89:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// GAUGE
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 71:
				return 1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 2
			case 69:
				return -1
			case 71:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 85:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 71:
				return 4
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return 5
			case 71:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 69:
				return -1
			case 71:
				return -1
			case 85:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// COUNTER
	{[]bool{false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 67:
				return 1
			case 69:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 78:
				return -1
			case 79:
				return 2
			case 82:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			case 85:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 78:
				return 4
			case 79:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 84:
				return 5
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return 6
			case 78:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 82:
				return 7
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// HISTOGRAM
	{[]bool{false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return 1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 2
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return 3
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return 5
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return 6
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return 7
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 8
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return 9
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 77:
				return -1
			case 79:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// SUMMARY
	{[]bool{false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 83:
				return 1
			case 85:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 85:
				return 2
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return 3
			case 82:
				return -1
			case 83:
				return -1
			case 85:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return 4
			case 82:
				return -1
			case 83:
				return -1
			case 85:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 5
			case 77:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 85:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return -1
			case 82:
				return 6
			case 83:
				return -1
			case 85:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 85:
				return -1
			case 89:
				return 7
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 77:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 85:
				return -1
			case 89:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// DESCRIPTION
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return 1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 2
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return 3
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return 4
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 5
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return 6
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return 7
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return 8
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return 9
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return 10
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return 11
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// LABELS
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 76:
				return 1
			case 83:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return 2
			case 66:
				return -1
			case 69:
				return -1
			case 76:
				return -1
			case 83:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 66:
				return 3
			case 69:
				return -1
			case 76:
				return -1
			case 83:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 66:
				return -1
			case 69:
				return 4
			case 76:
				return -1
			case 83:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 76:
				return 5
			case 83:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 76:
				return -1
			case 83:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 65:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 76:
				return -1
			case 83:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// ,[ \n\t]*
	{[]bool{false, true, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 44:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return 2
			case 10:
				return 2
			case 32:
				return 2
			case 44:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return 2
			case 10:
				return 2
			case 32:
				return 2
			case 44:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// "[^"]*"
	{[]bool{false, false, true, false}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 34:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return 2
			}
			return 3
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// '[^']*'
	{[]bool{false, false, true, false}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 39:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 39:
				return 2
			}
			return 3
		},
		func(r rune) int {
			switch r {
			case 39:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 39:
				return 2
			}
			return 3
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// [a-z_][a-z0-9_]*
	{[]bool{false, true, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 95:
				return 1
			}
			switch {
			case 48 <= r && r <= 57:
				return -1
			case 97 <= r && r <= 122:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 95:
				return 2
			}
			switch {
			case 48 <= r && r <= 57:
				return 2
			case 97 <= r && r <= 122:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 95:
				return 2
			}
			switch {
			case 48 <= r && r <= 57:
				return 2
			case 97 <= r && r <= 122:
				return 2
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// .
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			return 1
		},
		func(r rune) int {
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},
}

func NewLexer(in io.Reader) *Lexer {
	return NewLexerWithInit(in, nil)
}

func (yyLex *Lexer) Stop() {
	yyLex.ch_stop <- true
}

// Text returns the matched text.
func (yylex *Lexer) Text() string {
	return yylex.stack[len(yylex.stack)-1].s
}

// Line returns the current line number.
// The first line is 0.
func (yylex *Lexer) Line() int {
	if len(yylex.stack) == 0 {
		return 0
	}
	return yylex.stack[len(yylex.stack)-1].line
}

// Column returns the current column number.
// The first column is 0.
func (yylex *Lexer) Column() int {
	if len(yylex.stack) == 0 {
		return 0
	}
	return yylex.stack[len(yylex.stack)-1].column
}

func (yylex *Lexer) next(lvl int) int {
	if lvl == len(yylex.stack) {
		l, c := 0, 0
		if lvl > 0 {
			l, c = yylex.stack[lvl-1].line, yylex.stack[lvl-1].column
		}
		yylex.stack = append(yylex.stack, frame{0, "", l, c})
	}
	if lvl == len(yylex.stack)-1 {
		p := &yylex.stack[lvl]
		*p = <-yylex.ch
		yylex.stale = false
	} else {
		yylex.stale = true
	}
	return yylex.stack[lvl].i
}
func (yylex *Lexer) pop() {
	yylex.stack = yylex.stack[:len(yylex.stack)-1]
}
func (yylex Lexer) Error(e string) {
	panic(e)
}

// Lex runs the lexer. Always returns 0.
// When the -s option is given, this function is not generated;
// instead, the NN_FUN macro runs the lexer.
func (yylex *Lexer) Lex(lval *yySymType) int {
OUTER0:
	for {
		switch yylex.next(0) {
		case 0:
			{ /* eat up comments */
			}
		case 1:
			{
				ii := indent(yylex.Text())
				fmt.Println("lexer -- ", mmm[ii.tkId], ii.tkVal, ii.tkId)
				lval.identVal = ii.tkVal
				return ii.tkId
			}
		case 2:
			{ /*fmt.Println("lexer -- ", "SPACES2", yylex.Text())*/ /* eat up whitespace */
			}
		case 3:
			{
				fmt.Println("lexer ------------- EOL")
				return EOL
			}
		case 4:
			{
				return DATASET
			}
		case 5:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return FOR_SERVICE
			}
		case 6:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return FOR_STACK
			}
		case 7:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return DEFINE_AUTH
			}
		case 8:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return AS
			}
		case 9:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return SET
			}
		case 10:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return TO
			}
		case 11:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return GET
			}
		case 12:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return POST
			}
		case 13:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return FROM
			}
		case 14:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return HELP
			}
		case 15:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return EXTRACT_USING
			}
		case 16:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return METRIC
			}
		case 17:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return NAME
			}
		case 18:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return TYPE
			}
		case 19:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return GAUGE
			}
		case 20:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return COUNTER
			}
		case 21:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return HISTOGRAM
			}
		case 22:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return SUMMARY
			}
		case 23:
			{
				fmt.Println("lexer -- ", yylex.Text())
				return DESCRIPTION
			}
		case 24:
			{ /*fmt.Println("lexer -- ", yylex.Text(), yylex.Text());*/
				return LABELS
			}
		case 25:
			{
				lval.strval = yylex.Text() /*fmt.Println("lexer -- ", "COMMA", yylex.Text());*/
				return COMMA
			}
		case 26:
			{
				lval.strval = yylex.Text()
				fmt.Println("lexer -- ", "STR_LITERAL", yylex.Text())
				return STR_LITERAL
			}
		case 27:
			{
				lval.strval = yylex.Text()
				fmt.Println("lexer -- ", yylex.Text())
				return RESOURCE_PATH
			}
		case 28:
			{
				fmt.Println("lexer -- ", "IDENTIFIER", yylex.Text())
				return IDENTIFIER
			}
		case 29:
			{
				fmt.Println("lexer -- ", "*", yylex.Text())
			}
		default:
			break OUTER0
		}
		continue
	}
	yylex.pop()

	return 0
}

type tkIdent struct {
	tkId  int
	tkVal int
}

var indent_level int
var indent_stack []int

func indent(whitespace string) tkIdent {
	level := len(whitespace) - 1
	idx_last_eol := strings.LastIndexByte(whitespace, 10)
	if idx_last_eol != -1 {
		level -= idx_last_eol
	}
	if level > indent_level {
		// Open block
		indent_stack = append(indent_stack, indent_level)
		indent_level = level
		return tkIdent{tkId: BLK, tkVal: indent_level}
	} else {
		if level == indent_level {
			// Same block
			return tkIdent{tkId: EOL, tkVal: indent_level}
		} else {
			// Close block
			if level == 0 && indent_level == 0 {
				fmt.Println("lexerrrrrrrrrrrrrrrrrrrrrrr")
				return tkIdent{tkId: EOB, tkVal: indent_level}
			}

			idx := len(indent_stack)
			for level < indent_level && idx > 0 {
				tt := tkIdent{tkId: EOB, tkVal: indent_level}
				idx = idx - 1
				indent_level = indent_stack[idx]
				return tt
			}
			indent_stack = indent_stack[:idx]
			if level == indent_level {
				return tkIdent{tkId: EOL, tkVal: indent_level}
			} else {
				return tkIdent{tkId: BIE, tkVal: indent_level}
			}
		}
	}
}

var mmm map[int]string

func main() {
	mmm = map[int]string{BIE: "BIE", BLK: "BLK", EOB: "EOB", EOL: "EOL", CTX: "CTX"}

	// fmt.Println("lexer -- ", "BIE", BIE)
	// fmt.Println("lexer -- ", "BLK", BLK)
	// fmt.Println("lexer -- ", "EOB", EOB)
	// fmt.Println("lexer -- ", "EOL", EOL)
	// fmt.Println("lexer -- ", "CTX", CTX)
	indent_level = 0
	indent_stack = make([]int, 5)
	yyErrorVerbose = true
	filename := "/usr/share/gocode/src/github.com/simelo/rextporter/src/rxt/testdata/skyexample.rxt"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	e := yyParse(NewLexer(file))
	fmt.Println("lexer -- ", "Return code:", e)
}
