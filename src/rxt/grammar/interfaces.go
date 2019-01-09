package grammar

// TokenHandler emits tokens discovered by the RXT lexer
type TokenHandler interface {
	EmitInt(tokenid string, value int)
	EmitStr(tokenid, value string)
	EmitObj(tokenid string, value interface{})
}
