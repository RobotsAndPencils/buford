package main

import (
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

	service, err := push.NewService(push.Development, cert)
	if err != nil {
		log.Fatal(err)
	}
	if environment == "production" {
		service.Host = push.Production
	}

	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
	}

	id, err := service.Push(deviceToken, &push.Headers{}, p)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("apns-id:", id)
}
