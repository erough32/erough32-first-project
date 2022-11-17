package main

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

var pool *redis.Pool

func init() {
	pool = &redis.Pool{
		MaxIdle:     16,
		MaxActive:   0,
		IdleTimeout: 300,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", sessionsConfig("server"))
		},
	}
}

func addRead(bid string) {
	redisConn := pool.Get()
	defer redisConn.Close()

	var readMap string
	r, err := redis.String(redisConn.Do("Get", "read"))
	if err != nil {
		readMap = `{}`
	} else {
		readMap = r
	}
	m := make(map[string]int)
	json.Unmarshal([]byte(readMap), &m)
	m[bid] += 1
	mapString, _ := json.Marshal(m)
	redisConn.Do("Set", "read", string(mapString))
}

func setFav(bid string, add bool) {
	redisConn := pool.Get()
	defer redisConn.Close()

	var readMap string
	r, err := redis.String(redisConn.Do("Get", "fav"))
	if err != nil {
		readMap = `{}`
	} else {
		readMap = r
	}
	m := make(map[string]int)
	json.Unmarshal([]byte(readMap), &m)
	if add {
		m[bid] += 1
	} else {
		m[bid] -= 1
	}
	mapString, _ := json.Marshal(m)
	redisConn.Do("Set", "fav", string(mapString))
}

func addFavorites(bid string) {
	redisConn := pool.Get()
	defer redisConn.Close()

	var favoritesMap string
	r, err := redis.String(redisConn.Do("Get", "favorites"))
	if err != nil {
		favoritesMap = `{}`
	} else {
		favoritesMap = r
	}
	m := make(map[string]int)
	json.Unmarshal([]byte(favoritesMap), &m)
	m[bid] += 1
	mapString, _ := json.Marshal(m)
	redisConn.Do("Set", "favorites", string(mapString))
}

func addUser(uid string) {
	redisConn := pool.Get()
	defer redisConn.Close()

	redisConn.Do("SADD", "user", uid)
}
