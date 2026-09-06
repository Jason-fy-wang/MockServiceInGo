package common

const (
	OperationAdd     = "add"
	OperationDelete  = "delete"
	OperationClear   = "clear"
	OperationReplace = "replace"
)

type TLS struct {
	Enable   bool   `json:"enable"`
	CAFile   string `json:"caFile"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
	MTLS     bool   `json:"mtls"`
}

type RaftConfig struct {
	Enabled bool     `json:"enabled"`
	Address string   `json:"address"`
	Peers   []string `json:"peers"`
	TLS     TLS      `json:"tls"`
}

type StarterConfig struct {
	ServiceAddr string     `json:"serviceAddr"`
	TLS         TLS        `json:"tls"`
	RulesFile   string     `json:"rulesFile"`
	LogFile     string     `json:"logFile"`
	Debug       bool       `json:"debug"`
	Image       bool       `json:"image"`
	Raft        RaftConfig `json:"raft"`
}
