package gosql

import (
	"errors"
	"fmt"
)

func tokenFromKeyword(k keyword) token {
	return token{
		kind:  keywordKind,
		value: string(k),
	}
}

func tokenFromSymbol(s symbol) token {
	return token{
		kind:  symbolKind,
		value: string(s),
	}
}

func helpMessage(tokens []*token, cursor uint, msg string) {
	var c *token
	if cursor+1 < uint(len(tokens)) {
		c = tokens[cursor]
	} else {
		c = tokens[cursor]
	}

	fmt.Printf("[%d,%d]: %s, got: %s\n", c.loc.line, c.loc.col, msg, c.value)
}

func parseToken(tokens []*token, initialCursor uint, t token) (*token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}

	p := tokens[cursor];

	if t.equals(p) {
		return p, cursor + 1, true
	}

	return nil, initialCursor, false

}

func parseTokenKind(tokens []*token, initialCursor uint, kind tokenKind) (*token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}

	current := tokens[cursor]
	if current.kind == kind {
		return current, cursor + 1, true
	}

	return nil, initialCursor, false
}

func parseLiteralExpression(tokens []*token, initialCursor uint) (*expression, uint, bool) {
	cursor := initialCursor

	kinds := []tokenKind{identifierKind, numericKind, stringKind, boolKind}
	for _, kind := range kinds {
		t, newCursor, ok := parseTokenKind(tokens, cursor, kind)
		if ok {
			return &expression{
				literal: t,
				kind:    literalKind,
			}, newCursor, true
		}
	}

	return nil, initialCursor, false
}

func parseExpression(tokens []*token, initialCursor uint, delimiters []token) (*expression, uint, bool) {
	cursor := initialCursor

	var exp *expression
	_, newCursor, ok := parseToken(tokens, cursor, tokenFromSymbol(leftParenSymbol))

	if ok {
		cursor = newCursor
		rightParenToken := tokenFromSymbol(rightParenSymbol)

		exp, cursor, ok = parseExpression(tokens, cursor, append(delimiters, rightParenToken))
		if !ok {
			helpMessage(tokens, cursor, "Expected expression after opening paren")
			return nil, initialCursor, false
		}

		_, cursor, ok = parseToken(tokens, cursor, rightParenToken)

		if !ok {
			helpMessage(tokens, cursor, "Expected closing paren")
			return nil, initialCursor, false
		}

	} else {
		exp, cursor, ok = parseLiteralExpression(tokens, cursor)
		if !ok {
			return nil, initialCursor, false
		}
	}

	for _, d := range delimiters {
		_, _, ok = parseToken(tokens, cursor, d)
		if ok {
			return exp, cursor, true
		}
	}

	// Pratt parser work

	binOps := []token{
		tokenFromKeyword(andKeyword),
		tokenFromKeyword(orKeyword),
		tokenFromSymbol(eqSymbol),
		tokenFromSymbol(neqSymbol),
		tokenFromSymbol(neqSymbol2),
		tokenFromSymbol(gtSymbol),
		tokenFromSymbol(gteSymbol),
		tokenFromSymbol(ltSymbol),
		tokenFromSymbol(lteSymbol),
		tokenFromSymbol(concatSymbol),
		tokenFromSymbol(plusSymbol),
		tokenFromSymbol(minusSymbol),
	}

	binExp := binaryExpression{
		a: *exp,
	}

	binOpFound := false

	for _, op := range binOps {
		var t *token
		t, cursor, ok = parseToken(tokens, cursor, op)
		if ok {
			binExp.op = *t
			binOpFound = true
			break
		}
	}

	if !binOpFound {
		helpMessage(tokens, cursor, "Expected binary operator")
		return nil, initialCursor, false
	}

	b, newCursor, ok := parseExpression(tokens, cursor, delimiters)

	if !ok {
		helpMessage(tokens, cursor, "Expected right operand")
		return nil, initialCursor, false
	}

	binExp.b = *b

	return &expression{
		binary: &binExp,
		kind:   binaryKind,
	}, newCursor, true
}

