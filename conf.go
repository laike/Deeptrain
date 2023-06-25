package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
)

func ConnectRedis() *redis.Client {
	// connect to redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis:addr"),
		Password: viper.GetString("redis:password"),
		DB:       viper.GetInt("redis:db"),
	})
	_, err := rdb.Ping(context.Background()).Result()

	if err != nil {
		log.Fatalln("Failed to connect to Redis server: ", err)
	} else {
		log.Println("Connected to Redis server successfully")
	}

	if viper.GetBool("flush") {
		rdb.FlushAll(context.Background())
		log.Println("Flushed all cache")
	}

	return rdb
}

func ConnectMySQL() *sql.DB {
	// connect to MySQL
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s",
		viper.GetString("mysql:user"),
		viper.GetString("mysql:password"),
		viper.GetString("mysql:host"),
		viper.GetInt("mysql:port"),
		viper.GetString("mysql:db"),
	))
	if err != nil {
		log.Fatalln("Failed to connect to MySQL server: ", err)
	} else {
		log.Println("Connected to MySQL server successfully")
	}

}
