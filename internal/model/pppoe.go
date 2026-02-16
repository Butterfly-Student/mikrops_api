package model

type PppoeSecret struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Service       string `json:"service"`
	CallerID      string `json:"caller_id"`
	Profile       string `json:"profile"`
	LocalAddress  string `json:"local_address"`
	RemoteAddress string `json:"remote_address"`
	Routes        string `json:"routes"`
	Comment       string `json:"comment"`
	Disabled      bool   `json:"disabled"`
}

type PppoeProfile struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name" validate:"required"`
	LocalAddress   string `json:"local_address"`
	RemoteAddress  string `json:"remote_address"`
	Bridge         string `json:"bridge"`
	ChangeTCPMSS   string `json:"change_tcp_mss"`
	RateLimit      string `json:"rate_limit"`
	OnlyOne        string `json:"only_one"`
	UseMPLS        string `json:"use_mpls"`
	UseCompression string `json:"use_compression"`
	UseEncryption  string `json:"use_encryption"`
	UseIPv6        string `json:"use_ipv6"`
	DNSServer      string `json:"dns_server"`
	Comment        string `json:"comment"`
}

type PppoeActive struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	CallerID      string `json:"caller_id"`
	Address       string `json:"address"`
	Uptime        string `json:"uptime"`
	Encoding      string `json:"encoding"`
	SessionID     string `json:"session_id"`
	LimitBytesIn  int64  `json:"limit_bytes_in"`
	LimitBytesOut int64  `json:"limit_bytes_out"`
	Radius        bool   `json:"radius"`
}

type PppoeCallbackData struct {
	User       string `json:"user"`
	IP         string `json:"ip-address"`
	CallerID   string `json:"caller-id"`
	SessionID  string `json:"session-id"`
	Interface  string `json:"interface"`
	Uptime     string `json:"uptime"`
	BytesIn    int64  `json:"bytes-in"`
	BytesOut   int64  `json:"bytes-out"`
	PacketsIn  int64  `json:"packets-in"`
	PacketsOut int64  `json:"packets-out"`
	RouterID   uint   `json:"router_id"`
}

type WebSocketMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
