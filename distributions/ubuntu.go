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
	UBUNTU_MIRRORS_URL = "https://launchpad.net/ubuntu/+archivemirrors"
)

type Ubuntu struct{}

func (ub Ubuntu) Name() string { return "Ubuntu" }

func (ub Ubuntu) GetMirrors(source mirrors.MirrorSource, filename string) []mirrors.Mirror {
	switch source {
	case mirrors.SourceHTTP:
		// TODO: If error online fallback to internal mirrors
		return FetchUbuntuMirrors(UBUNTU_MIRRORS_URL)
	case mirrors.SourceJSON:
		return mirrors.ReadMirrorsJSON(filename)
	case mirrors.SourceTXT:
		return mirrors.ReadMirrorsTXT(filename)
	default:
		return []mirrors.Mirror{}
	}
}

// TODO: Error handling
func FetchUbuntuMirrors(URL string) []mirrors.Mirror {
	// Definitions
	countries := []string{}
	countryCnt := 0
	mirrorsList := []mirrors.Mirror{}

	// Make request
	resp, err := soup.Get(URL)
	if err != nil {
		log.Fatal("Can not parse Ubuntu mirrors list: ", err)
	}
	// Parse HTML mirror list
	doc := soup.HTMLParse(resp)
	maincontent := doc.Find("div", "id", "maincontent")
	table := maincontent.Find("table", "id", "mirrors_list")
	if table.Pointer == nil {
		log.Fatal("Can not parse Ubuntu mirrors list")
	}
	// Parse table rows
	tbody := table.Find("tbody")
	if tbody.Pointer == nil {
		log.Fatal("Can not parse Ubuntu mirrors list")
	}
	trs := tbody.FindAll("tr")
	for _, tr := range trs {
		// Ignore break sections
		if tr.Attrs()["class"] == "section-break" {
			continue
		}
		// Find country
		if tr.Attrs()["class"] == "head" {
			country := strings.TrimSpace(tr.Find("th").Text())
			countries = append(countries, utils.CorrectCountryName(country))
			countryCnt++
			continue
		}
		// Find mirror protocols
		tds := tr.FindAll("td")
		if len(tds) < 2 {
			continue
		}
		aTags := tds[1].FindAll("a")
		for _, a := range aTags {
			// Find mirror link
			link := a.Attrs()["href"]
			// Parse URL
			urlStr, err := url.Parse(link)
			if err != nil {
				log.Println("Error: Failed to parse URL: ", err)
				urlStr, _ = url.Parse("")
			}
			// Append mirror to slice
			mirrorsList = append(mirrorsList, mirrors.Mirror{
				Country:     countries[countryCnt-1],
				CountryCode: utils.GetCountryCode(countries[countryCnt-1]),
				URL:         urlStr,
			})
		}
	}
	return mirrorsList
}
