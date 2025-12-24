package middleware

import (
	"testing"
)

// TestGlossary_BasicTermProtection 测试基本术语保护
func TestGlossary_BasicTermProtection(t *testing.T) {
	terms := map[string]string{
		"AWS": "Amazon Web Services",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		// 模拟翻译，保持占位符不变
		return texts, nil
	})

	input := []string{"AWS is a cloud platform"}
	result, err := handler(input, "zh-CN")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Amazon Web Services is a cloud platform"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}
}

// TestGlossary_MultipleTerms 测试多个术语同时保护
func TestGlossary_MultipleTerms(t *testing.T) {
	terms := map[string]string{
		"AWS":        "Amazon Web Services",
		"Docker":     "Docker",
		"Kubernetes": "Kubernetes",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{
		"AWS is a cloud platform",
		"Use Docker and Kubernetes to deploy applications",
	}
	result, err := handler(input, "zh-CN")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"Amazon Web Services is a cloud platform",
		"Use Docker and Kubernetes to deploy applications",
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("test %d: expected %q, got %q", i, expected[i], result[i])
		}
	}
}

// TestGlossary_WordBoundaryMatching 测试完整单词匹配
func TestGlossary_WordBoundaryMatching(t *testing.T) {
	terms := map[string]string{
		"API": "API",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	// 测试：API 应该被匹配，但 API 的一部分不应该被匹配
	input := []string{
		"Use API to build applications",
		"The API is great",
	}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证占位符被正确替换
	if result[0] != "Use API to build applications" {
		t.Errorf("expected %q, got %q", "Use API to build applications", result[0])
	}
	if result[1] != "The API is great" {
		t.Errorf("expected %q, got %q", "The API is great", result[1])
	}
}

// TestGlossary_OverlappingTerms 测试重叠术语处理
func TestGlossary_OverlappingTerms(t *testing.T) {
	terms := map[string]string{
		"AI模型":      "AI模型",
		"AI":        "AI",
		"机器学习模型": "机器学习模型",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"开发AI模型和AI应用，以及机器学习模型"}
	result, err := handler(input, "en")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "开发AI模型和AI应用，以及机器学习模型"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}

	// 验证长术语优先匹配（通过占位符数量）
	// 如果 AI模型 被正确识别为一个术语，应该只有 3 个占位符
	// 而不是 4 个（AI模型 中的 AI 被单独匹配）
	if result[0] != expected {
		// 如果结果不对，说明术语没有被正确保护
		t.Errorf("overlapping terms not handled correctly")
	}
}

// TestGlossary_EmptyGlossary 测试空术语表
func TestGlossary_EmptyGlossary(t *testing.T) {
	terms := map[string]string{}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"AWS is a cloud platform"}
	result, err := handler(input, "zh-CN")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 空术语表应该不改变原文
	if result[0] != input[0] {
		t.Errorf("expected %q, got %q", input[0], result[0])
	}
}

// TestGlossary_NilGlossary 测试 nil 术语表
func TestGlossary_NilGlossary(t *testing.T) {
	var terms map[string]string = nil

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"AWS is a cloud platform"}
	result, err := handler(input, "zh-CN")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// nil 术语表应该不改变原文
	if result[0] != input[0] {
		t.Errorf("expected %q, got %q", input[0], result[0])
	}
}

// TestGlossary_PunctuationBoundaries 测试标点符号边界
func TestGlossary_PunctuationBoundaries(t *testing.T) {
	terms := map[string]string{
		"API": "应用程序接口",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{
		"Use API, SDK, and CLI.",
		"API is great!",
		"(API)",
	}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"Use 应用程序接口, SDK, and CLI.",
		"应用程序接口 is great!",
		"(应用程序接口)",
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("test %d: expected %q, got %q", i, expected[i], result[i])
		}
	}
}

// TestGlossary_CaseInsensitive 测试大小写不敏感
func TestGlossary_CaseInsensitive(t *testing.T) {
	terms := map[string]string{
		"AWS": "Amazon Web Services",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"AWS and aws"}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AWS 应该匹配所有大小写变体
	expected := "Amazon Web Services and Amazon Web Services"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}
}

// TestGlossary_MultipleTexts 测试多个文本批次
func TestGlossary_MultipleTexts(t *testing.T) {
	terms := map[string]string{
		"Docker": "Docker",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{
		"Use Docker containers",
		"Docker is lightweight",
		"Deploy with Docker",
	}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"Use Docker containers",
		"Docker is lightweight",
		"Deploy with Docker",
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("test %d: expected %q, got %q", i, expected[i], result[i])
		}
	}
}

// TestGlossary_PlaceholderPreservation 测试占位符在翻译中保持不变
func TestGlossary_PlaceholderPreservation(t *testing.T) {
	terms := map[string]string{
		"Docker": "Docker",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		// 模拟翻译服务返回的内容（占位符应该保持不变）
		// 在实际场景中，翻译服务应该保持 {ID_n} 不变
		return texts, nil
	})

	input := []string{"Docker containers are lightweight"}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 占位符应该被替换回术语
	expected := "Docker containers are lightweight"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}
}

// TestGlossary_TermWithSpecialCharacters 测试包含特殊字符的术语
func TestGlossary_TermWithSpecialCharacters(t *testing.T) {
	terms := map[string]string{
		"C++": "C++",
		"C#":  "C#",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"Learn C++ and C# programming"}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 特殊字符应该被正确转义和匹配
	expected := "Learn C++ and C# programming"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}
}

// TestGlossary_NoMatchInText 测试文本中不包含术语
func TestGlossary_NoMatchInText(t *testing.T) {
	terms := map[string]string{
		"AWS": "Amazon Web Services",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"Google is also a cloud platform"}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 没有匹配的术语，原文应该保持不变
	if result[0] != input[0] {
		t.Errorf("expected %q, got %q", input[0], result[0])
	}
}

// TestGlossary_LongerTermPriority 测试长术语优先匹配
func TestGlossary_LongerTermPriority(t *testing.T) {
	terms := map[string]string{
		"Machine Learning": "机器学习",
		"Machine":          "机器",
	}

	handler := Glossary(terms)(func(texts []string, toLang string) ([]string, error) {
		return texts, nil
	})

	input := []string{"Machine Learning is a subset of Machine"}

	result, err := handler(input, "zh-CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "Machine Learning" 应该被作为一个整体匹配
	// 而不是 "Machine" 被匹配两次
	expected := "机器学习 is a subset of 机器"
	if result[0] != expected {
		t.Errorf("expected %q, got %q", expected, result[0])
	}
}
