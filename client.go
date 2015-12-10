package buford

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Client error responses
var (
	ErrBadDeviceToken = errors.New("bad device token")
)

// NewClient sets up an HTTPS client
func NewClient(cert tls.Certificate) *http.Client {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: config}

	return &http.Client{Transport: transport}
}

type response struct {
	Reason string `json:"reason"`
	// timestamp, other fields?
}

// Push notification
func Push(client *http.Client, gateway string, deviceToken string, payload []byte) error {
	urlStr := fmt.Sprintf("https://%v/3/device/%v", gateway, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apns-expiration", "0")
	// apns-id, other headers such as priority

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// read entire response body
	// TODO: could decode while reading instead if not logging body too
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// logging full responses while learning API
	log.Println(string(body))

	var response response
	json.Unmarshal(body, &response)

	switch response.Reason {
	case "BadDeviceToken":
		return ErrBadDeviceToken
	}
	return fmt.Errorf("Error response: %v", response.Reason)
}
