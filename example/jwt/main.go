package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/RobotsAndPencils/buford/push"

	"github.com/dgrijalva/jwt-go"
)

func main() {
	var deviceToken, filename, keyID, teamID, bundleID string

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "k", "", "Path to private signing key")
	flag.StringVar(&keyID, "kid", "", "Key ID")
	flag.StringVar(&teamID, "t", "", "TeamID")
	flag.StringVar(&bundleID, "b", "", "Bundle ID for app")
	flag.Parse()

	privateBytes, err := ioutil.ReadFile(filename)
	exitOnError(err)

	block, _ := pem.Decode(privateBytes)
	if block == nil {
		log.Fatal("Key file must be PEM encoded.")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	exitOnError(err)

	service := push.NewService(http.DefaultClient, push.Development)

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

	h := &push.Headers{Authorization: tokenString, Topic: bundleID}

	id, err := service.Push(deviceToken, h, []byte(`{"aps":{"alert":"Hello HTTP/2"}}`))
	exitOnError(err)

	fmt.Println("apns-id:", id)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
