package model

import (
	"time"
)

type FirewallRule struct {
	Chain          string    `json:"chain"`
	Protocol       string    `json:"protocol"`
	SrcAddress     string    `json:"src_address"`
	SrcAddressList string    `json:"src_address_list"`
	DstPort        int       `json:"dst_port"`
	DstPortList    string    `json:"dst_port_list"`
	DstAddress     string    `json:"dst_address"`
	DstAddressList string    `json:"dst_address_list"`
	ToPorts        string    `json:"to_ports"`
	ToAddress      string    `json:"to_address"`
	Action         string    `json:"action"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
}

type FirewallRuleInput struct {
	Chain          string `json:"chain" validate:"required"`
	Protocol       string `json:"protocol" validate:"required,oneof=tcp udp icmp"`
	SrcAddress     string `json:"src_address"`
	SrcAddressList string `json:"src_address_list"`
	DstPort        int    `json:"dst_port" validate:"min=1,max=65535"`
	DstPortList    string `json:"dst_port_list"`
	DstAddress     string `json:"dst_address"`
	DstAddressList string `json:"dst_address_list"`
	ToPorts        string `json:"to_ports"`
	ToAddress      string `json:"to_address"`
	Action         string `json:"action" validate:"required,oneof=accept reject drop redirect"`
	Comment        string `json:"comment"`
}

func FirewallRulePrepare(v *FirewallRuleInput) {
	if v.Chain == "" {
		v.Chain = "forward"
	}
	if v.Protocol == "" {
		v.Protocol = "tcp"
	}
	if v.Action == "" {
		v.Action = "accept"
	}
}
