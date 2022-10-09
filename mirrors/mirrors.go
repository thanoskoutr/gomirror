package mirrors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/thanoskoutr/gomirror/utils"
)

const (
	HTTP_TIMEOUT = 7
)

type Mirror struct {
	// TODO: Make it enum or Locale
	Country     string   `json:"country,omitempty"`
	CountryCode string   `json:"country_code,omitempty"`
	URL         *url.URL `json:"url"`
	// TODO: Make it enum
	Protocol      Protocol          `json:"protocol"`
	Architectures string            `json:"architectures,omitempty"`
	Statistics    *MirrorStatistics `json:"statistics,omitempty"`
}

func (m Mirror) String() string {
	if len(m.Architectures) != 0 {
		return fmt.Sprintf("{Country: %v, CountryCode: %v, URL: %v, Statistics: %v, Architectures: %v}", m.Country, m.CountryCode, m.URL, m.Statistics, m.Architectures)
	}
	return fmt.Sprintf("{Country: %v, CountryCode: %v, URL: %v, Statistics: %v}", m.Country, m.CountryCode, m.URL, m.Statistics)
}

func (m *Mirror) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Country       string            `json:"country,omitempty"`
		CountryCode   string            `json:"country_code,omitempty"`
		URL           string            `json:"url"`
		Protocol      string            `json:"protocol"`
		Architectures string            `json:"architectures,omitempty"`
		Statistics    *MirrorStatistics `json:"statistics,omitempty"`
	}{
		Country:       m.Country,
		CountryCode:   m.CountryCode,
		URL:           m.URL.String(),
		Protocol:      m.URL.Scheme,
		Architectures: m.Architectures,
		Statistics:    m.Statistics,
	})
}

func (m *Mirror) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	var err error
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v["country"] != nil {
		m.Country = v["country"].(string)
	}
	if v["country_code"] != nil {
		m.CountryCode = v["country_code"].(string)
	} else {
		m.CountryCode = utils.GetCountryCode(m.Country)
	}
	if v["url"] != nil {
		m.URL, err = url.Parse(v["url"].(string))
		if err != nil {
			return err
		}
	} else {
		m.URL, _ = url.Parse("")
	}
	if v["protocol"] != nil {
		m.Protocol, err = ToProtocol(v["protocol"].(string))
		if err != nil {
			return err
		}
	} else {
		m.Protocol, err = ToProtocol(m.URL.Scheme)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Error handling
func (m Mirror) GetStatus() int {
	client := &http.Client{}
	req, err := http.NewRequest("GET", m.URL.String(), nil)
	if err != nil {
		log.Fatal("Can not get HTTP status of mirror: ", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Can not get HTTP status of mirror: ", err)
	}
	return resp.StatusCode
}

// TODO: Clean caches before request (DNS)
// TODO: Check for server delays (on multiple requests)
// TODO: Error handling
// TODO: Handle other protocols
func (m Mirror) GetTime() time.Duration {
	// Unsupported Protocols
	if m.URL.Scheme != "http" && m.URL.Scheme != "https" && m.URL.Scheme != "ftp" {
		log.Println("Error: Unsupported Protocol: ", m.URL.Scheme)
		return math.MaxInt64
	}
	// Create new request
	req, err := http.NewRequest("GET", m.URL.String(), nil)
	if err != nil {
		log.Fatal("Error: Can not create request: ", err)
	}
	// Initialize an HTTP tracer
	trace := &httptrace.ClientTrace{}
	// Wrap request with the HTTP tracer context
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	// Make request and keep time
	client := &http.Client{
		Timeout: HTTP_TIMEOUT * time.Second,
	}
	start := time.Now()
	_, err = client.Do(req)
	if err != nil {
		log.Println("Error: Can not fetch mirror: ", err)
		return time.Since(start)
	}
	return time.Since(start)
}

// TODO: Decide on manual approach or use library
// TODO: Using ping library will require sudo privileges to run program
func (m Mirror) Ping() time.Duration {
	pinger, err := ping.NewPinger(m.URL.Host)
	if err != nil {
		log.Println("Error: Can not create pinger: ", err)
		return 0
	}
	pinger.Count = 2
	pinger.Timeout = 5 * time.Second
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		log.Println("Error: Can not ping target host: ", err)
	}
	// TODO: Analyze statistics more: https://pkg.go.dev/github.com/go-ping/ping#Statistics
	return pinger.Statistics().AvgRtt
}

// TODO: Add more fields (httptrace)
type MirrorStatistics struct {
	ResponseTimeHTTP    time.Duration
	ResponseTimePing    time.Duration // AvgRtt
	AvgResponseTimeHTTP time.Duration
	Speed               int // in MB/s
}

func (s MirrorStatistics) String() string {
	return fmt.Sprintf("{ResponseTimeHTTP: %v, AvgResponseTimeHTTP: %v}", s.ResponseTimeHTTP, s.AvgResponseTimeHTTP)
}

// TODO: Handle field types for Nanoseconds (int32, float32, string)
func (s *MirrorStatistics) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ResponseTimeHTTP    int `json:"http_response,omitempty"`
		ResponseTimePing    int `json:"ping_response,omitempty"`
		AvgResponseTimeHTTP int `json:"avg_http_response,omitempty"`
		Speed               int `json:"speed,omitempty"`
	}{
		ResponseTimeHTTP:    int(s.ResponseTimeHTTP.Nanoseconds()),
		ResponseTimePing:    int(s.ResponseTimePing.Nanoseconds()),
		AvgResponseTimeHTTP: int(s.AvgResponseTimeHTTP.Nanoseconds()),
		Speed:               s.Speed,
	})
}

