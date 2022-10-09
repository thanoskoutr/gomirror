package utils

import "testing"

func TestGetCountry(t *testing.T) {
	expectedCountry := "Greece"
	country := GetCountry()
	if country != expectedCountry {
		t.Fatalf("Expected: %v, Got: %v", expectedCountry, country)
	}
}
