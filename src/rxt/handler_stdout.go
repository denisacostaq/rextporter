package rxt

import "fmt"

// TokenWriter outputs tokens to stdout
type TokenWriter struct {
}

// EmitInt ...
func (tw *TokenWriter) EmitInt(tokenid string, value int) {
	println(tokenid, value)
}

// EmitStr ...
func (tw *TokenWriter) EmitStr(tokenid, value string) {
	println(tokenid, value)
}

// EmitObj ...
func (tw *TokenWriter) EmitObj(tokenid string, value interface{}) {
	println(tokenid, fmt.Sprintf("%v", value))
}
