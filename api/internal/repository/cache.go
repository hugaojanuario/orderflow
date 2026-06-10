package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache encapsula o cache de leitura no redis
type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Get retorna string vazia quando a chave não existe
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("erro ao ler do cache: %w", err)
	}

	return value, nil
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("erro ao gravar no cache: %w", err)
	}

	return nil
}
