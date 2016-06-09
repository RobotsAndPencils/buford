package main

import (
	"encoding/json"
	"flag"
	"log"
	"sync"
	"time"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/push"
)

func main() {
	var deviceToken, filename, password, environment string
	var workers uint
	var number int

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file")
	flag.StringVar(&environment, "e", "development", "Environment")
	flag.UintVar(&workers, "w", 20, "Workers to send notifications")
	flag.IntVar(&number, "n", 100, "Number of notifications to send")
	flag.Parse()

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	cert, err := certificate.Load(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	// establish a connection to Apple
	client, err := push.NewClient(cert)
	if err != nil {
		log.Fatal(err)
	}

	service := push.NewService(client, push.Development, workers)
	if environment == "production" {
		service.Host = push.Production
	}

	// wait group to wait for all responses
	var wg sync.WaitGroup

	// process responses
	go func() {
		count := 1
		for {
			id, device, err := service.Response()
			if err != nil {
				log.Printf("(%d) device: %s, error: %v", count, device, err)
			} else {
				log.Printf("(%d) device: %s, apns-id: %s", count, device, id)
			}
			count++
			wg.Done()
		}
	}()

	// prepare notification(s) to send
	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
	}
	bytes, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	// send notifications
	for i := 0; i < number; i++ {
		wg.Add(1)
		service.Push(deviceToken, nil, bytes)
	}
	service.Shutdown()
	wg.Wait()
	elapsed := time.Since(start)

	log.Printf("Time for %d responses: %s (%s ea.)", number, elapsed, elapsed/time.Duration(number))
}
