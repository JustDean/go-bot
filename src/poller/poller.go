package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func RunPoller(ctx context.Context, token string) chan struct{} {
	// Takes messages from tg and pushes to queue
	done := make(chan struct{})
	var msgLimit uint8 = 100
	msgPipe := make(chan []byte, msgLimit)
	tgPollerIsDone := runTgPoller(ctx, msgPipe, token, msgLimit)
	queuePusherIsDone := runQueuePusher(msgPipe)

	go func() {
		<-ctx.Done()
		<-tgPollerIsDone
		close(msgPipe)
		<-queuePusherIsDone
		done <- struct{}{}
	}()

	return done
}

func runTgPoller(ctx context.Context, msgPipe chan []byte, token string, msgLimit uint8) chan struct{} {
	done := make(chan struct{})
	go func() {
		log.Println("Tg poller is starting")
		var offset uint = 0
		for {
			query := fmt.Sprintf("timeout=%d", 30)
			if offset > 0 {
				query = fmt.Sprintf("%s&offset=%d&limit=%d", query, offset, msgLimit)
			}

			url := url.URL{
				Scheme:   "https",
				Host:     "api.telegram.org",
				Path:     fmt.Sprintf("bot%s/getUpdates", token),
				RawQuery: query,
			}
			resp, tgErr := http.Get(url.String())
			if tgErr != nil {
				log.Printf("Error occured getting updates from tg. Error: %s\n", tgErr)
			}
			var data TgUpdateResponseBase

			body, bodyParseError := io.ReadAll(resp.Body)
			if bodyParseError != nil {
				log.Printf("Error occured parsing tg update. Error: %s\n", bodyParseError)
			}
			json.Unmarshal(body, &data)
			if !data.Ok {
				// TODO handle properly
				continue
			}

			updates := data.Updates
			if len(updates) > 0 {
				offset = updates[len(updates)-1].UpdateId + 1
				msgPipe <- body
			}
			resp.Body.Close()

			select {
			case <-ctx.Done():
				{
					log.Println("Poller is shutting down")
					done <- struct{}{}
					return
				}
			default:
			}
		}
	}()
	return done
}

func runQueuePusher(msgPipe chan []byte) chan struct{} {
	done := make(chan struct{})
	go func() {
		conn, connErr := amqp.Dial("amqp://guest:guest@localhost:5672/") // todo in config
		if connErr != nil {
			log.Fatalf("Failed to connect to rabbitmq. Error: %s\n", connErr)
		}

		amqpCh, chErr := conn.Channel()
		if chErr != nil {
			log.Fatalf("Failed to create rabbitmq channel. Error: %s\n", chErr)
		}
		amqpQ, qErr := amqpCh.QueueDeclare(
			"tg_incomming_messages", // name
			false,                   // durable
			false,                   // delete when unused
			false,                   // exclusive
			false,                   // no-wait
			nil,                     // arguments
		)
		if qErr != nil {
			log.Fatalf("Failed to create rabbitmq queue. Error: %s\n", qErr)
		}
		for messages := range msgPipe {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			pushErr := amqpCh.PublishWithContext(ctx,
				"",         // exchange
				amqpQ.Name, // routing key
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        messages,
				})
			cancel()
			if pushErr != nil {
				log.Printf("Failed to push message to rabbitmq. Error: %s\n", pushErr)
			} else {
				log.Println("Successfully pushed to rabbitmq")
			}
		}
		log.Println("Shutting down rabbitmq connection")
		amqpCh.Close()
		conn.Close()
		done <- struct{}{}
	}()
	return done
}
