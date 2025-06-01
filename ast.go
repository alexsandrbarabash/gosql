package gosql

type AstKind uint

type expressionKind uint

const (
	literalKind expressionKind = iota
	binaryKind
)

type expression struct {
	literal *token
	binary  *binaryExpression
	kind    expressionKind
}

type binaryExpression struct {
	a  expression
	b  expression
	op token
}

const (
	SelectKind AstKind = iota
	CreateTableKind
	InsertKind
)

type InsertStatement struct {
	table  token
	values *[]*expression
}

type columnDefinition struct {
	name     token
	datatype token
}

type CreateTableStatement struct {
	name token
	cols *[]*columnDefinition
}

type selectItem struct {
	exp      *expression
	asteriks bool
	as       *token
}

type fromItem struct {
	table *token
}

type SelectStatement struct {
	item  *[]*selectItem
	from  *fromItem
	where *expression
}

type Statement struct {
	SelectStatement      *SelectStatement
	CreateTableStatement *CreateTableStatement
	InsertStatement      *InsertStatement
	Kind                 AstKind
}

type Ast struct {
	Statements []*Statement
}
