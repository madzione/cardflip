package sqlite

import (
	"fmt"
	"strings"
)

// parseCreateTable extracts what a scan needs from a CREATE TABLE statement:
// column names, the INTEGER PRIMARY KEY alias of rowid tables, and the record
// order of WITHOUT ROWID tables (primary key columns first).
func parseCreateTable(sql string) (*table, error) {
	toks := tokenize(sql)
	open := -1
	for i, t := range toks {
		if t == "(" {
			open = i
			break
		}
	}
	if open < 0 {
		return nil, fmt.Errorf("%w: cannot parse %q", ErrCorrupt, sql)
	}
	depth, close := 0, -1
	for i := open; i < len(toks); i++ {
		switch toks[i] {
		case "(":
			depth++
		case ")":
			depth--
			if depth == 0 {
				close = i
			}
		}
		if close >= 0 {
			break
		}
	}
	if close < 0 {
		return nil, fmt.Errorf("%w: unbalanced parentheses in %q", ErrCorrupt, sql)
	}

	t := &table{rowidCol: -1}
	for i := close + 1; i+1 < len(toks); i++ {
		if strings.EqualFold(toks[i], "without") && strings.EqualFold(toks[i+1], "rowid") {
			t.withoutID = true
		}
	}

	var defs [][]string
	var cur []string
	depth = 0
	for _, tok := range toks[open+1 : close] {
		switch {
		case tok == "(":
			depth++
		case tok == ")":
			depth--
		case tok == "," && depth == 0:
			defs = append(defs, cur)
			cur = nil
			continue
		}
		cur = append(cur, tok)
	}
	defs = append(defs, cur)

	var types []string
	var pk []string
	for _, def := range defs {
		if len(def) == 0 {
			continue
		}
		switch strings.ToUpper(def[0]) {
		case "CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN":
			if cols := tablePrimaryKey(def); cols != nil {
				pk = cols
			}
			continue
		}
		name := unquote(def[0])
		typ := ""
		if len(def) > 1 && isWord(def[1]) && !isConstraintKeyword(def[1]) {
			typ = strings.ToUpper(def[1])
		}
		t.columns = append(t.columns, name)
		types = append(types, typ)
		for j := 1; j+1 < len(def); j++ {
			if strings.EqualFold(def[j], "primary") && strings.EqualFold(def[j+1], "key") {
				pk = []string{name}
			}
		}
	}
	if len(t.columns) == 0 {
		return nil, fmt.Errorf("%w: no columns in %q", ErrCorrupt, sql)
	}

	index := func(name string) int {
		for i, c := range t.columns {
			if strings.EqualFold(c, name) {
				return i
			}
		}
		return -1
	}
	if !t.withoutID {
		t.order = identity(len(t.columns))
		if len(pk) == 1 {
			if i := index(pk[0]); i >= 0 && types[i] == "INTEGER" {
				t.rowidCol = i
			}
		}
		return t, nil
	}

	if len(pk) == 0 {
		return nil, fmt.Errorf("%w: WITHOUT ROWID table without primary key", ErrCorrupt)
	}
	inPK := map[int]bool{}
	for _, name := range pk {
		i := index(name)
		if i < 0 {
			return nil, fmt.Errorf("%w: primary key column %q not declared", ErrCorrupt, name)
		}
		if !inPK[i] {
			inPK[i] = true
			t.order = append(t.order, i)
		}
	}
	for i := range t.columns {
		if !inPK[i] {
			t.order = append(t.order, i)
		}
	}
	return t, nil
}

// tablePrimaryKey returns the columns of a "[CONSTRAINT x] PRIMARY KEY (a, b)"
// table constraint, or nil for other constraints.
func tablePrimaryKey(def []string) []string {
	for i := 0; i+2 < len(def); i++ {
		if strings.EqualFold(def[i], "primary") && strings.EqualFold(def[i+1], "key") && def[i+2] == "(" {
			var cols []string
			for _, tok := range def[i+3:] {
				if tok == ")" {
					break
				}
				if isWord(tok) && !isOrderKeyword(tok) {
					cols = append(cols, unquote(tok))
				}
			}
			return cols
		}
	}
	return nil
}

func isConstraintKeyword(tok string) bool {
	switch strings.ToUpper(tok) {
	case "PRIMARY", "NOT", "NULL", "UNIQUE", "CHECK", "DEFAULT", "COLLATE", "REFERENCES", "CONSTRAINT", "GENERATED", "AS":
		return true
	}
	return false
}

func isOrderKeyword(tok string) bool {
	switch strings.ToUpper(tok) {
	case "ASC", "DESC", "COLLATE":
		return true
	}
	return false
}

func isWord(tok string) bool {
	return tok != "" && tok != "(" && tok != ")" && tok != ","
}

func unquote(tok string) string {
	if len(tok) >= 2 {
		switch {
		case tok[0] == '"' && tok[len(tok)-1] == '"',
			tok[0] == '`' && tok[len(tok)-1] == '`',
			tok[0] == '\'' && tok[len(tok)-1] == '\'',
			tok[0] == '[' && tok[len(tok)-1] == ']':
			return tok[1 : len(tok)-1]
		}
	}
	return tok
}

// tokenize splits SQL into words, quoted identifiers/strings and the
// punctuation "(", ")" and ",". Comments are dropped.
func tokenize(sql string) []string {
	var toks []string
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '-' && i+1 < len(sql) && sql[i+1] == '-':
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(sql) && sql[i+1] == '*':
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 {
				i = len(sql)
			} else {
				i += end + 4
			}
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '(' || c == ')' || c == ',':
			toks = append(toks, string(c))
			i++
		case c == '"' || c == '`' || c == '\'' || c == '[':
			end := byte(c)
			if c == '[' {
				end = ']'
			}
			j := i + 1
			for j < len(sql) && sql[j] != end {
				j++
			}
			if j < len(sql) {
				j++
			}
			toks = append(toks, sql[i:j])
			i = j
		default:
			j := i
			for j < len(sql) && !strings.ContainsRune(" \t\n\r(),\"`'[", rune(sql[j])) {
				j++
			}
			if j == i {
				j++
			}
			toks = append(toks, sql[i:j])
			i = j
		}
	}
	return toks
}
