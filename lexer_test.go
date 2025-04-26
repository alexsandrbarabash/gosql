package gosql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexNumeric(t *testing.T) {
	tests := []struct {
		number bool
		value  string
	}{
		{
			number: true,
			value:  "105",
		},
		{
			number: true,
			value:  "105 ",
		},
		{
			number: true,
			value:  "123.",
		},
		{
			number: true,
			value:  "123.145",
		},
		{
			number: true,
			value:  "1e5",
		},
		{
			number: true,
			value:  "1.e21",
		},
		{
			number: true,
			value:  "1.1e2",
		},
		{
			number: true,
			value:  "1.1e-2",
		},
		{
			number: true,
			value:  "1.1e+2",
		},
		{
			number: true,
			value:  "1e-1",
		},
		{
			number: true,
			value:  ".1",
		},
		{
			number: true,
			value:  "4.",
		},
		{
			number: false,
			value:  "e4",
		},
		{
			number: false,
			value:  "1..",
		},
		{
			number: false,
			value:  "1ee4",
		},
		{
			number: false,
			value:  " 1",
		},
	}

	for _, tt := range tests {
		_, _, result := lexNumeric(tt.value, cursor{})

		if tt.number != result {
			t.Errorf("lexNumeric() failed test case %v", tt.value)
		}
	}

}

func TestLexString(t *testing.T) {
	tests := []struct {
		string bool
		value  string
	}{
		{
			string: false,
			value:  "a",
		},
		{
			string: true,
			value:  "'abc'",
		},
		{
			string: true,
			value:  "'a b'",
		},
		{
			string: true,
			value:  "'a' ",
		},
		{
			string: true,
			value:  "'a '' b'",
		},
		// false tests
		{
			string: false,
			value:  "'",
		},
		{
			string: false,
			value:  "",
		},
		{
			string: false,
			value:  " 'foo'",
		},
	}

	for _, tt := range tests {
		_, _, result := lexString(tt.value, cursor{})

		if tt.string != result {
			t.Errorf("lexString() failed test case %v", tt.value)
		}
	}
}

func TestTokenLexSymbol(t *testing.T) {
	tests := []struct {
		symbol bool
		value  string
	}{
		{
			symbol: true,
			value:  "( ",
		},
		{
			symbol: true,
			value:  ")",
		},
	}

	for _, tt := range tests {
		_, _, result := lexSymbol(tt.value, cursor{})

		if tt.symbol != result {
			t.Errorf("lexSymbol() failed test case %v, %v", tt.value, result)
		}
	}
}

func TestTokenLexIdentifier(t *testing.T) {
	tests := []struct {
		Identifier bool
		input      string
		value      string
	}{
		{
			Identifier: true,
			input:      "a",
			value:      "a",
		},
		{
			Identifier: true,
			input:      "abc",
			value:      "abc",
		},
		{
			Identifier: true,
			input:      "abc ",
			value:      "abc",
		},
		{
			Identifier: true,
			input:      `" abc "`,
			value:      ` abc `,
		},
		{
			Identifier: true,
			input:      "a9$",
			value:      "a9$",
		},
		{
			Identifier: true,
			input:      "userName",
			value:      "username",
		},
		{
			Identifier: true,
			input:      `"userName"`,
			value:      "userName",
		},
		{
			Identifier: false,
			input:      `"`,
		},
		{
			Identifier: false,
			input:      "_sadsfa",
		},
		{
			Identifier: false,
			input:      "9sadsfa",
		},
		{
			Identifier: false,
			input:      " abc",
		},
	}

	for _, tt := range tests {
		token, _, result := lexIdentifier(tt.input, cursor{})

		if tt.Identifier != result {
			t.Errorf("lexIdentifier() failed test case %v", tt.input)
		}

		if token != nil && token.value != tt.value {
			t.Errorf("lexIdentifier() failed test case %v", tt.input)
		}
	}
}

