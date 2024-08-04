package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
)

type FixTransform struct {
	From string `csv:"from"`
	To   string `csv:"to"`
}

func LoadTranslationFixes(filepath string) ([]FixTransform, error) {
	if filepath == "" {
		return nil, nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file: %v", err)
	}
	defer file.Close()

	list := []FixTransform{}
	if err := gocsv.UnmarshalWithoutHeaders(file, &list); err != nil {
		return nil, fmt.Errorf("error unmarshalling CSV: %v", err)
	}

	return list, nil
}

func ApplyTranslationFixes(texts []string, fixes []FixTransform) {
	for _, fix := range fixes {
		for i := range texts {
			texts[i] = strings.ReplaceAll(texts[i], fix.From, fix.To)
		}
	}
}
