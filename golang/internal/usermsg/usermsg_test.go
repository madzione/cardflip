package usermsg

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

func TestText(t *testing.T) {
	s := Text("deckNameTaken", "A deck named %q already exists.", "Spanish")
	if want := "\x02{\"code\":\"deckNameTaken\",\"args\":[\"Spanish\"]}\x03A deck named \"Spanish\" already exists.\x04"; s != want {
		t.Fatalf("Text = %q, want %q", s, want)
	}
	if Code(s) != "deckNameTaken" || !slices.Equal(Arguments(s), []string{"Spanish"}) {
		t.Fatalf("Code, Arguments = %q, %q", Code(s), Arguments(s))
	}
	if p := Plain(s); p != `A deck named "Spanish" already exists.` {
		t.Fatalf("Plain = %q", p)
	}
}

// A wrapping error keeps its context around the tagged text.
func TestWrapped(t *testing.T) {
	inner := errors.New(Text("noCards", "No cards."))
	err := fmt.Errorf("import: %w", inner)
	if p := Plain(err.Error()); p != "import: No cards." {
		t.Fatalf("Plain = %q", p)
	}
	if Code(err.Error()) != "noCards" {
		t.Fatalf("Code = %q", Code(err.Error()))
	}
}

// A tagged argument stays tagged in the arguments, so the app translates
// it, and reads as English in the English wording.
func TestNestedArgument(t *testing.T) {
	inner := errors.New(Text("fieldEmpty", "The field %d is empty.", 2))
	s := Text("noteSkipped", "Skipped %q: %v", "hola", inner)
	if args := Arguments(s); args[1] != inner.Error() {
		t.Fatalf("argument = %q, want the tagged error", args[1])
	}
	if p := Plain(s); p != `Skipped "hola": The field 2 is empty.` {
		t.Fatalf("Plain = %q", p)
	}
}

func TestUntagged(t *testing.T) {
	if Plain("plain") != "plain" || Code("plain") != "" || Arguments("plain") != nil {
		t.Fatal("an untagged text changed")
	}
	if Text("x", "%d%%", 5) != "\x02{\"code\":\"x\",\"args\":[\"5\"]}\x035%\x04" {
		t.Fatalf("Text = %q", Text("x", "%d%%", 5))
	}
}
