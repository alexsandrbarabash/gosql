package gosql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTable(t *testing.T) {
	mb := NewMemoryBackend()

	ast, err := Parse("CREATE TABLE test(x INT, y INT, z TEXT);")
	assert.Nil(t, err)

	err = mb.CreateTable(ast.Statements[0].CreateTableStatement)
	assert.Nil(t, err)
	assert.Equal(t, mb.tables["test"].name, "test")

	assert.Equal(t, mb.tables["test"].columns, []string{"x", "y", "z"})
	assert.Equal(t, mb.tables["test"].columnTypes, []ColumnType{IntType, IntType, TextType})
}

func TestInsert(t *testing.T) {
	mb := NewMemoryBackend()

	ast, err := Parse("INSERT INTO test VALUES(100, 200, 300);")
	assert.Nil(t, err)
	assert.NotEqual(t, ast, nil)
	err = mb.Insert(ast.Statements[0].InsertStatement)
	assert.Equal(t, err, ErrTableDoesNotExist)

	ast, err = Parse("CREATE TABLE test(x INT, y INT, z TEXT);")
	assert.Nil(t, err)
	assert.NotEqual(t, ast, nil)
	err = mb.CreateTable(ast.Statements[0].CreateTableStatement)
	assert.Nil(t, err)

	ast, err = Parse(`INSERT INTO test VALUES(100, 200, '300');`)
	assert.Nil(t, err)
	assert.NotEqual(t, ast, nil)
	err = mb.Insert(ast.Statements[0].InsertStatement)
	assert.Nil(t, err)

	assert.Equal(t, mb.tables["test"].rows[0], []MemoryCell{
		mb.tokenToCell(&token{kind: numericKind, value: "100"}),
		mb.tokenToCell(&token{kind: numericKind, value: "200"}),
		mb.tokenToCell(&token{kind: stringKind, value: "300"}),
	})
}

func TestSelect(t *testing.T) {
	mb := NewMemoryBackend()

	ast, err := Parse("CREATE TABLE test(x INT, y INT, z INT);")
	assert.Nil(t, err)
	err = mb.CreateTable(ast.Statements[0].CreateTableStatement)
	assert.Nil(t, err)

	ast, err = Parse("INSERT INTO test VALUES(100, 200, 300);")
	assert.Nil(t, err)
	assert.NotEqual(t, ast, nil)
	err = mb.Insert(ast.Statements[0].InsertStatement)
	assert.Nil(t, err)

	ast, err = Parse("SELECT x, y FROM test;")
	assert.Nil(t, err)
	assert.NotEqual(t, ast, nil)
	results, err := mb.Select(ast.Statements[0].SelectStatement)
	assert.Nil(t, err)
	assert.Equal(t, results.Rows[0], []Cell{
		mb.tokenToCell(&token{kind: numericKind, value: "100"}),
		mb.tokenToCell(&token{kind: numericKind, value: "200"}),
	})
}
