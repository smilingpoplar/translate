package google_test

import (
	"testing"

	"github.com/smilingpoplar/translate/google"
)

func TestTranslate(t *testing.T) {
	t.Parallel()
	text := "hello world"
	expect := "你好世界"
	g := google.New()
	got, err := g.Translate(text)
	if err != nil {
		t.Fatal(err)
	}
	if expect != got {
		t.Errorf("expect %s, got %s", expect, got)
	}
}
