package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/RobotsAndPencils/buford/pushpackage"
)

var (
	website = pushpackage.Website{
		Name:                "Buford",
		PushID:              "web.com.github.RobotsAndPencils.buford",
		AllowedDomains:      []string{"https://c73445e5.ngrok.io"},
		URLFormatString:     `http://c73445e5.ngrok.io/%@/?q=%@`,
		AuthenticationToken: "19f8d7a6e9fb8a7f6d9330dabe",
		WebServiceURL:       "https://c73445e5.ngrok.io",
	}

	templates = template.Must(template.ParseFiles("index.html"))
)

func requestPermission(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", website)
}

func pushPackages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")

	img, err := os.Open("../../pushpackage/fixtures/gopher.png")
	if err != nil {
		log.Fatal(err)
	}
	defer img.Close()

	iconset := pushpackage.IconSet{
		{
			Name:   "icon_128x128@2x.png",
			Reader: img,
		},
	}

	err = pushpackage.New(w, &website, iconset)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", requestPermission)
	http.HandleFunc("/v1/pushPackages/"+website.PushID, pushPackages)
	http.ListenAndServe(":5000", nil)
}
