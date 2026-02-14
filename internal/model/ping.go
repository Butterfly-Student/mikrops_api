package model

// PingRequest represents a ping request configuration
type PingRequest struct {
	Address  string `json:"address" validate:"required"`
	Count    int    `json:"count"`    // 0 = continuous
	Size     int    `json:"size"`     // packet size in bytes, default: 56
	Interval string `json:"interval"` // e.g., "1s", default: "1s"
}

// PingResult represents a single ping result
type PingResult struct {
	Host         string `json:"host"`
	Seq          int    `json:"seq"`
	Size         int    `json:"size"`
	TTL          int    `json:"ttl"`
	Time         string `json:"time"`   // e.g., "5ms"
	Status       string `json:"status"` // "success", "timeout", "error"
	SentTime     string `json:"sent_time"`
	ReceivedTime string `json:"received_time"`
}

// PingSummary represents aggregated ping statistics
type PingSummary struct {
	Host       string  `json:"host"`
	Sent       int     `json:"sent"`
	Received   int     `json:"received"`
	PacketLoss float64 `json:"packet_loss"`
	MinTime    string  `json:"min_time"`
	AvgTime    string  `json:"avg_time"`
	MaxTime    string  `json:"max_time"`
}
