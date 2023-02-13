package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/telenornms/svipul"
)

var sleeptime = flag.Duration("sleep", -time.Second, "sleep between iterations, negative value means only one execution")
var expire = flag.Duration("ttl", 10*time.Second, "expiry time. Minimum: 1ms")

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		tpoll.Fatalf("failed to connect to rabbitMQ: %s", err)
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		tpoll.Fatalf("failed to connect to open a channel: %s", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"tpoll", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		tpoll.Fatalf("failed to declare a queue: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	flag.Parse()
	if expire.Milliseconds() < 1 {
		tpoll.Fatalf("TTL must be at least 1ms")
	}
	var bs [][]byte
	args := flag.Args()
	if len(os.Args) < 1 {
		tpoll.Fatalf("no order-file supplied")
	}
	for _, fil := range args {
		b, err := os.ReadFile(fil)
		if err != nil {
			tpoll.Fatalf("failed to read %s", fil)
		}
		bs = append(bs, b)
	}
	ttl := fmt.Sprintf("%d", expire.Milliseconds())
	tpoll.Debugf("expire: %s", ttl)
	for {
		for _, b := range bs {
			err = ch.PublishWithContext(ctx,
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/json",
					Expiration:  ttl,
					Body:        []byte(b),
				})
			if err != nil {
				tpoll.Fatalf("failed to publish a message: %s", err)
			}
			tpoll.Logf("Sent %d bytes", len(b))
		}
		if *sleeptime < 0 {
			tpoll.Logf("negative sleeptime, exiting after 1 publish")
			os.Exit(0)
		}
		tpoll.Logf("Sleeping %s", sleeptime)
		time.Sleep(*sleeptime)
	}
}
