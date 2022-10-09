package mirrors

type Distributor interface {
	Name() string
	GetMirrors(source MirrorSource, filename string) []Mirror
}
