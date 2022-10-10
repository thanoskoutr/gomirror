package mirrors

import (
	"math"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadMirrorsJSON(t *testing.T) {
	input := "../inputs/in.template.json"
	expected := []Mirror{
		{
			Country:     "Argentina",
			CountryCode: "AR",
			URL:         &url.URL{Scheme: "https", Host: "mirrors.dc.clear.net.ar", Path: "/ubuntu/"},
			Protocol:    443,
		},
		{
			URL:      &url.URL{Scheme: "http", Host: "mirrors.dc.clear.net.ar", Path: "/ubuntu/"},
			Protocol: 80,
		},
		{
			Country:     "Austria",
			CountryCode: "AT",
			URL:         &url.URL{},
			Protocol:    21,
		},
		{
			Country:     "Greece",
			CountryCode: "GR",
			URL:         &url.URL{Scheme: "http", Host: "ftp.cc.uoc.gr", Path: "/mirrors/linux/ubuntu/packages/"},
			Protocol:    80,
		},
		{
			Country:     "France",
			CountryCode: "FR",
			URL:         &url.URL{Scheme: "https", Host: "mirror.ubuntu.ikoula.com", Path: "/"},
			Protocol:    443,
		},
		{
			Country:     "South Africa",
			CountryCode: "ZA",
			URL:         &url.URL{Scheme: "ftp", Host: "mirror.wiru.co.za", Path: "/ubuntu/"},
			Protocol:    21,
		},
	}
	actual := ReadMirrorsJSON(input)
	assert.Equal(t, expected, actual)
}

func TestGetTime(t *testing.T) {
	input := ReadMirrorsJSON("../inputs/in.template.json")
	expected := time.Duration(math.MaxInt64)
	actual := time.Duration(0)

	for _, m := range input {
		actual = m.GetTime()
		if m.URL.Scheme != "http" && m.URL.Scheme != "https" {
			assert.Equal(t, expected, actual)
		}
	}
}
