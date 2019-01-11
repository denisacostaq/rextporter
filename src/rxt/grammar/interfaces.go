package grammar

// TokenHandler emits tokens discovered by the RXT lexer
type TokenHandler interface {
	EmitInt(tokenid string, value int) int
	EmitStr(tokenid, value string) int
	EmitObj(tokenid string, value interface{}) int
	Error(e string)
}
