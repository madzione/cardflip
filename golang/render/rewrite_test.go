package render

import (
	"fmt"
	"testing"
)

func TestRenameField(t *testing.T) {
	in := "{{Front}} {{text:Front}} {{#Front}}x{{Front}}{{/Front}} {{Back}} {{FrontSide}}"
	want := "{{Word}} {{text:Word}} {{#Word}}x{{Word}}{{/Word}} {{Back}} {{FrontSide}}"
	if got := RenameField(in, "Front", "Word"); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestRenameFieldsAtOnce(t *testing.T) {
	// Swapping names works because every reference is renamed once.
	in := "{{Front}} {{#Back}}{{type:Back}}{{/Back}} {{FrontSide}}"
	want := "{{Back}} {{#Front}}{{type:Front}}{{/Front}} {{FrontSide}}"
	if got := RenameFields(in, map[string]string{"Front": "Back", "Back": "Front"}); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveField(t *testing.T) {
	in := "{{FrontSide}}<hr id=answer>{{Back}}{{#Back}} has back{{/Back}}{{^Back}} no back{{/Back}}{{hint:Back}}"
	want := "{{FrontSide}}<hr id=answer> no back"
	if got := RemoveField(in, "Back"); got != want {
		t.Fatalf("got %q", got)
	}
}

func TestFieldsUsed(t *testing.T) {
	got := FieldsUsed("{{#Hint}}{{Hint}}{{/Hint}} {{cloze:Text}} {{Text}}")
	if fmt.Sprint(got) != "[Hint Text]" {
		t.Fatalf("got %v", got)
	}
}
