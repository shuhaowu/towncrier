package webreceiver

type ReceiverConfig struct {
	ListenHost string
	ListenPort int
	PathPrefix string            // Without the trailing slash
	Tokens     map[string]string // api token => origin name
}
