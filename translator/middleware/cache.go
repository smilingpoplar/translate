package middleware

import (
	"github.com/smilingpoplar/translate/util"
)

func Cache(service string) Middleware {
	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			results := make([]string, len(texts))

			// 检查缓存，收集未缓存的文本
			uncached := make(map[int]string) // 索引 => 文本
			for i, text := range texts {
				if cached, found := util.Get(service, toLang, text); found {
					results[i] = cached
				} else {
					uncached[i] = text
				}
			}

			// 所有文本都已缓存，直接返回
			if len(uncached) == 0 {
				return results, nil
			}

			// 构建待翻译文本列表
			indices := make([]int, 0, len(uncached))
			textsToTranslate := make([]string, 0, len(uncached))
			for idx, text := range uncached {
				indices = append(indices, idx)
				textsToTranslate = append(textsToTranslate, text)
			}

			// 调用翻译服务
			translatedTexts, err := handler(textsToTranslate, toLang)
			if err != nil {
				return nil, err
			}

			// 缓存并合并已翻译的结果
			for i, translated := range translatedTexts {
				idx, text := indices[i], textsToTranslate[i]
				results[idx] = translated
				util.Set(service, toLang, text, translated)
			}
			util.SaveCache()

			return results, nil
		}
	}
}
