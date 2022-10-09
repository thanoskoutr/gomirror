package distributions

import (
	"log"
	"net/url"
	"strings"

	"github.com/anaskhan96/soup"
	"github.com/thanoskoutr/gomirror/mirrors"
	"github.com/thanoskoutr/gomirror/utils"
)

const (
	DEBIAN_MIRRORS_URL = "https://www.debian.org/mirror/list"
)

type Debian struct{}

func (deb Debian) Name() string { return "Debian" }

func (deb Debian) GetMirrors(source mirrors.MirrorSource, filename string) []mirrors.Mirror {
	switch source {
	case mirrors.SourceHTTP:
		// TODO: If error online fallback to internal mirrors
		return FetchDebianMirrors(DEBIAN_MIRRORS_URL)
	case mirrors.SourceJSON:
		return mirrors.ReadMirrorsJSON(filename)
	case mirrors.SourceTXT:
		return mirrors.ReadMirrorsTXT(filename)
	default:
		return []mirrors.Mirror{}
	}
}

// TODO: Error handling
func FetchDebianMirrors(URL string) []mirrors.Mirror {
	// Definitions
	countries := []string{}
	countryCnt := 0
	mirrorsList := []mirrors.Mirror{}

	// Make request
	resp, err := soup.Get(URL)
	if err != nil {
		log.Fatal("Can not parse Debian mirrors list: ", err)
	}
	// Parse HTML mirror list
	doc := soup.HTMLParse(resp)
	content := doc.Find("div", "id", "content")
	tables := content.FindAll("table")
	if len(tables) < 3 {
		log.Fatal("Can not parse Debian mirrors list")
	}
	// Parse table rows
	trs := tables[1].Find("tbody").FindAll("tr")
	for _, tr := range trs {
		tds := tr.FindAll("td")
		// Find Architectures
		archs := ""
		if len(tds) >= 3 {
			archs = strings.TrimSpace(tds[2].Text())
		}
		// Parse table headers
		for _, td := range tds {
			// Find country
			big := td.Find("big")
			if big.Pointer != nil {
				country := strings.TrimSpace(big.Find("strong").Text())
				countries = append(countries, country)
				countryCnt++
			}
			// Find mirror link
			a := td.Find("a")
			if a.Pointer != nil {
				link := a.Attrs()["href"]
				// Parse URL
				urlStr, err := url.Parse(link)
				if err != nil {
					log.Println("Error: Failed to parse URL:", err)
					urlStr, _ = url.Parse("")
				}
				// Append mirror to slice
				mirrorsList = append(mirrorsList, mirrors.Mirror{
					Country:       countries[countryCnt-1],
					CountryCode:   utils.GetCountryCode(countries[countryCnt-1]),
					URL:           urlStr,
					Architectures: archs,
				})
			}
		}
	}
	return mirrorsList
}
