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

// Notification to send.
type Notification struct {
	DeviceToken string
	Headers     *push.Headers
	Payload     []byte
}

// Response from sending notification.
type Response struct {
	ApnsID       string
	Err          error
	Notification *Notification
}

func worker(service *push.Service, in <-chan Notification, out chan<- Response) {
	for {
		n, more := <-in
		if !more {
			return
		}
		id, err := service.PushBytes(n.DeviceToken, n.Headers, n.Payload)
		out <- Response{ApnsID: id, Err: err, Notification: &n}
	}
}

func main() {
	var deviceToken, filename, password, environment string
	var workers, number int

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file")
	flag.StringVar(&environment, "e", "development", "Environment")
	flag.IntVar(&workers, "w", 20, "Workers to send notifications")
	flag.IntVar(&number, "n", 100, "Number of notifications to send")
	flag.Parse()

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	cert, err := certificate.Load(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	// establish a connection to Apple
	service, err := push.NewService(push.Development, cert)
	if err != nil {
		log.Fatal(err)
	}
	if environment == "production" {
		service.Host = push.Production
	}

	notifications := make(chan Notification)
	responses := make(chan Response)

	// startup workers to send notifications
	for i := 0; i < workers; i++ {
		go worker(service, notifications, responses)
	}

	// wait group to wait for all responses
	var wg sync.WaitGroup

	// process responses
	go func() {
		count := 1
		for {
			resp := <-responses
			device := resp.Notification.DeviceToken
			if resp.Err != nil {
				log.Printf("(%d) device: %s, error: %v", count, device, err)
			} else {
				log.Printf("(%d) device: %s, apns-id: %s", count, device, resp.ApnsID)
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

	n := Notification{
		DeviceToken: deviceToken,
		Headers:     &push.Headers{},
		Payload:     bytes,
	}

	start := time.Now()
	// send notifications
	for i := 0; i < number; i++ {
		wg.Add(1)
		notifications <- n
	}
	sentDuration := time.Since(start)
	close(notifications)

	wg.Wait()
	responseDuration := time.Since(start)

	log.Printf("Time to send %d notifications: %s (%s ea.)", number, sentDuration, sentDuration/time.Duration(number))
	log.Printf("Total time for %d responses: %s (%s ea.)", number, responseDuration, responseDuration/time.Duration(number))
}
