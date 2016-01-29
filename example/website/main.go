package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/RobotsAndPencils/buford/pushpackage"
	"golang.org/x/crypto/pkcs12"
)

var (
	website = pushpackage.Website{
		Name:                "Buford",
		PushID:              "web.com.github.RobotsAndPencils.buford",
		AllowedDomains:      []string{"https://59e9995b.ngrok.io"},
		URLFormatString:     `https://59e9995b.ngrok.io/%@/?q=%@`,
		AuthenticationToken: "19f8d7a6e9fb8a7f6d9330dabe",
		WebServiceURL:       "https://59e9995b.ngrok.io",
	}

	templates = template.Must(template.ParseFiles("index.html"))
)

func requestPermissionHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", website)
}

func MustOpen(name string) *os.File {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func pushPackagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")

	const iconPath = "../../pushpackage/fixtures/"

	icon128x := MustOpen(filepath.Join(iconPath, "gopher.png"))
	defer icon128x.Close()
	icon128 := MustOpen(filepath.Join(iconPath, "gopher.png"))
	defer icon128.Close()
	icon32x := MustOpen(filepath.Join(iconPath, "gopher.png"))
	defer icon32x.Close()
	icon32 := MustOpen(filepath.Join(iconPath, "gopher.png"))
	defer icon32.Close()
	icon16x := MustOpen(filepath.Join(iconPath, "gopher.png"))
	defer icon16x.Close()
	icon16 := MustOpen(filepath.Join(iconPath, "gopher.png"))
	defer icon16.Close()

	iconset := pushpackage.IconSet{
		{Name: "icon_128x128@2x.png", Reader: icon128x},
		{Name: "icon_128x128.png", Reader: icon128},
		{Name: "icon_32x32@2x.png", Reader: icon32x},
		{Name: "icon_32x32.png", Reader: icon32},
		{Name: "icon_16x16@2x.png", Reader: icon16x},
		{Name: "icon_16x16.png", Reader: icon16},
	}

	if err := pushpackage.New(w, &website, iconset, Cert, PrivateKey); err != nil {
		log.Fatal(err)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	var logs struct {
		Logs []string `json:"logs"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logs); err == io.EOF {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	for _, msg := range logs.Logs {
		log.Println(msg)
	}
}

var (
	Cert       *x509.Certificate
	PrivateKey *rsa.PrivateKey
)

func main() {
	p12, err := ioutil.ReadFile("../../cert-website.p12")
	if err != nil {
		log.Fatal(err)
	}

	key, cert, err := pkcs12.Decode(p12, "")
	if err != nil {
		log.Fatal(err)
	}
	Cert = cert
	PrivateKey = key.(*rsa.PrivateKey)

	http.HandleFunc("/", requestPermissionHandler)
	http.HandleFunc("/v1/pushPackages/"+website.PushID, pushPackagesHandler)
	http.HandleFunc("/v1/log", logHandler)
	http.ListenAndServe(":5000", nil)
}
