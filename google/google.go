// google翻译的web接口参数来自：
// https://github.com/foyoux/pygtrans/blob/main/src/pygtrans/Translate.py

package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const BaseURL = "https://translate.google.com/translate_a/t"

type Google struct {
	Client *http.Client
}

func New() *Google {
	return &Google{
		Client: &http.Client{},
	}
}

func (g *Google) Translate(text string) (string, error) {
	// 构造请求
	queryParams := url.Values{}
	queryParams.Set("sl", "auto")
	queryParams.Set("tl", "zh-CN")
	queryParams.Set("ie", "UTF-8")
	queryParams.Set("oe", "UTF-8")
	queryParams.Set("client", "at")
	queryParams.Set("dj", "1")
	queryParams.Set("format", "html")
	queryParams.Set("v", "1.0")
	apiURL := fmt.Sprintf("%s?%s", BaseURL, queryParams.Encode())

	postData := url.Values{}
	postData.Set("q", text)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(postData.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userAgent := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"
	req.Header.Set("User-Agent", userAgent)

	// 发送请求
	resp, err := g.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	// 解析响应
	var data [][]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %w, resp body: %s", err, string(body))
	}

	if len(data) == 0 || len(data[0]) != 2 {
		return "", fmt.Errorf("error resp data: %v", data)
	}

	return data[0][0], nil
}
