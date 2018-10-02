package gateway

import (
	"time"

	"github.com/go-redis/redis"
)

func getUrlInRedis(serviceEndpoint string) (address <-chan string) {
	c := make(chan string)

	go func(svcName string) {
		client := redis.NewClient(
			&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			})

		res, err := client.Get(serviceEndpoint).Result()
		if err != nil {
			c <- ""
		}
		c <- res
	}(serviceEndpoint)

	return c
}

func setUrlInRedis(serviceEndpoint, address string) {
	go func(svcEnp, addr string) {
		client := redis.NewClient(
			&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			})
		client.Set(serviceEndpoint, address, time.Duration(time.Hour*24))
	}(serviceEndpoint, address)
}