func parseExpressions(tokens []*token, initialCursor uint, delimiters []token) (*[]*expression, uint, bool) {
	cursor := initialCursor

	exps := []*expression{}
outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for delimiter
		current := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		// Look for comma
		if len(exps) > 0 {
			var ok bool
			_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(commaSymbol))
			if !ok {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}
		}

		// Look for expression
		exp, newCursor, ok := parseExpression(tokens, cursor, []token{tokenFromSymbol(commaSymbol), tokenFromSymbol(rightParenSymbol)})
		if !ok {
			helpMessage(tokens, cursor, "Expected expression")
			return nil, initialCursor, false
		}
		cursor = newCursor

		exps = append(exps, exp)
	}

	return &exps, cursor, true
}

func parseSelectItem(tokens []*token, initialCursor uint, delimiters []token) (*[]*selectItem, uint, bool) {
	cursor := initialCursor

	s := []*selectItem{}

outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		current := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		var ok bool
		if len(s) > 0 {
			_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(commaSymbol))
			if !ok {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}
		}

		var si selectItem

		_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(asteriskSymbol))

		if ok {
			si = selectItem{asteriks: true}
		} else {
			asToken := tokenFromKeyword(asKeyword)
			delimiters := append(delimiters, tokenFromSymbol(commaSymbol), asToken)
			exp, newCursor, ok := parseExpression(tokens, cursor, delimiters)

			if !ok {
				helpMessage(tokens, cursor, "Expected expression")
				return nil, initialCursor, false
			}

			cursor = newCursor
			si.exp = exp

			_, cursor, ok = parseToken(tokens, cursor, asToken)

			if ok {
				id, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
				if !ok {
					helpMessage(tokens, cursor, "Expected identifier after AS")
				}

				cursor = newCursor
				si.as = id
			}
		}

		s = append(s, &si)

	}

	return &s, cursor, true
}

func parseFromItem(tokens []*token, initialCursor uint, _ []token) (*fromItem, uint, bool) {
	ident, newCursor, ok := parseTokenKind(tokens, initialCursor, identifierKind)
	if !ok {
		return nil, initialCursor, false
	}

	return &fromItem{table: ident}, newCursor, true
}

func parseSelectStatement(tokens []*token, initialCursor uint, delimiter token) (*SelectStatement, uint, bool) {
	var ok bool
	cursor := initialCursor

	_, cursor, ok = parseToken(tokens, cursor, tokenFromKeyword(selectKeyword))
	if !ok {
		return nil, initialCursor, false
	}

	slct := SelectStatement{}

	fromToken := tokenFromKeyword(fromKeyword)
	item, newCursor, ok := parseSelectItem(tokens, cursor, []token{fromToken, delimiter})
	if !ok {
		return nil, initialCursor, false
	}

	slct.item = item
	cursor = newCursor

	whereToken := tokenFromKeyword(whereKeyword)
	delimiters := []token{delimiter, whereToken}

	_, cursor, ok = parseToken(tokens, cursor, fromToken)
	if ok {

		from, newCursor, ok := parseFromItem(tokens, cursor, delimiters)
		if !ok {
			helpMessage(tokens, cursor, "Expected FROM token")
			return nil, initialCursor, false
		}

		slct.from = from
		cursor = newCursor
	}

	_, cursor, ok = parseToken(tokens, cursor, whereToken)
	if ok {
		where, newCursor, ok := parseExpression(tokens, cursor, []token{delimiter})

		if !ok {
			helpMessage(tokens, cursor, "Expected WHERE conditionals")
		}
		slct.where = where
		cursor = newCursor
	}

	return &slct, cursor, true
}

func parseInsertStatement(tokens []*token, initialCursor uint, delimiter token) (*InsertStatement, uint, bool) {
	cursor := initialCursor

	// Look for INSERT
	_, cursor, ok := parseToken(tokens, cursor, tokenFromKeyword(insertKeyword))
	if !ok {
		return nil, initialCursor, false
	}

	// Look for INTO
	_, cursor, ok = parseToken(tokens, cursor, tokenFromKeyword(intoKeyword))
	if !ok {
		helpMessage(tokens, cursor, "Expected into")
		return nil, initialCursor, false
	}

	// Look for table name
	table, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for VALUES
	_, cursor, ok = parseToken(tokens, cursor, tokenFromKeyword(valuesKeyword))
	if !ok {
		helpMessage(tokens, cursor, "Expected VALUES")
		return nil, initialCursor, false
	}

	// Look for left paren
	_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(leftParenSymbol))
	if !ok {
		helpMessage(tokens, cursor, "Expected left paren")
		return nil, initialCursor, false
	}

	// Look for expression list
	values, newCursor, ok := parseExpressions(tokens, cursor, []token{tokenFromSymbol(rightParenSymbol)})
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for right paren
	_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(rightParenSymbol))
	if !ok {
		helpMessage(tokens, cursor, "Expected right paren")
		return nil, initialCursor, false
	}

	return &InsertStatement{
		table:  *table,
		values: values,
	}, cursor, true
}

