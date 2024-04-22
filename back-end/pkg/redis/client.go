package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"go-jellyfin-api/pkg/model"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	AddItems(items model.Items) error
	GetItem(key string) (model.ItemsElement, error)
	GetRandomNumberOfItems(noOfItems int) ([]model.ItemsElement, error)
}

type RedisClient struct {
	rdb *redis.Client
	ctx context.Context
}

func NewClient(context context.Context) RedisClient {
	ctx := context
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6739",
		Password: "",
		DB:       0,
	})
	return RedisClient{
		ctx: ctx,
		rdb: rdb,
	}
}

func (r RedisClient) AddItems(items model.Items) error {
	pipe := r.rdb.Pipeline()
	for _, i := range items.ItemElements {
		key := fmt.Sprintf("movie:%s", i.Id)
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
