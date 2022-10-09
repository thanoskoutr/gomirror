package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/thanoskoutr/gomirror/distributions"
	"github.com/thanoskoutr/gomirror/mirrors"
	"github.com/thanoskoutr/gomirror/utils"
)

const (
	STATISTICS_ROUNDS = 1
)

func main() {
	// ------------------------------------ CONFIGURATION
	// Initialize
	var distroMirrors *mirrors.DistributionMirrors = &mirrors.DistributionMirrors{}
	var country string
	var countryCode string
	var mirrorSourceType mirrors.MirrorSource
	var mirrorSourceFile string

	// TODO: Add cobra for more configurable CLI
	// TODO: Distro should be required argument
	// TODO: Add flag for displaying configuration options
	// TODO: Add flag for writing to file
	// Supported Flags
	var (
		distro       = flag.String("distro", "", "The distribution to rank mirrors. Supported: \"Ubuntu\", \"Debian\", \"Arch\"")
		mode         = flag.String("mode", "rank", "Mode of operation. Supported: \"rank\": Rank all mirrors based on your location, \"best\": Find best mirror based on your location")
		sourceType   = flag.String("source", "http", "The type of source for mirror list. Supported: \"http\", \"json\", \"txt\"")
		sourceFile   = flag.String("file", "", "The file with the mirrors. Valid only for \"json\", \"txt\" source")
		countryInput = flag.String("country", "", "The country where the system is located")
		output       = flag.String("output", "stdout", "The output format for the results. Supported: \"stdout\", \"json\", \"txt\", \"csv\"")
	)
	// Parse Flags
	flag.Parse()
	fmt.Fprintf(os.Stderr, "Configuration Options:\n")
	fmt.Fprintf(os.Stderr, "----------------------\n")

	// Validate Distribution
	if len(*distro) == 0 {
		fmt.Fprintf(os.Stderr, "Select a distribution to start\n")
		os.Exit(1)
	}
	switch *distro {
	case "Ubuntu":
		distroMirrors.Distribution = distributions.Ubuntu{}
	case "Debian":
		distroMirrors.Distribution = distributions.Debian{}
	case "Arch":
		distroMirrors.Distribution = distributions.Arch{}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported distribution: %v\n", *distro)
		os.Exit(1)
	}

	// Validate Country
	if len(*countryInput) == 0 {
		country = utils.GetCountry()
	} else {
		country = *countryInput
	}
	countryCode = utils.GetCountryCode(country)
	fmt.Fprintf(os.Stderr, "Country: %v\n", country)
	fmt.Fprintf(os.Stderr, "Country Code: %v\n", countryCode)

	// Validate Mirrors Source Type
	mirrorSourceType, err := mirrors.ToMirrorSource(*sourceType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unsupported source type: %v\n", *sourceType)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Mirror Source Type: %v\n", mirrorSourceType)

	// Validate Mirrors Source file
	mirrorSourceFile = *sourceFile
	if len(mirrorSourceFile) != 0 && mirrorSourceType == mirrors.SourceHTTP {
		fmt.Fprintf(os.Stderr, "No valid source type specified for source file: %v\n", mirrorSourceFile)
		os.Exit(1)
	}
	if len(mirrorSourceFile) == 0 && mirrorSourceType != mirrors.SourceHTTP {
		fmt.Fprintf(os.Stderr, "No Mirror file specified\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Mirror Source File: %v\n", mirrorSourceFile)

	// Validate Mode
	switch *mode {
	case "rank":
	case "best":
		fmt.Fprintf(os.Stderr, "Operation Mode: %v\n", *mode)

	default:
		fmt.Fprintf(os.Stderr, "Unsupported mode: %v\n", *mode)
		os.Exit(1)
	}

	// Validate Output
	switch *output {
	case "stdout":
	case "json":
	case "csv":
		fmt.Fprintf(os.Stderr, "Output format: %v\n", *mode)
	case "txt":
	default:
		fmt.Fprintf(os.Stderr, "Unsupported output format: %v\n", *output)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "----------------------\n")

	// ------------------------------------ OPERATIONS
	// Add Mirrors manually (for testing)
	distroMirrors.Mirrors = []*mirrors.Mirror{
		{
			Country:     "Armenia",
			CountryCode: "AM",
			URL:         &url.URL{Scheme: "http", Host: "ftp.am.debian.org", Path: "debian/"},
			Protocol:    21,
		},
		{
			Country:     "Australia",
			CountryCode: "AU",
			URL:         &url.URL{Scheme: "http", Host: "ftp.au.debian.org", Path: "debian/"},
			Protocol:    21,
		},
	}
	// Update Mirrors
	distroMirrors.UpdateMirrors(mirrorSourceType, mirrorSourceFile)

	// DEBUG: Print Mirrors
	// fmt.Printf("Mirrors: %v\n", distroMirrors)
	// fmt.Printf("Mirrors Len: %v\n", len(distroMirrors.Mirrors))
	// DEBUG: Print Mirrors JSON
	// distroMirrorsJson, _ := json.Marshal(distroMirrors)
	// fmt.Println(string(distroMirrorsJson))

	// Update Statistics
	distroMirrors.UpdateMirrorStatistics(STATISTICS_ROUNDS)
	// DEBUG: Print Mirrors
	// fmt.Printf("Mirrors (Updated): %v\n", distroMirrors)

	// Operate based on selected mode
	switch *mode {
	case "rank":
		// Sort Mirrors
		distroMirrors.SortMirrors()
		fmt.Fprintf(os.Stderr, "Ranked Mirrors:\n")
	case "best":
		// Print best Mirror (based on HTTP response)
		bestMirror := distroMirrors.BestMirror()
		fmt.Fprintf(os.Stderr, "Best Mirror (relative):\n")
		distroMirrors.Mirrors = []*mirrors.Mirror{&bestMirror}
	}

	// Output results
	// TODO: Make them methods of Distribution Mirrors
	// TODO: Create common output for all formats
	switch *output {
	case "stdout":
		// TODO: Align whitespaces
		fmt.Printf("%v %v %v %v %v\n", "Rank", "Distribution", "Country", "URL", "Avg Time")
		for i, distroMirror := range distroMirrors.Mirrors {
			fmt.Printf("%v: %v %v %v %v\n", i, distroMirrors.Distribution.Name(), distroMirror.Country, distroMirror.URL, distroMirror.Statistics.AvgResponseTimeHTTP)
		}
	case "json":
		distroMirrorsJson, _ := json.Marshal(distroMirrors)
		fmt.Println(string(distroMirrorsJson))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		record := []string{"Rank", "Distribution", "Country", "URL", "Avg Time"}
		if err := w.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "error writing record to file: %v\n", err)
			os.Exit(1)
		}
		for i, distroMirror := range distroMirrors.Mirrors {
			record := []string{fmt.Sprintf("%v", i), distroMirrors.Distribution.Name(), distroMirror.Country, distroMirror.URL.String(), fmt.Sprintf("%v", distroMirror.Statistics.AvgResponseTimeHTTP)}
			if err := w.Write(record); err != nil {
				fmt.Fprintf(os.Stderr, "error writing record to file: %v\n", err)
				os.Exit(1)
			}
		}
	}

}
