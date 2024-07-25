package cacheHelper

import (
	"context"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Redis struct {
	Databases []RedisDatabase
}

type RedisDatabase struct {
	Name      string
	Env       string
	Addr      string
	Password  string
	DB        int
	KeyPrefix string
}

func (r Redis) Get() map[string]RedisDatabase {
	databaseMap := make(map[string]RedisDatabase, len(r.Databases))
	for _, database := range r.Databases {
		databaseMap[database.Name] = database
	}
	return databaseMap
}

func (rd *RedisDatabase) Client() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     rd.Addr,
		Password: rd.Password, // no password set
		DB:       rd.DB,       // use default DB
	})
}

func (rd *RedisDatabase) Key(key string) string {
	return rd.KeyPrefix + ":" + key
}

func (rd *RedisDatabase) GetCache(key string, createNewCache func() (any, time.Duration), cacheInterface any) (any, error) {
	ctx := context.Background()
	var cacheBytes []byte
	var cacheString string
	var err error
	var client = rd.Client()
	if cacheString, err = client.Get(ctx, key).Result(); errors.Is(err, redis.Nil) {
		newCacheInterface, timeout := createNewCache()
		cacheBytes, err = json.Marshal(newCacheInterface)
		if err = client.Set(ctx, key, cacheBytes, timeout).Err(); err != nil {
			return nil, err
		} else {
			return newCacheInterface, nil
		}
	} else if err != nil {
		return nil, err
	} else {
		if err = json.Unmarshal([]byte(cacheString), &cacheInterface); err != nil {
			return nil, err
		}
		return cacheInterface, nil
	}
}
