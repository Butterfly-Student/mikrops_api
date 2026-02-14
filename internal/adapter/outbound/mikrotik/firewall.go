package mikrotik_outbound_adapter

import (
	"fmt"

	"go-template/internal/model"

	"github.com/palantir/stacktrace"
)

// AddFirewallRule adds a firewall rule to MikroTik router
func (a *mikrotikClientAdapter) AddFirewallRule(router *model.MikrotikRouter, rule model.FirewallRule) error {
	client, err := a.getClient(router)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik client")
	}

	// Build command based on rule properties
	cmd := buildFirewallCommand(rule)

	_, err = client.Run(cmd...)
	if err != nil {
		return stacktrace.Propagate(err, "failed to add firewall rule")
	}

	return nil
}

// RemoveFirewallRule removes a firewall rule from MikroTik router
// Note: This removes based on matching rule properties
func (a *mikrotikClientAdapter) RemoveFirewallRule(router *model.MikrotikRouter, rule model.FirewallRule) error {
	client, err := a.getClient(router)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get mikrotik client")
	}

	// List all rules to find matching one
	reply, err := client.Run("/ip/firewall/nat/print")
	if err != nil {
		return stacktrace.Propagate(err, "failed to list firewall rules")
	}

	// Find matching rule ID
	var ruleID string
	for _, re := range reply.Re {
		if re.Map["chain"] == rule.Chain &&
			re.Map["action"] == rule.Action &&
			re.Map["protocol"] == rule.Protocol &&
			re.Map["src-address"] == rule.SrcAddress &&
			re.Map["dst-address"] == rule.DstAddress {
			ruleID = re.Map[".id"]
			break
		}
	}

	if ruleID == "" {
		return stacktrace.NewError("firewall rule not found")
	}

	_, err = client.Run("/ip/firewall/nat/remove", fmt.Sprintf("=.id=%s", ruleID))
	if err != nil {
		return stacktrace.Propagate(err, "failed to remove firewall rule")
	}

	return nil
}

// ListFirewallRules lists all firewall NAT rules from MikroTik router
func (a *mikrotikClientAdapter) ListFirewallRules(router *model.MikrotikRouter) ([]model.FirewallRule, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mikrotik client")
	}

	reply, err := client.Run("/ip/firewall/nat/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list firewall rules")
	}

	var rules []model.FirewallRule
	for _, re := range reply.Re {
		dstPort := 0
		if portStr := re.Map["dst-port"]; portStr != "" {
			fmt.Sscanf(portStr, "%d", &dstPort)
		}

		rule := model.FirewallRule{
			Chain:          re.Map["chain"],
			Action:         re.Map["action"],
			Protocol:       re.Map["protocol"],
			SrcAddress:     re.Map["src-address"],
			SrcAddressList: re.Map["src-address-list"],
			DstAddress:     re.Map["dst-address"],
			DstAddressList: re.Map["dst-address-list"],
			DstPort:        dstPort,
			DstPortList:    re.Map["dst-port-list"],
			ToAddress:      re.Map["to-addresses"],
			ToPorts:        re.Map["to-ports"],
			Comment:        re.Map["comment"],
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// buildFirewallCommand builds the RouterOS command for adding a firewall rule
func buildFirewallCommand(rule model.FirewallRule) []string {
	cmd := []string{"/ip/firewall/nat/add"}

	if rule.Chain != "" {
		cmd = append(cmd, fmt.Sprintf("=chain=%s", rule.Chain))
	}
	if rule.Action != "" {
		cmd = append(cmd, fmt.Sprintf("=action=%s", rule.Action))
	}
	if rule.Protocol != "" {
		cmd = append(cmd, fmt.Sprintf("=protocol=%s", rule.Protocol))
	}
	if rule.SrcAddress != "" {
		cmd = append(cmd, fmt.Sprintf("=src-address=%s", rule.SrcAddress))
	}
	if rule.SrcAddressList != "" {
		cmd = append(cmd, fmt.Sprintf("=src-address-list=%s", rule.SrcAddressList))
	}
	if rule.DstAddress != "" {
		cmd = append(cmd, fmt.Sprintf("=dst-address=%s", rule.DstAddress))
	}
	if rule.DstAddressList != "" {
		cmd = append(cmd, fmt.Sprintf("=dst-address-list=%s", rule.DstAddressList))
	}
	if rule.DstPort > 0 {
		cmd = append(cmd, fmt.Sprintf("=dst-port=%d", rule.DstPort))
	}
	if rule.DstPortList != "" {
		cmd = append(cmd, fmt.Sprintf("=dst-port-list=%s", rule.DstPortList))
	}
	if rule.ToAddress != "" {
		cmd = append(cmd, fmt.Sprintf("=to-addresses=%s", rule.ToAddress))
	}
	if rule.ToPorts != "" {
		cmd = append(cmd, fmt.Sprintf("=to-ports=%s", rule.ToPorts))
	}
	if rule.Comment != "" {
		cmd = append(cmd, fmt.Sprintf("=comment=%s", rule.Comment))
	}

	return cmd
}
