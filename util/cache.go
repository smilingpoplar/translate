package util

import (
	"fmt"
	"time"

	"github.com/tidwall/buntdb"
)

type Cache struct {
	db  *buntdb.DB
	ttl time.Duration
}

func NewCache(file string, ttl time.Duration) (*Cache, error) {
	db, err := buntdb.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open db %s: %w", file, err)
	}
	return &Cache{db: db, ttl: ttl}, nil
}

func (c *Cache) Get(key string) (string, bool) {
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

func (c *Cache) Set(key, value string) {
	_ = c.db.Update(func(tx *buntdb.Tx) error {
		tx.Set(key, value, &buntdb.SetOptions{
			Expires: true,
			TTL:     c.ttl,
		})
		return nil
	})
}

func (c *Cache) Close() error {
	return c.db.Close()
}
