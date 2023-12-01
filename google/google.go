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

	"github.com/smilingpoplar/translate/util"
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

func (g *Google) Translate(texts []string) ([]string, error) {
	return util.Retry(func() ([]string, error) {
		return g.translate(texts)
	})
}

func (g *Google) translate(texts []string) ([]string, error) {
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
	for _, text := range texts {
		postData.Add("q", text)
	}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(postData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userAgent := "GoogleTranslate/6.18.0.06.376053713 (Linux; U; Android 11; GM1900)"
	req.Header.Set("User-Agent", userAgent)

	// 发送请求
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, util.ErrTooManyRequests
		}
		return nil, fmt.Errorf("error http status: %s", http.StatusText(resp.StatusCode))
	}

	// 解析响应
	var data [][]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w, resp body: %s", err, string(body))
	}

	if len(data) == 0 || len(data[0]) != 2 {
		return nil, fmt.Errorf("error resp data: %v", data)
	}

	var result []string
	for _, line := range data {
		result = append(result, line[0])
	}
	return result, nil
}
