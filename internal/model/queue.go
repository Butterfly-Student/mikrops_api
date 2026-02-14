package model

type PppoeQueue struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name" validate:"required"`
	Target         string `json:"target" validate:"required"`
	MaxLimit       string `json:"max_limit"` // "upload/download" e.g. "10M/10M"
	BurstLimit     string `json:"burst_limit"`
	BurstThreshold string `json:"burst_threshold"`
	BurstTime      string `json:"burst_time"`
	Priority       string `json:"priority"`
	Parent         string `json:"parent"`
	Comment        string `json:"comment"`
}

type QueueStats struct {
	Name          string `json:"name"`
	BytesIn       int64  `json:"bytes-in"`
	BytesOut      int64  `json:"bytes-out"`
	PacketsIn     int64  `json:"packets-in"`
	PacketsOut    int64  `json:"packets-out"`
	RateIn        int64  `json:"rate-in"`
	RateOut       int64  `json:"rate-out"`
	PacketRateIn  int64  `json:"packet-rate-in"`
	PacketRateOut int64  `json:"packet-rate-out"`
}
