package registry

type RegistryOptions struct {
	IsRegistry        bool
	Address        string
	Host           string
	Image          string
	Heartbeat      int
	Overcommit     float64
	TlsCaCert      string
	TlsCert        string
	TlsKey         string
	TlsVerify      bool
	ArbitraryFlags []string
}
