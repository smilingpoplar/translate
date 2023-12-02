package util

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func ReadLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading lines: %w", err)
	}
	return lines, nil
}

func WriteLines(w io.Writer, lines []string) error {
	writer := bufio.NewWriter(w)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error writing line: %w", err)
		}
	}

	return writer.Flush()
}
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
