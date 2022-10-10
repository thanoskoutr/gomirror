package distributions

import (
	"log"

	"github.com/thanoskoutr/gomirror/mirrors"
)

// TODO: Add generic Distributor implementation for using tool with non natively supported distributions
type CustomDistributor struct {
	CustomName string
}

func (d *CustomDistributor) SetName(name string) { d.CustomName = name }

func (d CustomDistributor) Name() string { return d.CustomName }

func (d CustomDistributor) GetMirrors(source mirrors.MirrorSource, filename string) []mirrors.Mirror {
	switch source {
	case mirrors.SourceHTTP:
		// TODO: Not supported for unknown distributions, except if in known and valid JSON or TXT format
		log.Fatal("Unimplemented functionality: sourceHTTP")
		return []mirrors.Mirror{}
	case mirrors.SourceJSON:
		return mirrors.ReadMirrorsJSON(filename)
	case mirrors.SourceTXT:
		return mirrors.ReadMirrorsTXT(filename)
	default:
		return []mirrors.Mirror{}
	}
}
