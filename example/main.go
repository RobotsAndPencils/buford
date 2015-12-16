package main

import (
	"flag"
	"log"

	"github.com/RobotsAndPencils/buford"
	"github.com/RobotsAndPencils/buford/cert"
)

func main() {
	var deviceToken, filename, password, environment string

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file.")
	flag.StringVar(&environment, "e", "development", "Environment")
	flag.Parse()

	cert, err := cert.Load(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	service := buford.Service{
		Client: buford.NewClient(cert),
		Host:   buford.Sandbox,
	}
	if environment == "production" {
		service.Host = buford.Live
	}

	err = service.Push(deviceToken, buford.Headers{}, []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`))
	if err != nil {
		log.Fatal(err)
	}
}
