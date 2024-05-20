package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-jellyfin-api/pkg/model"
)

type Client interface {
	AddItems(items *model.Items) error
	GetItem(key string) (model.ItemsElement, error)
	GetRandomNumberOfItems(noOfItems int) ([]model.ItemsElement, error)
	GetItemsByKeys(keys []string) ([]model.ItemsElement, error)
	FindKeyByPartialTitle(title string) ([]model.ItemsElement, error)
}

type RedisClient struct {
	rdb *redis.Client
	ctx context.Context
}

func NewClient(context context.Context) RedisClient {
	ctx := context
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	return RedisClient{
		ctx: ctx,
		rdb: rdb,
	}
}

func (r RedisClient) AddItems(items *model.Items) error {
	pipe := r.rdb.Pipeline()
	for _, i := range items.ItemElements {
		title := i.NormaliseTitle()
		key := fmt.Sprintf("movie:%s:%s", title, i.Id)
		structBytes, err := json.Marshal(i)
		if err != nil {
			fmt.Println(err)
			continue
		}
		pipe.Set(r.ctx, key, structBytes, 0)
	}
	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r RedisClient) GetItem(key string) (model.ItemsElement, error) {
	item, err := r.rdb.Get(r.ctx, key).Result()
	if err != nil {
		fmt.Println("Failed to get item for key " + key)
		return model.ItemsElement{}, err
	}
	var itemElement model.ItemsElement
	jsonErr := json.Unmarshal([]byte(item), &itemElement)
	if jsonErr != nil {
		fmt.Println("Failed to marshal item for key " + key)
		return model.ItemsElement{}, jsonErr
	}
	return itemElement, nil

}

func (r RedisClient) FindKeyByPartialTitle(title string) ([]model.ItemsElement, error) {

	keys, err := r.rdb.Keys(r.ctx, "movie:*"+title+"*").Result()
	if err != nil {
		fmt.Println("Failed to find keys for partial title " + title)
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	var items []model.ItemsElement
	for _, key := range keys {
		unmarshalledItem, err := r.GetItem(key)
		if err != nil {
			return nil, err
		}
		items = append(items, unmarshalledItem)
	}
	return items, nil
}

func (r RedisClient) GetRandomNumberOfItems(noOfItems int) ([]model.ItemsElement, error) {
	var items []model.ItemsElement
	for i := 0; i < noOfItems; i++ {
		item, err := r.rdb.RandomKey(r.ctx).Result()
		if err != nil {
			fmt.Println("Failed to get RandomKey from redis")
			return nil, err
		}
		unmarshalledItem, err := r.GetItem(item)
		if err != nil {
			return nil, err
		}
		items = append(items, unmarshalledItem)
	}

	return items, nil
}

// GetItemsByKeys TODO this is ai generated, write tests and stuff
func (r RedisClient) GetItemsByKeys(keys []string) ([]model.ItemsElement, error) {
	var cursor uint64
	var n int
	var err error
	var element model.ItemsElement
	var items []model.ItemsElement

	pipe := r.rdb.Pipeline()

	for _, pattern := range keys {
		for {
			fmt.Println("searching for key " + pattern)
			var keys []string

			keys, cursor, err = r.rdb.Scan(context.Background(), cursor, pattern, 10).Result()

			if err != nil {
				return nil, err
			}

			for _, key := range keys {
				pipe.Get(context.Background(), key)
				n++
			}

			if n > 0 {
				results, err := pipe.Exec(context.Background())
				if err != nil && !errors.Is(err, redis.Nil) {
					return nil, err
				}

				for _, result := range results {
					val, _ := result.(*redis.StringCmd).Result()
					if err := json.Unmarshal([]byte(val), &element); err == nil {
						items = append(items, element)
					}
				}
			}

			if cursor == 0 {
				break
			}
		}
	}
	return items, nil
}
