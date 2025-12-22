package util

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// MaxCacheEntries 最大缓存条目数
	MaxCacheEntries = 5000
	// CacheExpiration 缓存过期时间（20分钟）
	CacheExpiration = 20 * time.Minute
)

// cacheEntry 缓存条目
type cacheEntry struct {
	Translated string    `json:"translated"`
	CreatedAt  time.Time `json:"created_at"`
}

// translationCache 翻译缓存（带LRU清理和持久化）
type translationCache struct {
	mu        sync.RWMutex
	store     map[string]*list.Element // key -> list元素
	lruList   *list.List               // LRU链表
	cacheFile string                   // 缓存文件路径
}

var globalCache *translationCache

func init() {
	// 初始化全局缓存，支持持久化
	cacheFile := filepath.Join(os.TempDir(), "translate-cache.json")
	globalCache = &translationCache{
		store:     make(map[string]*list.Element),
		lruList:   list.New(),
		cacheFile: cacheFile,
	}

	// 从文件加载缓存
	globalCache.loadFromFile()
}

// GenerateCacheKey 生成缓存key: 服务名:目标语言:原文哈希
func GenerateCacheKey(serviceName, toLang, text string) string {
	hash := sha256.Sum256([]byte(text))
	textHash := hex.EncodeToString(hash[:8]) // 只用前8字节，节省空间
	return fmt.Sprintf("%s:%s:%s", serviceName, toLang, textHash)
}

// Get 从缓存获取翻译
func (c *translationCache) Get(serviceName, toLang, text string) (string, bool) {
	key := GenerateCacheKey(serviceName, toLang, text)
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, found := c.store[key]
	if !found {
		return "", false
	}

	entry := elem.Value.(*cacheEntry)

	// 检查是否过期
	if time.Since(entry.CreatedAt) > CacheExpiration {
		// 删除过期条目
		c.lruList.Remove(elem)
		delete(c.store, key)
		return "", false
	}

	// 将访问的元素移到链表前端（最近使用）
	c.lruList.MoveToFront(elem)

	return entry.Translated, true
}

// Set 设置缓存
func (c *translationCache) Set(serviceName, toLang, text, translated string) {
	key := GenerateCacheKey(serviceName, toLang, text)
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果key已存在，先删除旧的
	if elem, found := c.store[key]; found {
		c.lruList.Remove(elem)
		delete(c.store, key)
	}

	// 检查是否需要清理缓存（超过最大条目数）
	for len(c.store) >= MaxCacheEntries {
		c.evictOldest()
	}

	// 添加新条目到链表前端
	entry := &cacheEntry{
		Translated: translated,
		CreatedAt:  time.Now(),
	}
	elem := c.lruList.PushFront(entry)
	c.store[key] = elem
}

// evictOldest 清理最旧的缓存条目
func (c *translationCache) evictOldest() {
	if elem := c.lruList.Back(); elem != nil {
		c.lruList.Remove(elem)
		// 遍历查找对应的key（这种情况很少发生）
		for key, val := range c.store {
			if val == elem {
				delete(c.store, key)
				return
			}
		}
	}
}

// loadFromFile 从文件加载缓存
func (c *translationCache) loadFromFile() {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.cacheFile)
	if err != nil {
		return
	}

	var entries map[string]cacheEntry
	if json.Unmarshal(data, &entries) != nil {
		return
	}

	// 加载所有条目到内存缓存（过期检查在Get时进行）
	for key, entry := range entries {
		c.store[key] = c.lruList.PushFront(&entry)
	}
}

// saveToFile 保存缓存到文件
func (c *translationCache) saveToFile() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make(map[string]cacheEntry, len(c.store))
	for key, elem := range c.store {
		entries[key] = *elem.Value.(*cacheEntry)
	}

	jsonData, err := json.Marshal(entries)
	if err != nil {
		return
	}

	// 原子性写入
	tempFile := c.cacheFile + ".tmp"
	if os.WriteFile(tempFile, jsonData, 0644) == nil {
		os.Rename(tempFile, c.cacheFile)
	}
}

// GetCacheSize 获取缓存条目数
func GetCacheSize() int {
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	return len(globalCache.store)
}

// Get 从缓存获取翻译
func Get(serviceName, toLang, text string) (string, bool) {
	return globalCache.Get(serviceName, toLang, text)
}

// Set 设置缓存
func Set(serviceName, toLang, text, translation string) {
	globalCache.Set(serviceName, toLang, text, translation)
}

// SaveCache 手动保存缓存到文件
func SaveCache() {
	globalCache.saveToFile()
}

// ClearCache 清空缓存
func ClearCache() {
	globalCache.mu.Lock()
	globalCache.store = make(map[string]*list.Element)
	globalCache.lruList = list.New()
	globalCache.mu.Unlock()
}
