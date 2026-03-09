package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"gopkg.in/vansante/go-ffprobe.v2"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"CandyCane/storage"
	openai "github.com/sashabaranov/go-openai"
	"CandyCane/models"
	"CandyCane/databases"
	"CandyCane/firebase"
	"bytes"
	"CandyCane/transcript"
	"CandyCane/middleware"
	"strconv"
	"CandyCane/redis"
	"encoding/json"
	"context"
    "time"
    "log"
    "fmt"
    "io"
    "os"



)

type teeReadCloser struct {
    r io.Reader    // the TeeReader
    c io.Closer    // the thing to close
}

type ChatCompletionMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func (t *teeReadCloser) Read(p []byte) (int, error) {
    return t.r.Read(p)
}

func (t *teeReadCloser) Close() error {
    return t.c.Close()
}

func main() {
	app := fiber.New(fiber.Config{
    BodyLimit: 1024 * 1024 * 1024, // 50 MB
})
	app.Use(session.New())
	app.Use(cors.New(cors.Config{
    	AllowOrigins: []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175"},
     	AllowHeaders: []string{"*"},
      	AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	videoListDB := database.InitDatabase()
	videoSessionDB := database.InitDatabaseSessionHistory()


	s3Client, err:= storage.NewClient()
	if err != nil {
		panic(err)
	}

	firebaseApp := firebase_self.InitializeAppDefault()
	firebaseAuthClient := firebase_self.GetAuthClient(firebaseApp)

	rdb:= redis_self.Init_rdb()
	aiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	api := app.Group("/api")
	video := api.Group("/video")

	api.Post("/sessionInformation", func(c fiber.Ctx) error {
		p := new(models.RedisSessionModel)
		if err := c.Bind().Body(p); err != nil {
        	return err
    	}
     	key := p.WatchSessionId
      	val, err := rdb.Get(c.Context(), "sessions:"+key).Result()
        if err != nil {
            if err == redis.Nil {
                fmt.Println("Key does not exist")
                return err
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
		err = rdb.Set(c.Context(), "sessions:"+pushKey, jsonData, 15*time.Second).Err()
		if err != nil {
          		panic(err)
		}
		err = rdb.Set(c.Context(), "sessions:"+pushKey+":backup", jsonData, 20*time.Second).Err()
		if err != nil {
          		panic(err)
		}
        return c.Status(fiber.StatusOK).SendString("ok")

	})


	video.Get("/:id",middleware.VerifyTokenIdMiddleWare(firebaseAuthClient), func(c fiber.Ctx) error {
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
		sess := session.FromContext(c)
		UUID := sess.Get("UUID").(string)
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
		getCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()

		key := fmt.Sprint(sessionPayload["WatchSessionId"])
		err = rdb.Set(getCtx, "sessions:"+key, jsonData, 15*time.Second).Err()
		if err != nil {
			return err
		}

		fmt.Println("added to database")
		videosKey := "videos/" + videoId + "/video"
		videosUrl, err := storage.GeneratePresignedURL(s3Client, videosKey)
		if err != nil {
			panic(err)
		}
		data, err := rdb.HGetAll(redisCtx, "videos:"+videoId).Result()
		if err != nil {
    		log.Println(err)
		}
		reponsePayload := models.ResponsePayload{
			Url: videosUrl,
			SessionId: sessionId,
			Title: data["title"],
			Description: data["description"],
		}
		return c.JSON(reponsePayload)

	})

	video.Post("/upload", middleware.VerifyTokenIdMiddleWare(firebaseAuthClient) ,func(c fiber.Ctx) error {
		sess := session.FromContext(c)
		UUID := sess.Get("UUID")
		video, err := uuid.NewRandom()
		videoId := video.String()
		name := c.FormValue("name")
		description := c.FormValue("description")
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()

		header, err := c.FormFile("file")
		file, err := header.Open()
		if err != nil {
    		return err
		}
		defer file.Close()


		thumbnailHeader, err := c.FormFile("thumbnail")
		thumbnail, err := thumbnailHeader.Open()
		if err != nil {
    		return err
		}
		defer thumbnail.Close()

		probeCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    	defer cancel()

		data, err := ffprobe.ProbeReader(probeCtx, file)
		if err != nil {
    		log.Panicf("Error getting data: %v", err)
		}
		jsonData := map[string]interface{}{
			"video_id": videoId,
			"title": name,
			"thumbnail_url": "https://candycane.sgp1.cdn.digitaloceanspaces.com/videos/" + videoId + "/thumbnail",
			"description": description,
			"created_at": float64(time.Now().Unix()),
			"duration": data.Format.DurationSeconds,
		}
		rdb.HSet(redisCtx, "videos:"+videoId, jsonData)
		rdb.ZAdd(redisCtx, "videos:index", redis.Z{
			Score: float64(time.Now().Unix()),
			Member: videoId,
		})

		dataPayload := database.VideoInformation{
			Id: videoId,
			UUID: UUID.(string),
			Title: name,
			Description: description,
			CreationDate: time.Now().Unix(),
			VideoLength: data.Format.DurationSeconds,
		}
		file.Seek(0, io.SeekStart)
		err = storage.UploadData(file, videoId, s3Client)
		if err != nil {
			panic(err)
		}
		// Not financialy viable yet
		//file.Seek(0, io.SeekStart)
		//readByte, _ := io.ReadAll(file)
		//go func() {
		//	storage.GetTranscriptUpload(readByte, videoId, s3Client)
		//}
		transcriptHeader, err := c.FormFile("transcript")
		var transcriptReader io.ReadCloser
		if transcriptHeader != nil {

			transcriptFile, err := transcriptHeader.Open()
			if err != nil {
    		return err
			}
			defer transcriptFile.Close()

			err = storage.UploadTranscriptVtt(transcriptFile, videoId, s3Client)
			if err != nil {
				panic(err)
			}

			transcriptFile.Seek(0, io.SeekStart)
			transcriptReader, err = transcript.VttToJSON(transcriptFile)
			if err != nil {
				panic(err)
			}
			buf, _ := io.ReadAll(transcriptReader)
			log.Printf("transcript JSON size: %d bytes", len(buf))
			log.Printf("transcript JSON preview: %s", string(buf[:min(len(buf), 200)]))
			err = storage.UploadTranscriptJson(io.NopCloser(bytes.NewReader(buf)), videoId, s3Client)
			if err != nil {
				panic(err)
			}
		}
		err = storage.UploadThumbnail(thumbnail, videoId, s3Client)
		if err != nil {
			panic(err)
		}
		database.Upload(videoListDB, &dataPayload)
		return c.Status(fiber.StatusOK).SendString("Successfully uploaded")
	})

	api.Get("/progressMap", middleware.VerifyTokenIdMiddleWare(firebaseAuthClient),func(c fiber.Ctx) error {
		sess := session.FromContext(c)
		UUID := sess.Get("UUID")
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()

		val, err := rdb.Get(redisCtx, "user:"+UUID.(string)+":history").Result()
        if err != nil {
            if err == redis.Nil {
                fmt.Println("Key does not exist")
                jsonData, _ := database.GetSession(videoSessionDB, UUID.(string))
                err = rdb.Set(redisCtx, "user:"+UUID.(string)+":history", jsonData, 48*time.Hour).Err()
                if err != nil {
                	panic(err)
                }
                fmt.Println("Pushed key to redis cache")
                var history []map[string]interface{}
                err := json.Unmarshal(jsonData, &history)
                if err != nil {
                    return c.Status(500).SendString("invalid JSON data")
                }
                return c.JSON(history)
            } else {
                panic(err)
            }
        }
        if val == "[]" {
       		fmt.Println("Key does not exist")
         	jsonData, _ := database.GetSession(videoSessionDB, UUID.(string))
          	err = rdb.Set(redisCtx, "user:"+UUID.(string)+":history", jsonData, 48*time.Hour).Err()
           	if err != nil {
           		panic(err)
            }
            fmt.Println("Pushed key to redis cache")
            var history []map[string]interface{}
            err := json.Unmarshal(jsonData, &history)
            if err != nil {
            	return c.Status(500).SendString("invalid JSON data")
            }
            return c.JSON(history)
        }
        var history []map[string]interface{}
        if err := json.Unmarshal([]byte(val), &history); err != nil {
        	return err
        }
        return c.JSON(history)
	})
	api.Post("/specifiedPagination" ,func(c fiber.Ctx) error {
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
    	var items []string
     	if err := c.Bind().JSON(&items); err != nil {
      		return err
      	}
		pipe := rdb.Pipeline()
		cmds := make([]*redis.MapStringStringCmd, len(items))
        for i, id := range items {
        	cmds[i] = pipe.HGetAll(redisCtx, "videos:"+id)
        }
        pipe.Exec(redisCtx)
        var videos []interface{}
        for _, cmd := range cmds {
        	data, err := cmd.Result()
         	if err != nil {
          		return err
          	}
         	videos = append(videos, data)
        }
        return c.JSON(videos)
	})
	api.Get("/pagination" ,func(c fiber.Ctx) error {
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
     	limit, _ := strconv.Atoi(c.Query("limit", "12"))
		cursor := c.Query("cursor", "")


		maxScore := "+inf"
    	if cursor != "" {
        	maxScore = cursor
     	}

      	videoIDs, err := rdb.ZRevRangeByScore(redisCtx, "videos:index", &redis.ZRangeBy{
             Max:    maxScore,
             Min:    "-inf",
             Offset: 0,
             Count:  int64(limit + 1),
         }).Result()
         if err != nil {
             return c.Status(500).JSON(fiber.Map{"error": err.Error()})
         }
         hasMore := len(videoIDs) > limit
         if hasMore {
         	videoIDs = videoIDs[:limit]
         }

         pipe := rdb.Pipeline()
         cmds := make([]*redis.MapStringStringCmd, len(videoIDs))
         for i, id := range videoIDs {
         	cmds[i] = pipe.HGetAll(redisCtx, "videos:"+id)
         }
         pipe.Exec(redisCtx)

         videos := make([]map[string]string, 0, len(videoIDs))
         for _, cmd := range cmds {
         	data, err := cmd.Result()
          	if err == nil && len(data) > 0 {
           		videos = append(videos, data)
           }
         }

         var nextCursor *string
             if hasMore {
                 score, _ := rdb.ZScore(redisCtx, "videos:index", videoIDs[len(videoIDs)-1]).Result()
                 s := strconv.FormatFloat(score-0.001, 'f', -1, 64)
                 nextCursor = &s
             }

		return c.JSON(fiber.Map{
			"data": videos,
			"next_cursor": nextCursor,
			"has_more": hasMore,
		})
	})
	api.Get("/weekly", middleware.VerifyTokenIdMiddleWare(firebaseAuthClient), func(c fiber.Ctx) error {
		redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    	defer cancel()
		var array []string
		sess := session.FromContext(c)
		data, err := rdb.HGetAll(redisCtx, "user:"+sess.Get("UUID").(string)+":watchtime").Result()
		if err != nil {
    		log.Println(err)
		}
		yearday := time.Now().YearDay()
		weekday := int(time.Now().Weekday())
		for i := 0; i < 6; i++ {
			array = append(array, data[strconv.Itoa(yearday - weekday + 1 + i)])
		}
		array = append(array, data[strconv.Itoa(yearday - weekday)])
		return c.JSON(array)
	})
	api.Post("/ai", middleware.VerifyTokenIdMiddleWare(firebaseAuthClient),  func(c fiber.Ctx) error {
		redisCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		sess := session.FromContext(c)
		UUID := sess.Get("UUID")
		var chats []string
		body := c.Body()
		log.Println("RAW BODY:", string(body))
		if err := json.Unmarshal(body, &chats); err != nil {
    		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		data, err := rdb.HGetAll(redisCtx, "user:"+UUID.(string)+":ChatSession:"+chats[1]).Result()
		if len(data) == 0 {
			if err == nil {
				log.Println("log4")
				message:= []openai.ChatCompletionMessage{
					{Role: "user", Content: chats[0]},
				}

				resp, err := aiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
        			Model: "gpt-4",
          			Messages: []openai.ChatCompletionMessage{
            			{Role: "system", Content: "You are a teaching assistant, please answer student questions concisely."},
              			{Role: "user", Content: chats[0]},
            		},
             		MaxTokens: 200,
    			})
				if err != nil {
        			panic(err)
    			}
       			answer := []string{resp.Choices[0].Message.Content}
          		answerWithRole := openai.ChatCompletionMessage{
            		Role: "system",
              		Content: resp.Choices[0].Message.Content,
            	}
          		message = append(message, answerWithRole)
            	marshalledMessage, err := json.Marshal(message)
        		chatData := map[string]interface{}{
					"chat": marshalledMessage,
					"created_at": float64(time.Now().Unix()),
				}
				rdb.HSet(redisCtx, "user:"+UUID.(string)+":ChatSession:"+chats[1], chatData)
				return c.JSON(answer)
			} else {
				log.Println("log1")
				return err
			}

		} else {
			log.Println("log2")
			var messages []openai.ChatCompletionMessage
			history := data["chat"]
			err := json.Unmarshal([]byte(history), &messages)
			if err != nil {
				return err
			}
			message:= openai.ChatCompletionMessage{
				Role: "user",
				Content: chats[0],
			}
			input := append(messages, message)
			resp, err := aiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
        			Model: "gpt-4",
          			Messages: input,
             		MaxTokens: 200,
    			})
			if err != nil {
        			panic(err)
    		}
    		answer := []string{resp.Choices[0].Message.Content}
      		answerWithRole := openai.ChatCompletionMessage{
        		Role: "system",
          		Content: resp.Choices[0].Message.Content,
        	}
      		redisMessage := append(input, answerWithRole)
       		marshalledChat, err:= json.Marshal(redisMessage)
         	if err != nil {
          		return err
          	}
    		chatData := map[string]interface{}{
				"chat": marshalledChat,
				"created_at": float64(time.Now().Unix()),
			}
			rdb.HSet(redisCtx, "user:"+UUID.(string)+":ChatSession:"+chats[1], chatData)
			return c.JSON(answer)
		}

	})


	log.Fatal(app.Listen(":3000"))
}
