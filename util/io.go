package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func ReplaceWithDict(texts []string, dict map[string]string) []string {
	result := make([]string, len(texts))
	for i, text := range texts {
		for k, v := range dict {
			text = strings.ReplaceAll(text, k, v)
		}
		result[i] = text
	}
	return result
}

func FileExistsInParentDirs(name string) (string, error) {
	if filepath.IsAbs(name) || strings.Contains(name, string(filepath.Separator)) {
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			abs, _ := filepath.Abs(name)
			return abs, nil
		}
		abs, _ := filepath.Abs(name)
		return "", fmt.Errorf("no file: %s", abs)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getcwd: %w", err)
	}

	for {
		filePath := filepath.Join(currentDir, name)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			return filePath, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}

	return "", fmt.Errorf("no file even search upwards: %s", name)
}
