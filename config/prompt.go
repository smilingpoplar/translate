package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func GetPrompt(texts []string, toLang string) (string, error) {
	str := getPromptTemplate()
	str = strings.ReplaceAll(str, "{{lang}}", toLang)
	jsonStr, err := getJson(texts)
	if err != nil {
		return "", fmt.Errorf("error getting prompt: %v", err)
	}
	str = strings.ReplaceAll(str, "{{json}}", jsonStr)
	return str, nil
}

func getPromptTemplate() string {
	data, err := embedFS.ReadFile("prompt.txt")
	if err != nil {
		log.Fatalf("error reading txt file: %v", err)
	}
	return string(data)
}

type Translation struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

func getJson(texts []string) (string, error) {
	trans := []Translation{}
	for i, text := range texts {
		trans = append(trans, Translation{
			ID:   i,
			Text: text,
		})
	}
	jsonData, err := json.Marshal(trans)
	if err != nil {
		return "", fmt.Errorf("error marshaling json: %v", err)
	}
	return string(jsonData), nil
}
