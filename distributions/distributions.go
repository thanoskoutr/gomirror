package distributions

import (
	"fmt"

	"github.com/thanoskoutr/gomirror/mirrors"
)

type Distributor interface {
	Name() string
	// TODO: Return []*mirrors.Mirror
	GetMirrors(source mirrors.MirrorSource, filename string) []mirrors.Mirror
}

func ToDistribution(distro string) (Distributor, error) {
	switch distro {
	case "Ubuntu":
		return Ubuntu{}, nil
	case "Debian":
		return Debian{}, nil
	case "Arch":
		return Arch{}, nil
	default:
		return nil, fmt.Errorf("unsupported distribution: %v", distro)
	}
}
