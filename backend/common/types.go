package common

const (
	OperationAdd     = "add"
	OperationDelete  = "delete"
	OperationClear   = "clear"
	OperationReplace = "replace"
)

type StarterConfig struct {
	ServiceAddr string `json:"serviceAddr"`
	RulesFile   string `json:"rulesFile"`
	LogFile     string `json:"logFile"`
	Debug       bool   `json:"debug"`
	Image       bool   `json:"image"`
	Raft        struct {
		Enabled bool     `json:"enabled"`
		Address string   `json:"address"`
		Peers   []string `json:"peers"`
	} `json:"raft"`
}