func parseStatement(tokens []*token, initialCursor uint, delimiter token) (*Statement, uint, bool) {
	cursor := initialCursor

	// Look for a SELECT statement
	semicolonToken := tokenFromSymbol(semicolonSymbol)
	slct, newCursor, ok := parseSelectStatement(tokens, cursor, semicolonToken)

	if ok {
		return &Statement{
			Kind:            SelectKind,
			SelectStatement: slct,
		}, newCursor, true
	}

	// // Look for a INSERT statement
	inst, newCursor, ok := parseInsertStatement(tokens, cursor, semicolonToken)
	if ok {
		return &Statement{
			Kind:            InsertKind,
			InsertStatement: inst,
		}, newCursor, true
	}

	// Look for a CREATE statement
	crtTbl, newCursor, ok := parseCreateTableStatement(tokens, cursor, semicolonToken)
	if ok {
		return &Statement{
			Kind:                 CreateTableKind,
			CreateTableStatement: crtTbl,
		}, newCursor, true
	}

	return nil, initialCursor, false
}

func parseColumnDefinitions(tokens []*token, initialCursor uint, delimiter token) (*[]*columnDefinition, uint, bool) {
	cursor := initialCursor

	cds := []*columnDefinition{}
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for a delimiter
		current := tokens[cursor]
		if delimiter.equals(current) {
			break
		}

		// Look for a comma
		if len(cds) > 0 {
			var ok bool
			_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(commaSymbol))
			if !ok {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}
		}

		// Look for a column name
		id, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected column name")
			return nil, initialCursor, false
		}
		cursor = newCursor

		// Look for a column type
		ty, newCursor, ok := parseTokenKind(tokens, cursor, keywordKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected column type")
			return nil, initialCursor, false
		}
		cursor = newCursor

		cds = append(cds, &columnDefinition{
			name:     *id,
			datatype: *ty,
		})
	}

	return &cds, cursor, true
}

func parseCreateTableStatement(tokens []*token, initialCursor uint, delimiter token) (*CreateTableStatement, uint, bool) {
	cursor := initialCursor

	_, cursor, ok := parseToken(tokens, cursor, tokenFromKeyword(createKeyword))
	if !ok {
		return nil, initialCursor, false
	}

	_, cursor, ok = parseToken(tokens, cursor, tokenFromKeyword(tableKeyword))
	if !ok {
		return nil, initialCursor, false
	}

	name, newCursor, ok := parseTokenKind(tokens, cursor, identifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(leftParenSymbol))
	if !ok {
		helpMessage(tokens, cursor, "Expected left parenthesis")
		return nil, initialCursor, false
	}

	cols, newCursor, ok := parseColumnDefinitions(tokens, cursor, tokenFromSymbol(rightParenSymbol))
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(rightParenSymbol))
	if !ok {
		helpMessage(tokens, cursor, "Expected right parenthesis")
		return nil, initialCursor, false
	}

	return &CreateTableStatement{
		name: *name,
		cols: cols,
	}, cursor, true
}

func Parse(source string) (*Ast, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}

	a := Ast{}
	cursor := uint(0)
	for cursor < uint(len(tokens)) {
		stmt, newCursor, ok := parseStatement(tokens, cursor, tokenFromSymbol(semicolonSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expected statement")
			return nil, errors.New("Failed to parse, expected statement")
		}
		cursor = newCursor

		a.Statements = append(a.Statements, stmt)

		atLeastOneSemicolon := false

		for {
			_, cursor, ok = parseToken(tokens, cursor, tokenFromSymbol(semicolonSymbol))
			if ok {
				atLeastOneSemicolon = true
			} else {
				break
			}
		}

		if !atLeastOneSemicolon {
			helpMessage(tokens, cursor, "Expected semi-colon delimiter between statements")
			return nil, errors.New("Missing semi-colon between statements")
		}
	}

	return &a, nil
}
