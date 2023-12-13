package decorator

import (
	"fmt"

	"github.com/smilingpoplar/translate/translator"
)

type TextsRegroup struct {
	inner   translator.Translator
	MaxLen  int
	onTrans func([]string) error
}

func TextsRegroupDecorator(inner translator.Translator, maxLen int) *TextsRegroup {
	return &TextsRegroup{
		inner:  inner,
		MaxLen: maxLen,
	}
}

func (d *TextsRegroup) Translate(texts []string, toLang string) ([]string, error) {
	groups, err := regroupTexts(texts, d.MaxLen)
	if err != nil {
		return nil, fmt.Errorf("error group texts: %w", err)
	}

	var result []string
	for _, group := range groups {
		part, err := d.inner.Translate(group, toLang)
		if err != nil {
			return nil, err
		}
		result = append(result, part...)
		if d.onTrans != nil {
			if err := d.onTrans(part); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (d *TextsRegroup) OnTranslated(f func([]string) error) {
	d.onTrans = f
}
