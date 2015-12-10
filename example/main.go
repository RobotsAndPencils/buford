package main

import (
	"flag"
	"log"

	"github.com/RobotsAndPencils/buford"
)

func main() {
	var deviceToken, filename, password string

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file.")
	flag.Parse()

	cert, err := buford.LoadCert(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	service := buford.Service{
		Client: buford.NewClient(cert),
		Host:   buford.Sandbox,
	}

	err = service.Push(deviceToken, buford.Headers{}, []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`))
	if err != nil {
		log.Fatal(err)
	}
}
