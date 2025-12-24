package util

import (
	"fmt"
	"os"
	"regexp"

	"github.com/gocarina/gocsv"
)

type GlossaryEntry struct {
	From string `csv:"from"`
	To   string `csv:"to"`
}

func LoadGlossary(filepath string) (map[string]string, error) {
	if filepath == "" {
		return nil, nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading csv file: %v", err)
	}
	defer file.Close()

	var entries []GlossaryEntry
	if err := gocsv.UnmarshalWithoutHeaders(file, &entries); err != nil {
		return nil, fmt.Errorf("error unmarshalling csv: %v", err)
	}

	m := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.From != "" {
			m[entry.From] = entry.To
		}
	}
	return m, nil
}

func GeneratePlaceholder(id int) string {
	return fmt.Sprintf("{ID_%d}", id)
}

func BuildWordBoundaryRegex(word string) (*regexp.Regexp, error) {
	escaped := regexp.QuoteMeta(word)
	pattern := `(?i)\b` + escaped + `\b`
	return regexp.Compile(pattern)
}