// Examples: HTTP (80), HTTPS (443), FTP (21), Rsync (22)
type Protocol uint16

const (
	ProtoHTTP  Protocol = 80
	ProtoHTTPS Protocol = 443
	ProtoFTP   Protocol = 21
	ProtoRSYNC Protocol = 22
)

func (p Protocol) String() string {
	switch p {
	case ProtoHTTP:
		return "http"
	case ProtoHTTPS:
		return "https"
	case ProtoFTP:
		return "ftp"
	case ProtoRSYNC:
		return "rsync"
	default:
		return fmt.Sprintf("%d", p)
	}
}

func ToProtocol(proto string) (Protocol, error) {
	switch strings.ToLower(proto) {
	case "http":
		return ProtoHTTP, nil
	case "https":
		return ProtoHTTPS, nil
	case "ftp":
		return ProtoFTP, nil
	case "rsync":
		return ProtoRSYNC, nil
	default:
		return 0, fmt.Errorf("unsupported Protocol: %v", proto)
	}
}

// Supported sources: TXT, JSON, HTTP
type MirrorSource int

const (
	SourceHTTP MirrorSource = iota
	SourceJSON
	SourceTXT
)

func (s MirrorSource) String() string {
	switch s {
	case SourceHTTP:
		return "http"
	case SourceJSON:
		return "json"
	case SourceTXT:
		return "txt"
	default:
		return fmt.Sprintf("%d", s)
	}
}

func ToMirrorSource(source string) (MirrorSource, error) {
	switch strings.ToLower(source) {
	case "http":
		return SourceHTTP, nil
	case "json":
		return SourceJSON, nil
	case "txt":
		return SourceTXT, nil
	default:
		return -1, fmt.Errorf("unsupported source type: %v", source)
	}
}

// Read URL mirrors from JSON file
// TODO: File location
func ReadMirrorsJSON(filename string) []Mirror {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Can not read mirrors: ", err)
	}
	defer file.Close()
	// Read whole file
	mirrorsBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Can not read mirrors: ", err)
	}
	// Parse file as JSON
	var mirrorsJson map[string]interface{}
	err = json.Unmarshal(mirrorsBytes, &mirrorsJson)
	if err != nil {
		log.Fatal("Can not parse mirrors list: ", err)
	}
	// Get URLs array
	if mirrorsJson["urls"] == nil {
		log.Fatalf("Can not parse mirrors list: %v", mirrorsJson)
	}
	urls := mirrorsJson["urls"].([]interface{})
	// Create Mirrors array
	mirrorsList := make([]Mirror, len(urls))
	// Convert URLs object back to JSON
	urlsMarshaled, err := json.Marshal(urls)
	if err != nil {
		log.Fatal("Can not parse mirrors url list: ", err)
	}
	// Parse URLs as Mirrors array
	err = json.Unmarshal(urlsMarshaled, &mirrorsList)
	if err != nil {
		log.Fatal("Can not parse mirrors list: ", err)
	}
	return mirrorsList
}

// Read URL mirrors from TXT file
// TODO: File location
func ReadMirrorsTXT(filename string) []Mirror {
	// Definitions
	urls := []string{}
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Can not read mirrors: ", err)
	}
	defer file.Close()
	// Read file line-by-line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(scanner.Text()) != 0 {
			urls = append(urls, strings.TrimSpace(scanner.Text()))
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Can not read mirrors:", err)
	}

	// Create Mirrors array
	mirrors := make([]Mirror, len(urls))
	// Convert URLs object to Mirrors
	for i, u := range urls {
		// Parse URL
		urlStr, err := url.Parse(u)
		if err != nil {
			log.Println("Error: Failed to parse URL: ", err)
			urlStr, _ = url.Parse("")
		}
		mirrors[i] = Mirror{
			URL: urlStr,
		}
	}
	return mirrors
}
