package rxt

import "fmt"

// TokenWriter outputs tokens to stdout
type TokenWriter struct {
}

// EmitInt ...
func (tw *TokenWriter) EmitInt(tokenid string, value int) int {
	println(tokenid, value)
	return 0
}

// EmitStr ...
func (tw *TokenWriter) EmitStr(tokenid, value string) int {
	println(tokenid, value)
	return 0
}

// EmitObj ...
func (tw *TokenWriter) EmitObj(tokenid string, value interface{}) int {
	println(tokenid, fmt.Sprintf("%v", value))
	return 0
}

// Error ...
func (tw *TokenWriter) Error(e string) {
	println("UNK", e)
}
