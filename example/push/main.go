package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"
)

func main() {
	var deviceToken, filename, password, environment string

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file.")
	flag.StringVar(&environment, "e", "development", "Environment")
	flag.Parse()

	cert, err := certificate.Load(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	client, err := push.NewClient(cert)
	if err != nil {
		log.Fatal(err)
	}

	service := push.NewService(client, push.Development, 1)
	if environment == "production" {
		service.Host = push.Production
	}
	defer service.Shutdown()

	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
	}
	b, err := json.Marshal(p)
	if err != nil {
		log.Fatal(b)
	}

	service.Push(deviceToken, nil, b)
	id, deviceToken, err := service.Response()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("device: %v, apns-id %v", deviceToken, id)
}
