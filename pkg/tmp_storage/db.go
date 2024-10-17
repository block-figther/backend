package tmp_storage

import "github.com/redis/go-redis/v9"

var rdb *redis.Client

func GetRedis() *redis.Client {
	if rdb == nil {
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	return rdb
}
