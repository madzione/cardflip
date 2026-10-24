package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"cardflip/golang/internal/usermsg"
)

// codeArg is the position of the code among the arguments of the functions
// that tag user texts (see package usermsg).
var codeArg = map[string]int{
	"userErr": 0, "NewError": 0, "Text": 0, "Tag": 0, "unreadable": 0, "newError": 2, "bad": 0,
}

// dynamicCodes are the codes made at run time, by the expression that makes
// them.
var dynamicCodes = map[string][]string{
	`code+"Empty"`:         {"fieldNameEmpty", "cardTypeNameEmpty"},
	`code+"Invalid"`:       {"fieldNameInvalid", "cardTypeNameInvalid"},
	`steps.code+"TooMany"`: {"learningStepsTooMany", "relearningStepsTooMany"},
	`steps.code+"Range"`:   {"learningStepsRange", "relearningStepsRange"},
	`side.code`:            {"secondsQuestionRange", "secondsAnswerRange"},
	`code+"UnknownField"`:  {"templateFrontUnknownField", "templateBackUnknownField"},
	`code+"Mistake"`:       {"templateFrontMistake", "templateBackMistake"},
	`roleCode(m)`:          {"csvOneNoteTypeColumn", "csvOneDeckColumn", "csvOneGuidColumn", "csvOneTagsColumn"},
	`code`:                 nil, // the parameter of userErr, NewError, Text, unreadable and newError
	`e.Code`:               nil, // search.Error, made by newError
}

// Every code Go tags a text with has an English message in the app's
// translations, as "go" + the code, and every Go message there is used
// (app-shell spec: translated text).
func TestCodesHaveMessages(t *testing.T) {
	used := map[string]string{} // code: where
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (path == "fluffy" || d.Name() == "usermsg") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			var name string
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
				if x, ok := fun.X.(*ast.Ident); !ok || x.Name != "usermsg" && x.Name != "core" {
					return true
				}
			}
			i, ok := codeArg[name]
			if !ok || len(call.Args) <= i {
				return true
			}
			pos := fset.Position(call.Pos()).String()
			if lit, ok := call.Args[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				code, _ := strconv.Unquote(lit.Value)
				used[code] = pos
				return true
			}
			src := exprString(fset, path, call.Args[i])
			codes, ok := dynamicCodes[src]
			if !ok {
				t.Errorf("%s: code %s is neither a string nor a known expression; add it to dynamicCodes", pos, src)
			}
			for _, c := range codes {
				used[c] = pos
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(used) < 200 {
		t.Fatalf("found only %d codes; is the walk broken?", len(used))
	}

	data, err := os.ReadFile("../lib/l10n/app_en.arb")
	if err != nil {
		t.Fatal(err)
	}
	var arb map[string]any
	if err := json.Unmarshal(data, &arb); err != nil {
		t.Fatal(err)
	}
	for code, pos := range used {
		if _, ok := arb["go"+strings.ToUpper(code[:1])+code[1:]]; !ok {
			t.Errorf("%s: code %q has no message go%s%s in app_en.arb", pos, code, strings.ToUpper(code[:1]), code[1:])
		}
	}
	var unused []string
	for key := range arb {
		if rest, ok := strings.CutPrefix(key, "go"); ok && rest != "" && rest[:1] == strings.ToUpper(rest[:1]) {
			if _, ok := used[strings.ToLower(rest[:1])+rest[1:]]; !ok {
				unused = append(unused, key)
			}
		}
	}
	slices.Sort(unused)
	if len(unused) > 0 {
		t.Errorf("app_en.arb has messages for codes Go does not use: %v", unused)
	}
}

func exprString(fset *token.FileSet, path string, e ast.Expr) string {
	src, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(src[fset.Position(e.Pos()).Offset:fset.Position(e.End()).Offset])
}

// Errors reach the app tagged with a code and arguments, next to the English
// message (app-shell spec: error in German).
func TestErrorsCarryCodes(t *testing.T) {
	s, _ := opened(t)
	must(s.decks.Create("Languages::Spanish"))
	french := must(s.decks.Create("Languages::French"))
	err := s.decks.Rename(french, "Languages::Spanish")
	if err == nil {
		t.Fatal("rename to a sibling's name accepted")
	}
	msg := err.Error()
	if code, args := usermsg.Code(msg), usermsg.Arguments(msg); code != "deckNameTakenByOther" || !slices.Equal(args, []string{"Spanish"}) {
		t.Fatalf("code, args = %q, %q", code, args)
	}
	if p := usermsg.Plain(msg); p != `The name "Spanish" is already taken by another deck.` {
		t.Fatalf("English = %q", p)
	}

	// Service and search errors are tagged too.
	if err := s.app.SetFlagName(9, "x"); err == nil || usermsg.Code(err.Error()) != "unknownFlag" {
		t.Fatalf("flag error = %v", err)
	}
	_, err = s.browse.Search("(deck:x", "created", false)
	if err == nil || usermsg.Code(err.Error()) != "searchErrorAt" {
		t.Fatalf("search error = %v", err)
	}
	if inner := usermsg.Arguments(err.Error())[0]; usermsg.Code(inner) != "searchParenNotClosed" {
		t.Fatalf("search error message = %q", inner)
	}
}
