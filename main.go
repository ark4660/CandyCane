package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"CandyCane/storage"
	"CandyCane/models"
	"encoding/json"
	"context"
    "time"
    "log"
    "fmt"
)

func main() {
	app := fiber.New()
	app.Use(session.New())
	app.Use(cors.New(cors.Config{
    	AllowOrigins: []string{"http://localhost:5173"},
     	AllowHeaders: []string{"*"},
      	AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))


	ctx := context.Background()
	s3Client, err := storage.NewClient()

	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
        Addr:	  "redis-11310.crce194.ap-seast-1-1.ec2.cloud.redislabs.com:11310",
        Password: "D3jf8P5U3RPXgFP5NPjFIFeObUoAJtYB", // No password set
        DB:		  0,  // Use default DB
        Protocol: 2,  // Connection protocol
    })

	api := app.Group("/api")
	video := api.Group("/video")

	video.Post("/sessionInformation", func(c fiber.Ctx) error {
		p := new(models.RedisSessionModel)
		if err := c.Bind().Body(p); err != nil {
        	return err
    	}
     	key := p.WatchSessionId
      	val, err := rdb.Get(ctx, key).Result()
        if err != nil {
            if err == redis.Nil {
                fmt.Println("Key does not exist")
            } else {
                panic(err)
            }
        } else {
            fmt.Println("Value:", val)
        }
        var sessionData map[string]interface{}
        err = json.Unmarshal([]byte(val), &sessionData)
        if err != nil {
        	panic(err)
        }

        lastPos, ok := sessionData["LastPosition"].(float64)
        if !ok {
        	lastPos = 0
        }

        if dt := time.Now().Sub(time.Unix(int64(sessionData["LastTime"].(float64)), 0)); dt > 10 {
       		if t := p.LastPosition - lastPos ; t > 9 && t < 15 {
             	sessionData["TotalVerifiedSeconds"] = sessionData["TotalVerifiedSeconds"].(float64) + 10
         	}
          	sessionData["LastPosition"] = p.LastPosition
        }

        fmt.Println(sessionData)
        jsonData, err := json.Marshal(sessionData)
		if err != nil {
			panic(err)
		}
		pushKey := sessionData["WatchSessionId"].(string)
		err = rdb.Set(ctx, pushKey, jsonData, 30*time.Second).Err()
		if err != nil {
          		panic(err)
		}
        return c.Status(fiber.StatusOK).SendString("ok")

	})


	video.Get("/:id", func(c fiber.Ctx) error {
		sess := session.FromContext(c)
		var UUID interface{}

		if v := sess.Get("UUID"); v == nil {
			UUID = 124323423412
		} else {
			UUID = v
		}

		videoId := c.Params("id")
		session, err := uuid.NewRandom()
		sessionId := session.String()

		sessionPayload := map[string]interface{}{
			"WatchSessionId": sessionId,
			"Uuid": UUID,
			"VideoId": videoId,
			"LastPosition": 0,
			"TotalVerifiedSeconds": 0,
			"LastTime": time.Now().Unix(),
		}
		jsonData, err := json.Marshal(sessionPayload)
		if err != nil {
			panic(err)
		}
		key := fmt.Sprint(sessionPayload["WatchSessionId"])
		err = rdb.Set(ctx, key, jsonData, 30*time.Second).Err()
		if err != nil {
    		panic(err)
		}

		fmt.Println("added to database")
		videosKey := "videos/" + videoId + ".mkv"
		videosUrl, err := storage.GeneratePresignedURL(s3Client, videosKey)

		if err != nil {
			panic(err)
		}
		reponsePayload := models.ResponsePayload{
			Url: videosUrl,
			SessionId: sessionId,
		}
		return c.JSON(reponsePayload)

	})

	log.Fatal(app.Listen(":3000"))
}
