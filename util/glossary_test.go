package util

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadGlossary_ValidFile 测试加载有效的术语表文件
func TestLoadGlossary_ValidFile(t *testing.T) {
	// 创建临时术语表文件
	tmpDir := t.TempDir()
	glossaryFile := filepath.Join(tmpDir, "glossary.csv")
	content := `AWS,Amazon Web Services
Docker,Docker
Kubernetes,Kubernetes
AI模型,AI模型
`
	if err := os.WriteFile(glossaryFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// 加载术语表
	glossary, err := LoadGlossary(glossaryFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证加载的术语
	expectedCount := 4
	if len(glossary) != expectedCount {
		t.Errorf("expected %d terms, got %d", expectedCount, len(glossary))
	}

	expectedTerms := map[string]string{
		"AWS":  "Amazon Web Services",
		"AI模型": "AI模型",
	}

	for from, expectedTo := range expectedTerms {
		if to, ok := glossary[from]; !ok {
			t.Errorf("term %q not found in glossary", from)
		} else if to != expectedTo {
			t.Errorf("term %q: expected to %q, got %q", from, expectedTo, to)
		}
	}
}

// TestLoadGlossary_EmptyFile 测试加载空文件
func TestLoadGlossary_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	glossaryFile := filepath.Join(tmpDir, "empty.csv")

	// 创建空文件
	if err := os.WriteFile(glossaryFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, err := LoadGlossary(glossaryFile)
	// 空文件会导致 CSV 解析错误
	if err == nil {
		t.Error("expected error for empty file, got nil")
	}
}

// TestLoadGlossary_EmptyPath 测试空路径
func TestLoadGlossary_EmptyPath(t *testing.T) {
	glossary, err := LoadGlossary("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if glossary != nil {
		t.Errorf("expected nil for empty path, got %v", glossary)
	}
}

// TestLoadGlossary_NonExistentFile 测试不存在的文件
func TestLoadGlossary_NonExistentFile(t *testing.T) {
	_, err := LoadGlossary("/non/existent/file.csv")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

// TestLoadGlossary_InvalidCSV 测试单列 CSV（缺少目标列）
func TestLoadGlossary_InvalidCSV(t *testing.T) {
	tmpDir := t.TempDir()
	glossaryFile := filepath.Join(tmpDir, "invalid.csv")

	// 创建单列的 CSV 文件（缺少 to 列）
	content := `AWS
Docker
Kubernetes
`
	if err := os.WriteFile(glossaryFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	glossary, err := LoadGlossary(glossaryFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 单列 CSV 会被解析，from 字段有值，to 字段为空
	// 由于 from 不为空，会被添加到术语表中
	expectedCount := 3
	if len(glossary) != expectedCount {
		t.Errorf("expected %d terms for single-column CSV, got %d", expectedCount, len(glossary))
	}

	// 验证 from 存在，但 to 为空字符串
	if to, ok := glossary["AWS"]; !ok {
		t.Error("AWS should be in glossary")
	} else if to != "" {
		t.Errorf("AWS to should be empty, got %q", to)
	}
}

// TestLoadGlossary_SkipEmptyFrom 测试跳过空 from
func TestLoadGlossary_SkipEmptyFrom(t *testing.T) {
	tmpDir := t.TempDir()
	glossaryFile := filepath.Join(tmpDir, "glossary.csv")

	// 创建包含空 from 的文件
	content := `,Target1
AWS,Amazon Web Services
,Target2
Docker,Docker
`
	if err := os.WriteFile(glossaryFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	glossary, err := LoadGlossary(glossaryFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 空 from 的行应该被跳过
	expectedCount := 2
	if len(glossary) != expectedCount {
		t.Errorf("expected %d terms (empty froms skipped), got %d", expectedCount, len(glossary))
	}

	// 验证空 from 不在术语表中
	if _, ok := glossary[""]; ok {
		t.Error("empty from should not be in glossary")
	}
}

// TestLoadGlossary_UTF8Encoding 测试 UTF-8 编码
func TestLoadGlossary_UTF8Encoding(t *testing.T) {
	tmpDir := t.TempDir()
	glossaryFile := filepath.Join(tmpDir, "glossary.csv")

	// 创建包含中文、日文、emoji 的术语表
	content := `AWS,Amazon Web Services
人工智能,人工智能
機械学習,機械学習
😀,😀
`
	if err := os.WriteFile(glossaryFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	glossary, err := LoadGlossary(glossaryFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCount := 4
	if len(glossary) != expectedCount {
		t.Errorf("expected %d terms, got %d", expectedCount, len(glossary))
	}

	// 验证 UTF-8 字符被正确处理
	if glossary["人工智能"] != "人工智能" {
		t.Error("UTF-8 Chinese characters not handled correctly")
	}
	if glossary["😀"] != "😀" {
		t.Error("Emoji not handled correctly")
	}
}

// TestGeneratePlaceholder 测试占位符生成
func TestGeneratePlaceholder(t *testing.T) {
	tests := []struct {
		id       int
		expected string
	}{
		{0, "{ID_0}"},
		{1, "{ID_1}"},
		{10, "{ID_10}"},
		{999, "{ID_999}"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GeneratePlaceholder(tt.id)
			if result != tt.expected {
				t.Errorf("GeneratePlaceholder(%d) = %q, want %q", tt.id, result, tt.expected)
			}
		})
	}
}

// TestBuildWordBoundaryRegex 测试单词边界正则表达式
// 注意：\b 只对 ASCII 字母数字字符 [a-zA-Z0-9_] 起作用
// 对于中文、日文、特殊字符（如 C++ 中的 +）等，\b 不会正确匹配边界
// 这是正则表达式的固有限制，建议用户：
// 1. 对于 ASCII 术语：使用单词边界，效果最好
// 2. 对于中文术语：仍然可以使用，但可能匹配包含该术语的更长文本
// 3. 对于特殊字符术语：确保术语在上下文中是唯一的
func TestBuildWordBoundaryRegex(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		match    []string
		notMatch []string
	}{
		{
			name:  "simple word",
			word:  "API",
			match: []string{"API", "API is great", "Use API", "API.", "(API)"},
			notMatch: []string{
				"APIS", "APIs", "nAPI",
				"clipboard", "swagger",
			},
		},
		{
			name: "word with numbers",
			word: "AWS3",
			match: []string{
				"AWS3", "Use AWS3", "AWS3 is great",
			},
			notMatch: []string{
				"AWS33", "AWS3d",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := BuildWordBoundaryRegex(tt.word)
			if err != nil {
				t.Fatalf("BuildWordBoundaryRegex(%q) error: %v", tt.word, err)
			}

			// 测试应该匹配的字符串
			for _, text := range tt.match {
				if !regex.MatchString(text) {
					t.Errorf("regex should match %q, but it doesn't", text)
				}
			}

			// 测试不应该匹配的字符串
			for _, text := range tt.notMatch {
				if regex.MatchString(text) {
					t.Errorf("regex should not match %q, but it does", text)
				}
			}
		})
	}
}

// TestBuildWordBoundaryRegex_SpecialCharacters 测试特殊字符转义
func TestBuildWordBoundaryRegex_SpecialCharacters(t *testing.T) {
	// 包含正则表达式特殊字符的术语
	specialWords := []string{
		"C++",    // + 是重复字符
		"C#",     // # 是注释字符（在某些正则引擎中）
		".NET",   // . 是通配符
		"AWS+SDK", // + 是重复字符
	}

	for _, word := range specialWords {
		_, err := BuildWordBoundaryRegex(word)
		if err != nil {
			t.Errorf("BuildWordBoundaryRegex(%q) should not error, got: %v", word, err)
		}
	}
}
