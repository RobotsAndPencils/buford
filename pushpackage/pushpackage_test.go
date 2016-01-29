package pushpackage_test

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/RobotsAndPencils/buford/pushpackage"
	"golang.org/x/crypto/pkcs12"
)

func TestNew(t *testing.T) {
	website := &pushpackage.Website{
		Name:                "Bay Airlines",
		PushID:              "web.com.example.domain",
		AllowedDomains:      []string{"http://domain.example.com"},
		URLFormatString:     `http://domain.example.com/%@/?flight=%@`,
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

	p12, err := ioutil.ReadFile("../cert-website.p12")
	if err != nil {
		log.Fatal(err)
	}

	privateKey, cert, err := pkcs12.Decode(p12, "")
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err = pushpackage.New(buf, website, iconset, cert, privateKey.(*rsa.PrivateKey))
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"website.json":  `{"websiteName":"Bay Airlines","websitePushID":"web.com.example.domain","allowedDomains":["http://domain.example.com"],"urlFormatString":"http://domain.example.com/%@/?flight=%@","authenticationToken":"19f8d7a6e9fb8a7f6d9330dabe","webServiceURL":"https://example.com/push"}`,
		"manifest.json": `{"icon.iconset/icon_128x128@2x.png":"5d31b7d2ea66ec7087c3789b2c6ca2aad67e459c","website.json":"8225d6cdd71f00888ff576aaab8d7ec4a27553c7"}`,
	}

	z, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	for _, f := range z.File {
		if exp, ok := expected[f.Name]; ok {
			b, err := zipReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != exp {
				t.Errorf("Unexpected content for %s: %s", f.Name, b)
			}
		} else {
			log.Println(f.Name)
		}
	}
}

func zipReadFile(f *zip.File) ([]byte, error) {
	zf, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer zf.Close()
	return ioutil.ReadAll(zf)
}
