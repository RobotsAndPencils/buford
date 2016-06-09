# Buford

Apple Push Notification (APN) Provider for Go 1.6 and HTTP/2.

Please see [releases](https://github.com/RobotsAndPencils/buford/releases) for updates.

[![GoDoc](https://godoc.org/github.com/RobotsAndPencils/buford?status.svg)](https://godoc.org/github.com/RobotsAndPencils/buford) [![Build Status](https://travis-ci.org/RobotsAndPencils/buford.svg?branch=ci)](https://travis-ci.org/RobotsAndPencils/buford) ![MIT](https://img.shields.io/badge/license-MIT-blue.svg)

### Documentation

Buford uses Apple's new HTTP/2 Notification API that was announced at WWDC 2015 and [released on December 17, 2015](https://developer.apple.com/news/?id=12172015b).

[API documentation](https://godoc.org/github.com/RobotsAndPencils/buford/) is available from GoDoc.

Also see Apple's [Local and Remote Notification Programming Guide][notification], especially the sections on the JSON [payload][] and the [Notification API][notification-api].

[notification]: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/Introduction.html
[payload]: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/TheNotificationPayload.html#//apple_ref/doc/uid/TP40008194-CH107-SW1
[notification-api]: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/APNsProviderAPI.html#//apple_ref/doc/uid/TP40008194-CH101-SW1

#### Terminology

**APN** Apple Push Notification

**Provider** The Buford library is used to create a _provider_ of push notifications.

**Service** Apple provides the push notification service that Buford communications with.

**Client** An `http.Client` provides an HTTP/2 client to communication with the APN Service.

**Notification** A payload sent to a device token with some headers.

**Device Token** An identifier for an application on a given device.

**Payload** The JSON sent to a device.

**Headers** HTTP/2 headers are used to set priority and expiration.

### Installation

This library requires [Go 1.6.2](https://golang.org/dl/) or better.

Buford depends on several packages outside of the standard library, including the http2 package. Its certificate package depends on the pkcs12 and pushpackage depends on pkcs7. They can be retrieved or updated with:

```
go get -u golang.org/x/net/http2
go get -u golang.org/x/crypto/pkcs12
go get -u github.com/aai/gocrypto/pkcs7
```

I am still looking for feedback on the API so it may change. Please copy Buford and its dependencies into a `vendor/` folder at the root of your project.

### Examples

```go
package main

import (
	"log"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"
)

func main() {
	// set these variables appropriately
	filename := "/path/to/certificate.p12"
	password := ""
	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"

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

	err = service.Push(deviceToken, &push.Headers{}, p)
	if err != nil {
		log.Fatal(err)
	}
	id, _, err := service.Response()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("apns-id:", id)
}
```

#### Headers

You can specify an ID, expiration, priority, and other parameters via the Headers struct.

```go
headers := &push.Headers{
	ID:          "922D9F1F-B82E-B337-EDC9-DB4FC8527676",
	Expiration:  time.Now().Add(time.Hour),
	LowPriority: true,
}

id, err := service.Push(deviceToken, headers, p)
```

If no ID is specified APNS will generate and return a unique ID. When an expiration is specified, APNS will store and retry sending the notification until that time, otherwise APNS will not store or retry the notification. LowPriority should always be set when sending a ContentAvailable payload.

#### Custom values

To add custom values to an APS payload, use the Map method as follows:

```go
p := payload.APS{
	Alert: payload.Alert{Body: "Message received from Bob"},
}
pm := p.Map()
pm["acme2"] = []string{"bang", "whiz"}

id, err := service.Push(deviceToken, nil, pm)
```

The Push method will use json.Marshal to serialize whatever you send it.

#### Resend the same payload

Use json.Marshal to serialize your payload once and then send it to multiple device tokens with PushBytes.

```go
b, err := json.Marshal(p)
if err != nil {
	log.Fatal(err)
}

id, err := service.PushBytes(deviceToken, nil, b)
```

Whether you use Push or PushBytes, the underlying HTTP/2 connection to APNS will be reused.

#### Error responses

Push and PushBytes may return an `error`. It could be an error the JSON encoding or HTTP request, or it could be a `push.Error` which contains the response from Apple. To access the Reason and HTTP Status code, you must convert the `error` to a `push.Error` as follows:

```go
if e, ok := err.(*push.Error); ok {
	switch e.Reason {
	case push.ErrBadDeviceToken:
		// handle error
	}
}
```

### Website Push

Before you can send push notifications through Safari and the Notification Center, you must provide a push package, which is a signed zip file containing some JSON and icons.

Use `pushpackage` to write a zip to a `http.ResponseWriter` or to a file. It will create the `manifest.json` and `signature` files for you.

```go
pkg := pushpackage.New(w)
pkg.EncodeJSON("website.json", website)
pkg.File("icon.iconset/icon_128x128@2x.png", "static/icon_128x128@2x.png")
// other icons... (required)
if err := pkg.Sign(cert, nil); err != nil {
	log.Fatal(err)
}
```

NOTE: The filenames added to the zip may contain forward slashes but not back slashes or drive letters.

See `example/website/` and the [Safari Push Notifications][safari] documentation.

[safari]: https://developer.apple.com/library/mac/documentation/NetworkingInternet/Conceptual/NotificationProgrammingGuideForWebsites/PushNotifications/PushNotifications.html#//apple_ref/doc/uid/TP40013225-CH3-SW12

### Wallet (Passbook) Pass

A pass is a signed zip file with a .pkpass extension and a `application/vnd.apple.pkpass` MIME type. You can use `pushpackage` to write a .pkpass that contains a `pass.json` file.

See `example/wallet/` and the [Wallet Developer Guide][wallet].

[wallet]: https://developer.apple.com/library/prerelease/ios/documentation/UserExperience/Conceptual/PassKit_PG/index.html

### Related Projects

* [apns2](https://github.com/sideshow/apns2) (Go)
* [go-apns-server](https://github.com/CleverTap/go-apns-server) (Go mock server)
* [Push Encryption](https://github.com/GoogleChrome/push-encryption-go) (Go Web Push for Chrome and Firefox)
* [Lowdown](https://github.com/alloy/lowdown) (Ruby)
* [Apnotic](https://github.com/ostinelli/apnotic) (Ruby)
* [Pigeon](https://github.com/codedge-llc/pigeon) (Elixir, iOS and Android)

