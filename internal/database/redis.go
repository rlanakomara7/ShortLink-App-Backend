package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
