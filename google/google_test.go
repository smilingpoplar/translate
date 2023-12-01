package google_test

import (
	"testing"

	"github.com/smilingpoplar/translate/google"
)

func TestTranslate(t *testing.T) {
	t.Parallel()
	texts := []string{"hello", "world"}
	expect := []string{"你好", "世界"}
	g := google.New()
	got, err := g.Translate(texts)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(expect) {
		t.Fatalf("expect len %d, got len %d", len(expect), len(got))
	}
	for i := range got {
		if got[i] != expect[i] {
			t.Errorf("expect %s, got %s", expect, got)
		}
	}
}
