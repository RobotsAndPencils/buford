package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/http2"

	"github.com/RobotsAndPencils/buford/push"

	"github.com/dgrijalva/jwt-go"
)

func NewClient() (*http.Client, error) {
	config := &tls.Config{}
	transport := &http.Transport{TLSClientConfig: config}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	return &http.Client{Transport: transport}, nil
}

func main() {
	var deviceToken, filename, keyID, teamID, bundleID string
	var number int

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "k", "", "Path to private signing key")
	flag.StringVar(&keyID, "kid", "", "Key ID")
	flag.StringVar(&teamID, "t", "", "TeamID")
	flag.StringVar(&bundleID, "b", "", "Bundle ID for app")
	flag.IntVar(&number, "n", 100, "Number of notifications to send")
	flag.Parse()

	privateBytes, err := ioutil.ReadFile(filename)
	exitOnError(err)

	block, _ := pem.Decode(privateBytes)
	if block == nil {
		log.Fatal("Key file must be PEM encoded.")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	exitOnError(err)

	client, err := NewClient()
	exitOnError(err)

	service := push.NewService(client, push.Development)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": teamID,
		"iat": time.Now().Unix(),
	})
	token.Header["kid"] = keyID
	log.Printf("%#v\n", token)

	// push the notification:
	tokenString, err := token.SignedString(privateKey)
	exitOnError(err)
	log.Println(tokenString)

	queue := push.NewQueue(service, 20)
	var wg sync.WaitGroup

	// process responses
	// NOTE: Responses may be received in any order.
	go func() {
		count := 1
		for resp := range queue.Responses {
			if resp.Err != nil {
				log.Printf("(%d) device: %s, error: %v", count, resp.DeviceToken, resp.Err)
			} else {
				log.Printf("(%d) device: %s, apns-id: %s", count, resp.DeviceToken, resp.ID)
			}
			count++
			wg.Done()
		}
	}()

	h := &push.Headers{Authorization: tokenString, Topic: bundleID}
	b := []byte(`{"aps":{"alert":"Hello HTTP/2"}}`)

	// synchronous send to prime stream
	id, err := service.Push(deviceToken, h, []byte(`{"aps":{"alert":"Hello HTTP/2"}}`))
	exitOnError(err)
	log.Println("apns-id:", id)

	// concurrent send
	for i := 0; i < number; i++ {
		wg.Add(1)
		queue.Push(deviceToken, h, b)
	}
	// done sending notifications, wait for all responses and shutdown:
	wg.Wait()
	queue.Close()
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
