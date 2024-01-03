package google

import (
	"testing"

	"github.com/smilingpoplar/translate/translator/middleware"
)

func TestTranslate(t *testing.T) {
	t.Parallel()
	texts := []string{"hello", "world"}
	expect := []string{"你好", "世界"}
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	got, err := g.Translate(texts, "zh-CN")
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

func TestTranslateWithTextsLimit(t *testing.T) {
	t.Parallel()
	texts := []string{"hello\nworld", "hello"}
	expect := []string{"你好\n世界", "你好"}
	g, err := New()
	if err != nil {
		t.Fatal(err)
	}
	mw := middleware.TextsLimit(6)
	g.handler = mw(g.translate)

	got, err := g.Translate(texts, "zh-CN")
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
