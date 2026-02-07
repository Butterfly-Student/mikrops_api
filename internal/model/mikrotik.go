package model

type MikrotikConnection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Service  string `json:"service"`
	CallerID string `json:"caller_id"`
	Address  string `json:"address"`
	Uptime   string `json:"uptime"`
}

type MikrotikTraffic struct {
	Interface string `json:"interface"`
	RxBytes   int64  `json:"rx_bytes"`
	TxBytes   int64  `json:"tx_bytes"`
	RxPackets int64  `json:"rx_packets"`
	TxPackets int64  `json:"tx_packets"`
}

type MikrotikBandwidth struct {
	Target   string `json:"target"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

type MikrotikPPPoESecret struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Service  string `json:"service"`
	Profile  string `json:"profile"`
	Disabled string `json:"disabled"`
	Comment  string `json:"comment"`
}

type MikrotikPPPoESecretInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Service  string `json:"service"`
	Profile  string `json:"profile"`
	Comment  string `json:"comment"`
}

type MikrotikHotspotUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Profile  string `json:"profile"`
	Disabled string `json:"disabled"`
	Comment  string `json:"comment"`
}

type MikrotikHotspotUserInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Profile  string `json:"profile"`
	Comment  string `json:"comment"`
}

type MikrotikSimpleQueue struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Target    string `json:"target"`
	MaxLimit  string `json:"max_limit"`
	BurstLimit string `json:"burst_limit"`
	Disabled  string `json:"disabled"`
	Comment   string `json:"comment"`
}

type MikrotikSimpleQueueInput struct {
	Name       string `json:"name"`
	Target     string `json:"target"`
	MaxLimit   string `json:"max_limit"`
	BurstLimit string `json:"burst_limit"`
	Comment    string `json:"comment"`
}

type MikrotikProfile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RateLimit string `json:"rate_limit"`
	Comment   string `json:"comment"`
}

type MikrotikProfileInput struct {
	Name      string `json:"name"`
	RateLimit string `json:"rate_limit"`
	Comment   string `json:"comment"`
}
