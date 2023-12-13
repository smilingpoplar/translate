package google

import (
	"testing"

	"github.com/smilingpoplar/translate/translator/decorator"
)

func TestTranslate(t *testing.T) {
	t.Parallel()
	texts := []string{"hello", "world"}
	expect := []string{"你好", "世界"}
	g := New()
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
	g := New()
	if tl, ok := g.inner.(*decorator.TextsLimit); ok {
		tl.MaxLen = 6
	} else {
		t.Fatal("expect TextLimitDecorator")
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
