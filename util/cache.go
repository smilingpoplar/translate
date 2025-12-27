package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tidwall/buntdb"
)

type Cache struct {
	name string
	db   *buntdb.DB
	ttl  time.Duration
}

func NewCache(name string, ttl time.Duration) (*Cache, error) {
	file := filepath.Join(os.TempDir(), name+".cache")
	db, err := buntdb.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache %s: %w", file, err)
	}
	return &Cache{name: name, db: db, ttl: ttl}, nil
}

// 生成缓存key: 服务名:目标语言:原文哈希
func (c *Cache) generateKey(toLang, text string) string {
	key := fmt.Sprintf("%s:%s:%s", c.name, toLang, text)
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:8]) // 只用前8字节，节省空间
}

func (c *Cache) Get(toLang, text string) (string, bool) {
	key := c.generateKey(toLang, text)
	var value string
	err := c.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		value = val
		return nil
	})

	if err != nil {
		return "", false
	}

	return value, true
}

func (c *Cache) Set(toLang, text, translated string) {
	key := c.generateKey(toLang, text)
	_ = c.db.Update(func(tx *buntdb.Tx) error {
		tx.Set(key, translated, &buntdb.SetOptions{
			Expires: true,
			TTL:     c.ttl,
		})
		return nil
	})
}

func (c *Cache) Close() error {
	return c.db.Close()
}
