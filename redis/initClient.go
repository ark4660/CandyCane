package redis_self

import (
	"github.com/redis/go-redis/v9"
)

func Init_rdb() (*redis.Client){
	rdb := redis.NewClient(&redis.Options{
        Addr:	  "redis-11310.crce194.ap-seast-1-1.ec2.cloud.redislabs.com:11310",
        Password: "D3jf8P5U3RPXgFP5NPjFIFeObUoAJtYB", // No password set
        DB:		  0,  // Use default DB
        Protocol: 2,  // Connection protocol
    })
	return rdb
}
