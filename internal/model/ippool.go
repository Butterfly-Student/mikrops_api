package model

// IpPool represents a MikroTik IP Pool
type IpPool struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name" validate:"required"`
	Ranges   string `json:"ranges" validate:"required"` // e.g., "192.168.1.100-192.168.1.200"
	NextPool string `json:"next_pool"`
	Comment  string `json:"comment"`
}
