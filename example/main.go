package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/RobotsAndPencils/buford"
)

func main() {
	var deviceToken, filename, password string

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file.")
	flag.Parse()

	cert, err := buford.LoadCert(filename, password)

	// Setup HTTPS client
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	transport := &http.Transport{TLSClientConfig: config}
	client := &http.Client{Transport: transport}

	json := bytes.NewBufferString(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)
	u := fmt.Sprintf("https://api.sandbox.push.apple.com/3/device/%v", deviceToken)

	resp, err := client.Post(u, "application/json", json)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	log.Println("status", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(body))
}
