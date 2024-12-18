package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"Skillfactory/Comments/pkg/api"
	"Skillfactory/Comments/pkg/storage/models"
	"Skillfactory/Comments/pkg/storage/postgress"

	kfk "github.com/dontubaby/kafka_wrapper"
	middleware "github.com/dontubaby/mware"
)

func GetCommentsFromService(path string) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:80%s", path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("failed to create request: %v\n", err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body of redirect request: %v\n", err)
		return nil, err
	}
	return body, nil
}

func main() {
	//Подключение к БД комментариев
	commentDB, err := postgress.NewDB("gocomments")
	if err != nil {
		log.Printf("Error DB connection - %v", err)
	}
	defer commentDB.DB.Close()
	//Подключение к БД новостей
	newsDB, err := postgress.NewDB("gonews")
	if err != nil {
		log.Printf("Error DB connection - %v", err)
	}
	defer newsDB.DB.Close()
	//Инициализация консьюмера для чтения входящих сообщений сервиса комментариев
	c, err := kfk.NewConsumer([]string{"localhost:9093"}, "comments_input")
	if err != nil {
		log.Printf("Kafka consumer creating error - %v", err)
	}
	//Инициализация консьюмера для чтения данных необходимых для добавления комментария
	addingCommentsConsumer, err := kfk.NewConsumer([]string{"localhost:9093"}, "add_comments")
	if err != nil {
		log.Printf("Kafka consumer creating error - %v", err)
	}
	//Инициализация продюсера для создания исходящих сообщений сервиса комментариев
	p, err := kfk.NewProducer([]string{"localhost:9093"})
	if err != nil {
		log.Printf("Kafka producer creating error - %v", err)
	}

	ctxKafKa := context.Background()
	//Инициализация API
	api := api.New(newsDB, commentDB)
	//Горутина для получения комментариев
	go func() {
		for {
			log.Println("Start getting messages and redirecting")
			msg, err := c.GetMessages(ctxKafKa)

			if err != nil {
				log.Printf("error when reading message fron Kafka - %v", err)
			}
			if strings.Contains(string(msg.Value), "/comments/") {
				data, err := GetCommentsFromService(string(msg.Value))
				if err != nil {
					log.Printf("error reading data from Kafka message - %v", err)
				}
				if strings.Contains(string(msg.Value), "comments") {
					err = p.SendMessage(ctxKafKa, "comments", data)
					if err != nil {
						log.Printf("error when writing message in Kafka - %v", err)
					}
				}
			}
		}
	}()
	//Горутина для добавления комментариев
	go func() {
		for {
			msg, err := addingCommentsConsumer.GetMessages(ctxKafKa)
			if err != nil {
				log.Printf("error when reading message fron Kafka - %v", err)
			}

			var comment models.Comment

			err = json.Unmarshal(msg.Value, &comment)
			if err != nil {
				log.Println("Failed to decode JSON payload from Kafka")
				return
			}
			err = commentDB.AddComment(comment, newsDB)
			if err != nil {
				log.Printf("error adding comment to service - %v", err)
			}
		}
	}()
	//Инициализация роутера и добавляю к нему middleware
	router := api.Router()
	router.Use(middleware.RequestIDMiddleware, middleware.LoggingMiddleware)

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("cant loading .env file")
	}
	port := os.Getenv("PORT")
	//Запуск сервер
	log.Println("Comments service server start working at port :80!")
	err = http.ListenAndServe(port, router)
	if err != nil {
		log.Println(err)
	}
}
