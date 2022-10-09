package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pariz/gountries"
)

// Generic function that creates an HTTP GET request and returns the body.
func GetRequest(URL string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyText, nil
}

// Uses geo-location API services to locate the user from its external IP
func GetCountry() string {
	geolocationAPIs := map[string]string{
		"https://tnedi.me/json":               "country",
		"https://ident.me/json":               "country",
		"https://get.geojs.io/v1/ip/geo.json": "country",
		"https://freeipapi.com/api/json":      "countryName",
		"https://ipapi.co/json":               "country_name",
	}
	// If at least one API service returns a valid response return early
	for geolocationAPI, countryKey := range geolocationAPIs {
		// Send request to geolocation API
		resp, err := GetRequest(geolocationAPI)
		if err != nil {
			log.Printf("Error: Can not fetch geolocation info from %v: %v", geolocationAPI, err)
			// Retry with another service
			continue
		}
		respJson := map[string]interface{}{}
		err = json.Unmarshal(resp, &respJson)
		if err != nil {
			log.Printf("Error: Can not fetch geolocation info from %v: %v", geolocationAPI, err)
			log.Printf("Error: Invalid response from %v: %v", geolocationAPI, string(resp))
			// Retry with another service
			continue
		}
		// Get Country from response
		if respJson[countryKey] == nil {
			log.Printf("Error: Can not fetch geolocation info from %v: %v", geolocationAPI, err)
			log.Printf("Error: Invalid response from %v: %v", geolocationAPI, string(resp))
			// Retry with another service
			continue
		}
		country := strings.TrimSpace(respJson[countryKey].(string))
		return CorrectCountryName(country)
	}
	log.Fatal("Can not find country")
	return ""
}

// Returns the 2 letter country code of the given country.
func GetCountryCode(country string) string {
	query := gountries.New()
	countryCode, _ := query.FindCountryByName(country)
	return countryCode.Alpha2
}

// Returns the correct country for name for some country alternative naming.
func CorrectCountryName(country string) string {
	switch country {
	case "Iran, Islamic Republic of":
		return "Iran"
	case "Korea, Republic of":
		return "South Korea"
	case "Korea":
		return "South Korea"
	case "Macedonia, Republic of":
		return "Macedonia"
	case "Moldova, Republic of":
		return "Moldova"
	case "Tanzania, United Republic of":
		return "Tanzania"
	case "Viet Nam":
		return "Vietnam"
	default:
		return country
	}
}
