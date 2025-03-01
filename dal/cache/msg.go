package cache

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
)

type MsgCache struct {
	cache *cache.Cache
}

var msgCache *MsgCache

func GetMsgCache() *MsgCache {
	return msgCache
}

func NewMsgCache() {
	msgCache = &MsgCache{
		cache: cache.New(10*time.Hour, 10*time.Hour),
	}
}

func (c *MsgCache) Process(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *MsgCache) IfProcessed(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *MsgCache) Delete(key string) {
	c.cache.Delete(key)
}
