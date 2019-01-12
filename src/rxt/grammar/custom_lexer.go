package grammar

import (
	"fmt"
	"log"
)

var tks2id = map[string]int{
	"GET":           GET,
	"POST":          POST,
	"COUNTER":       COUNTER,
	"GAUGE":         GAUGE,
	"HISTOGRAM":     HISTOGRAM,
	"SUMMARY":       SUMMARY,
	"IDENTIFIER":    IDENTIFIER,
	"STR LITERAL":   STR_LITERAL,
	"RESOURCE PATH": RESOURCE_PATH,
	"FROM":          FROM,
	"HELP":          HELP,
	"LABELS":        LABELS,
	"METRIC":        METRIC,
	"NAME":          NAME,
	"SET":           SET,
	"TYPE":          TYPE,
	"TO":            TO,
	"DESCRIPTION":   DESCRIPTION,
	"WITH OPTIONS":  WITH_OPTIONS,
	"AS":            AS,
	"BIE":           BIE,
	"BLK":           BLK,
	"EOB":           EOB,
	"EOL":           EOL,
	"CTX":           CTX,
	"EOF":           EOF,
	"COMMA":         COMMA,
	"DEFINE AUTH":   DEFINE_AUTH,
	"EXTRACT USING": EXTRACT_USING,
	"FOR SERVICE":   FOR_SERVICE,
	"FOR STACK":     FOR_STACK,
	"DATASET":       DATASET,
}

type CustomLexer struct {
	l  *Lexer
	ls *lexerState
}

func NewCustomLexer(lex *Lexer, ls *lexerState) *CustomLexer {
	cl := &CustomLexer{l: lex, ls: ls}
	ls.handler = cl
	return cl
}

func (cl CustomLexer) EmitInt(tokenid string, value int) int {
	fmt.Println("PARSER EmitInt", tokenid, value)
	var mVal int
	switch tokenid {
	case "BLK", "EOL", "EOB":
		var found bool
		mVal, found = tks2id[tokenid]
		if !found {
			log.Panicln("invalid value", value)
		}
	default:
		fmt.Println("default:")
		mVal = value
	}
	return mVal
}

func (cl CustomLexer) EmitStr(tokenid string, value string) int {
	fmt.Println("PARSER EmitStr", tokenid, value)
	var mVal int
	if tokenid == "KEY" {
		var found bool
		mVal, found = tks2id[value]
		if !found {
			log.Panicln("invalid value", value)
		}
	} else if tokenid == "STR" {
		return STR_LITERAL
	} else if tokenid == "VAR" {
		return IDENTIFIER
	} else {
		log.Panicln("aaaaaaaaaaaaaaaaaaa", tokenid, value)
	}
	return mVal
}

func (cl CustomLexer) EmitObj(tokenid string, value interface{}) int {
	fmt.Println("PARSER EmitObj", tokenid, value)
	return 0
}

func (l CustomLexer) Lex(lval *yySymType) int {
	fmt.Println("PARSER Lex")
	return l.l.Lex(l.ls)
}

func (l CustomLexer) Error(e string) {
	fmt.Println("PARSER Error", e)
	l.l.Error(e)
}
