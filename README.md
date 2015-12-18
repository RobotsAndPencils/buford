# Buford

Apple Push Notification (APN) Provider for Go 1.6 and HTTP/2.

This is an _alpha_ release with some missing features.

### Installation

To use this library you can install [Go 1.6 beta 1 binaries](https://groups.google.com/forum/#!topic/golang-nuts/24zV9JeBoEE) or [install Go from source](https://golang.org/doc/install/source).

Other than the standard library, Buford depends on the pkcs12 package, which is available at:

```
go get -u golang.org/x/crypto/pkcs12
```

The API is not yet stable. Please use a tool like [Godep](https://github.com/tools/godep) to vendor Buford and it's dependencies.

### Example

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
	filename := "/path/to/certifate.p12"
	password := ""
	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"

	cert, err := certificate.Load(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	service := push.Service{
		Client: push.NewClient(cert),
		Host:   push.Sandbox,
	}

	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
	}

	err = service.Push(deviceToken, push.Headers{}, p)
	if err != nil {
		log.Fatal(err)
	}
}
```

A more complete example can be found in [the example folder](https://github.com/RobotsAndPencils/buford/tree/master/example).
