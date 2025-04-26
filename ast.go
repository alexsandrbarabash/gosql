package gosql

type AstKind uint

type expressionKind uint

const (
    literalKind expressionKind = iota
)

type expression struct {
    literal *token
    kind    expressionKind
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

type SelectStatement struct {
    item []*expression
    from token
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