func TestLex(t *testing.T) {
	tests := []struct {
		input  string
		Tokens []token
		err    error
	}{

		{
			input: "select a",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "a",
					kind:  identifierKind,
				},
			},
		},

		// {
		// 	input: "select true",
		// 	Tokens: []token{
		// 		{
		// 			loc:   location{col: 0, line: 0},
		// 			value: string(selectKeyword),
		// 			kind:  keywordKind,
		// 		},
		// 		{
		// 			loc:   location{col: 7, line: 0},
		// 			value: "true",
		// 			kind:  BoolKind,
		// 		},
		// 	},
		// },

		{
			input: "select 1",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "1",
					kind:  numericKind,
				},
			},
			err: nil,
		},

		// {
		// 	input: "select 'foo' || 'bar';",
		// 	Tokens: []token{
		// 		{
		// 			loc:   location{col: 0, line: 0},
		// 			value: string(selectKeyword),
		// 			kind:  keywordKind,
		// 		},
		// 		{
		// 			loc:   location{col: 7, line: 0},
		// 			value: "foo",
		// 			kind:  stringKind,
		// 		},
		// 		{
		// 			loc:   location{col: 13, line: 0},
		// 			value: string(ConcatSymbol),
		// 			kind:  symbolKind,
		// 		},
		// 		{
		// 			loc:   location{col: 16, line: 0},
		// 			value: "bar",
		// 			kind:  stringKind,
		// 		},
		// 		{
		// 			loc:   location{col: 21, line: 0},
		// 			value: string(SemicolonSymbol),
		// 			kind:  symbolKind,
		// 		},
		// 	},
		// 	err: nil,
		// },

		{
			input: "CREATE TABLE u (id INT, name TEXT)",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(createKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: string(tableKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 13, line: 0},
					value: "u",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 15, line: 0},
					value: "(",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 16, line: 0},
					value: "id",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 19, line: 0},
					value: "int",
					kind:  keywordKind,
				},
				{
					loc:   location{col: 22, line: 0},
					value: ",",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 24, line: 0},
					value: "name",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 29, line: 0},
					value: "text",
					kind:  keywordKind,
				},
				{
					loc:   location{col: 33, line: 0},
					value: ")",
					kind:  symbolKind,
				},
			},
		},
		{
			input: "insert into users Values (105, 233)",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(insertKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: string(intoKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 12, line: 0},
					value: "users",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 18, line: 0},
					value: string(valuesKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 25, line: 0},
					value: "(",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 26, line: 0},
					value: "105",
					kind:  numericKind,
				},
				{
					loc:   location{col: 30, line: 0},
					value: ",",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 32, line: 0},
					value: "233",
					kind:  numericKind,
				},
				{
					loc:   location{col: 36, line: 0},
					value: ")",
					kind:  symbolKind,
				},
			},
			err: nil,
		},
		{
			input: "SELECT id FROM users;",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "id",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 10, line: 0},
					value: string(fromKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 15, line: 0},
					value: "users",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 20, line: 0},
					value: ";",
					kind:  symbolKind,
				},
			},
			err: nil,
		},
		{
			input: "INT",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "int",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "SELECT",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "select",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "FROM",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "from",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "AS",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "as",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "TABLE",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "table",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "CREATE",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "create",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "INSERT",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "insert",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "INTO",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "into",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "VALUES",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "values",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "TEXT",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "text",
					kind:  keywordKind,
				},
			},
		},
		{
			input: "('300')",
			Tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: "(",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 1, line: 0},
					value: "300",
					kind:  stringKind,
				},
				{
					loc:   location{col: 6, line: 0},
					value: ")",
					kind:  symbolKind,
				},
			},
		},
	}

	for _, test := range tests {
		tokens, err := lex(test.input)
		assert.Equal(t, test.err, err, test.input)
		assert.Equal(t, len(test.Tokens), len(tokens), test.input)

		for i, tok := range tokens {
			assert.Equal(t, &test.Tokens[i], tok, test.input)
		}
	}
}
