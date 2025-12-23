package util

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

type FixTransform struct {
	From string `csv:"from"`
	To   string `csv:"to"`
}

func LoadFixes(filepath string) (map[string]string, error) {
	if filepath == "" {
		return nil, nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading csv file: %v", err)
	}
	defer file.Close()

	var fixes []FixTransform
	if err := gocsv.UnmarshalWithoutHeaders(file, &fixes); err != nil {
		return nil, fmt.Errorf("error unmarshalling csv: %v", err)
	}

	m := make(map[string]string, len(fixes))
	for _, fix := range fixes {
		m[fix.From] = fix.To
	}
	return m, nil
}
