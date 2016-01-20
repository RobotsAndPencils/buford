package pushpackage_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/RobotsAndPencils/buford/pushpackage"
)

func TestNew(t *testing.T) {
	website := &pushpackage.Website{
		Name:                "Bay Airlines",
		PushID:              "web.com.example.domain",
		AllowedDomains:      []string{"http://domain.example.com"},
		URLFormatString:     "http://domain.example.com/%@/?flight=%@",
		AuthenticationToken: "19f8d7a6e9fb8a7f6d9330dabe",
		WebServiceURL:       "https://example.com/push",
	}

	r, err := os.Open("fixtures/gopher.png")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	iconset := pushpackage.IconSet{
		{
			Name:   "icon_128x128@2x.png",
			Reader: r,
		},
	}

	buf := new(bytes.Buffer)
	err = pushpackage.New(buf, website, iconset)
	if err != nil {
		t.Fatal(err)
	}
}
