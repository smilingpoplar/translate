package decorator

import (
	"fmt"
	"strings"

	"github.com/smilingpoplar/translate/translator"
)

type TextLimit struct {
	inner  translator.Translator
	maxLen int
}

func TextLimitDecorator(inner translator.Translator, maxLen int) *TextLimit {
	return &TextLimit{
		inner:  TextsRegroupDecorator(inner, maxLen),
		maxLen: maxLen,
	}
}

// 将所有texts合并成一段，发往只接收text的翻译接口
func (d *TextLimit) Translate(texts []string, toLang string) ([]string, error) {
	texts, info, err := splitLongTexts([]string{strings.Join(texts, "\n")}, d.maxLen-len(texts)+1)
	if err != nil || len(info.Mapping) > 1 {
		return nil, fmt.Errorf("error split long text: %w", err)
	}
	if len(info.Mapping) == 1 {
		texts = texts[1:]
	}
	return d.inner.Translate(texts, toLang)
}

func (d *TextLimit) OnTranslated(f func([]string) error) {
	d.inner.(*TextsRegroup).OnTranslated(f)
}
