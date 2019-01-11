package grammar

import "os"
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

	// ,[ \n\t]*
	{[]bool{false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
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
			case 75:
				return -1
			case 76:
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
			case 75:
				return -1
			case 76:
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
			case 75:
				return -1
			case 76:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 75:
				return -1
			case 76:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 75:
				return 5
			case 76:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return 6
			case 75:
				return -1
			case 76:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 75:
				return -1
			case 76:
				return 7
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 75:
				return 8
			case 76:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

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
			case 66:
				return -1
			case 69:
				return -1
			case 79:
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
			case 66:
				return -1
			case 69:
				return 3
			case 79:
				return -1
			switch r {
			case 9:
				return 2
			case 10:
				return 2
			case 32:
				return 2
			case 66:
				return -1
			case 69:
				return 3
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 79:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return 5
			case 69:
				return -1
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 69:
				return 6
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 79:
				return 7
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return 8
			case 69:
				return -1
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 66:
				return -1
			case 69:
				return -1
			case 79:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},
				return -1
			case 10:
			case 32:
				return -1
			case 69:
				return -1
			case 76:
				return -1
			case 79:
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
			case 32:
				return 2
			case 69:
				return 3
			case 76:
				return -1
			case 79:
				return -1
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
			case 32:
				return 2
			case 69:
			case 76:
				return -1
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return 1
			case 32:
				return 1
			case 32:
				return -1
			case 69:
			case 76:
				return -1
			case 79:
				return 4
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 69:
				return -1
			case 76:
				return 5
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 69:
				return 6
			case 76:
				return -1
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 69:
				return -1
			case 76:
				return -1
			case 79:
				return 7
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 69:
				return -1
			case 76:
				return 8
			case 79:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 10:
				return -1
			case 32:
				return -1
			case 69:
				return -1
			case 76:
				return -1
			case 79:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

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

	// DATASET|FOR SERVICE|FOR STACK|DEFINE AUTH|AS|SET|TO|GET|POST|FROM|EXTRACT USING|METRIC|NAME|TYPE|GAUGE|COUNTER|HISTOGRAM|SUMMARY|DESCRIPTION|LABELS
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, true, false, false, false, false, false, false, true, true, false, false, true, false, false, true, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, true, false, false, true, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, true, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 1
			case 66:
				return -1
			case 67:
				return 2
			case 68:
				return 3
			case 69:
				return 4
			case 70:
				return 5
			case 71:
				return 6
			case 72:
				return 7
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return 8
			case 77:
				return 9
			case 78:
				return 10
			case 79:
				return -1
			case 80:
				return 11
			case 82:
				return -1
			case 83:
				return 12
			case 84:
				return 13
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 116
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 110
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 85
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 86
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return 73
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 56
			case 80:
				return -1
			case 82:
				return 57
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 50
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 51
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 42
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 37
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 32
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 29
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 26
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 18
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return 19
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 14
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return 15
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return 16
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 17
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 25
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return 20
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return 21
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 22
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 23
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return 24
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 27
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 28
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return 30
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 31
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 33
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 34
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 35
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return 36
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return 38
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 39
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return 40
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 41
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 43
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 44
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 45
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return 46
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 47
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 48
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return 49
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return 53
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 52
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return 54
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 55
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 60
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 58
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return 59
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 61
			case 65:
				return -1
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 62
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 63
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 64
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 68
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 65
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return 66
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return 67
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return 69
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 70
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return 71
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 72
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 74
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 75
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 76
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return 77
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 78
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 79
			case 65:
				return -1
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return 80
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 81
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
	// WITH OPTIONS
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return 2
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return 3
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return 4
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 5
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return 6
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return 7
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return 8
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return 9
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return 10
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return 11
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return 12
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 87:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return -1
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 82
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return 83
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return 84
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 105
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return 87
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 88
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 97
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return 89
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 90
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 91
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return 92
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 93
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return 94
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return 95
			case 80:
				return -1
			case 82:
				return -1
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return 96
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return 98
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 99
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return 100
			case 65:
				return -1
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 101
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return 102
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 103
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return 104
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 32:
				return -1
			case 65:
				return 106
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 107
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 108
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 109
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return 111
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return 112
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
				return 113
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return 114
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
				return -1
			case 78:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 82:
				return 115
			case 83:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
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
			case 66:
				return -1
			case 67:
				return -1
			case 68:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 71:
				return -1
			case 72:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 76:
				return -1
			case 77:
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
			case 85:
				return -1
			case 86:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

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
func (lex *Lexer) emitIndentTokens(state *lexerState) int {
	for state.nextIndent < len(state.pendIndent) {
		tok := &state.pendIndent[state.nextIndent]
		state.nextIndent++
		if tokId := state.handler.EmitInt(tok.tokTyp, tok.tokVal); tokId != 0 {
			return tokId
				return EOB
		}
	}
	state.pendIndent = nil
	state.nextIndent = 0
	return 0
}

func (lex *Lexer) Lex(state *lexerState) int {
	// Clear return value
	// Emit any pending indentation tokens
	token := lex.Text
	emitStr := state.handler.EmitStr
	emitObj := state.handler.EmitObj
	emitErr := state.handler.Error
	emitInd := func() int { return lex.emitIndentTokens(state) }
	indent := state.indent
	retVal := func(value int) int { state.retVal = value; return value }
	if tokId := emitInd(); tokId != 0 {
		return tokId
	}
	func(yylex *Lexer) {
		if !yylex.stale {
				return TO
				if !state.started {
					state.started = true
					if retVal(emitObj("CTX", state.rootEnv)) != 0 {
						return
					}
				}
			}
		}
	OUTER0:
		for {
			switch yylex.next(0) {
			case 0:
				{ /* eat up comments */
				}
			case 1:
				{
					if retVal(emitStr("PNC", token()[:1])) != 0 {
						return
					}
				}
			case 2:
				{
					indent(token())
					if retVal(emitInd()) != 0 {
						return
					}
				}
			case 3:
				{ /* eat up whitespace */
				}
			case 4:
				{
					if retVal(emitStr("KEY", token())) != 0 {
						return
					}
				}
			case 5:
				{
					if retVal(emitStr("STR", token())) != 0 {
						return
					}
				}
			case 6:
				{
					if retVal(emitStr("STR", token())) != 0 {
						return
					}
				}
			case 7:
				{
					if retVal(emitStr("VAR", token())) != 0 {
						return
					}
				}
			case 8:
				{
					emitErr("Unexpected token " + token())
				return RESOURCE_PATH
			}
		case 30:
			{ /*fmt.Println(yylex.Line() + 1, "lexer -- ", "IDENTIFIER", yylex.Text());*/
				return IDENTIFIER
			}
		case 31:
			{ /*fmt.Println(yylex.Line() + 1, "lexer -- ", "*", yylex.Text())*/
				}
			default:
				break OUTER0
			}
			continue
		}
		yylex.pop()
		{ /* nothing to do at end of file */
		}
	}(lex)
	return state.retVal
}

func (lex Lexer) Error(e string) {
	panic(e)
}

func LexTheRxt(handler TokenHandler, rootEnv interface{}) {
	lex := NewLexer(os.Stdin)
	tokenId := 1
	state := lexerState{
		handler: handler,
	}
	for tokenId > 0 {
		tokenId = lex.Lex(&state)
	}
}
