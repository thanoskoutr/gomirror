package mirrors

import (
	"net/url"
	"testing"

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
	}
	actual := ReadMirrorsJSON(input)
	assert.Equal(t, expected, actual)
}
