package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/smilingpoplar/translate/util"
)

const kCacheTTL = 20 * time.Minute

// 生成缓存key: 服务名:目标语言:原文哈希
func generateCacheKey(serviceName, toLang, text string) string {
	hash := sha256.Sum256([]byte(text))
	textHash := hex.EncodeToString(hash[:8]) // 只用前8字节，节省空间
	return fmt.Sprintf("%s:%s:%s", serviceName, toLang, textHash)
}

func Cache(service string) Middleware {
	file := filepath.Join(os.TempDir(), service+".cache")
	cache, err := util.NewCache(file, kCacheTTL)
	if err != nil {
		log.Printf("Warning: failed to init cache for %s: %v", service, err)
		return func(handler Handler) Handler {
			return handler
		}
	}

	return func(handler Handler) Handler {
		return func(texts []string, toLang string) ([]string, error) {
			results := make([]string, len(texts))

			// 检查缓存，收集未缓存的文本
			uncached := make(map[int]string) // 索引 => 文本
			for i, text := range texts {
				key := generateCacheKey(service, toLang, text)
				if cached, found := cache.Get(key); found {
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

			// 合并且缓存已翻译的结果
			for i, translated := range translatedTexts {
				idx, text := indices[i], textsToTranslate[i]
				results[idx] = translated
				key := generateCacheKey(service, toLang, text)
				cache.Set(key, translated)
			}

			return results, nil
		}
	}
}
