package distributions

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/thanoskoutr/gomirror/mirrors"
	"github.com/thanoskoutr/gomirror/utils"
)

const (
	ARCH_MIRRORS_URL = "https://archlinux.org/mirrors/status/json/"
)

type Arch struct{}

func (arch Arch) Name() string { return "Arch" }

func (arch Arch) GetMirrors(source mirrors.MirrorSource, filename string) []mirrors.Mirror {
	switch source {
	case mirrors.SourceHTTP:
		// TODO: If error online fallback to internal mirrors
		return FetchArchMirrors(ARCH_MIRRORS_URL)
	case mirrors.SourceJSON:
		return mirrors.ReadMirrorsJSON(filename)
	case mirrors.SourceTXT:
		return mirrors.ReadMirrorsTXT(filename)
	default:
		return []mirrors.Mirror{}
	}
}

// TODO: Error handling
// TODO: Add error checking for JSON fields
// TODO: Take note of other JSON attributes for mirrors: delay, score, active
func FetchArchMirrors(URL string) []mirrors.Mirror {
	// Make request
	resp, err := utils.GetRequest(URL)
	if err != nil {
		log.Fatal("Can not parse Arch mirrors list: ", err)
	}
	// Parse response as JSON
	respJson := map[string]interface{}{}
	err = json.Unmarshal(resp, &respJson)
	if err != nil {
		log.Fatal("Can not parse Arch mirrors list: ", err)
	}
	// Get URLs array
	urls := respJson["urls"].([]interface{})
	// Create Mirrors array
	mirrorsList := make([]mirrors.Mirror, len(urls))
	// Convert URLs object to Mirrors
	for i, u := range urls {
		urlMap := u.(map[string]interface{})
		// Parse URL
		urlStr, err := url.Parse(urlMap["url"].(string))
		if err != nil {
			log.Println("Error: Failed to parse URL: ", err)
		}
		mirrorsList[i] = mirrors.Mirror{
			Country:     urlMap["country"].(string),
			CountryCode: urlMap["country_code"].(string),
			URL:         urlStr,
		}
	}
	return mirrorsList
}
