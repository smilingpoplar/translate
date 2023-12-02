package util

import (
	"testing"
)

func TestTextTooLong(t *testing.T) {
	t.Parallel()

	list := []string{"hello", "wonderful", "world"}
	_, err := RegroupTexts(list, 5)
	if err == nil {
		t.Fatal("expect error: text too long")
	}
}

func TestRegroupTexts(t *testing.T) {
	t.Parallel()

	list := []string{"hello", "wonderful", "world"}
	want := [][]string{{"hello", "wonderful"}, {"world"}}
	got, err := RegroupTexts(list, 15)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		if len(got[i]) == 0 {
			t.Fatalf("got empty list %d", i)
		}
		for j := range got[i] {
			if got[i][j] != want[i][j] {
				t.Errorf("expect %s, got %s", want[i][j], got[i][j])
			}
		}
	}
}
