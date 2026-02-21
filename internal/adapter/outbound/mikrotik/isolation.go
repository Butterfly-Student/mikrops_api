package mikrotik_outbound_adapter

import (
	"fmt"
	"strings"

	"go-template/internal/model"
)

const isolationCommentPrefix = "mikrops-isolation"

// SetupIsolation creates the isolation PPP profile and firewall rules on the router.
// It is idempotent — existing resources with matching comments are skipped.
func (a *mikrotikClientAdapter) SetupIsolation(router *model.MikrotikRouter, config model.IsolationConfig) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	// 1. Create or update the isolation PPP profile
	profileComment := isolationCommentPrefix + "-profile"
	existingProfile, _ := client.Run("/ppp/profile/print", "?name="+config.ProfileName)
	if existingProfile != nil && len(existingProfile.Re) > 0 {
		// Update existing profile
		id := existingProfile.Re[0].Map[".id"]
		cmd := []string{
			"/ppp/profile/set",
			"=.id=" + id,
			"=address-list=" + config.AddressList,
			"=rate-limit=" + config.RateLimit,
			"=comment=" + profileComment,
		}
		if config.DNSServer != "" {
			cmd = append(cmd, "=dns-server="+config.DNSServer)
		}
		if _, err := client.RunArgs(cmd); err != nil {
			return fmt.Errorf("failed to update isolation profile: %w", err)
		}
	} else {
		// Create new profile
		cmd := []string{
			"/ppp/profile/add",
			"=name=" + config.ProfileName,
			"=address-list=" + config.AddressList,
			"=rate-limit=" + config.RateLimit,
			"=comment=" + profileComment,
		}
		if config.DNSServer != "" {
			cmd = append(cmd, "=dns-server="+config.DNSServer)
		}
		if _, err := client.RunArgs(cmd); err != nil {
			return fmt.Errorf("failed to create isolation profile: %w", err)
		}
	}

	// 2. Create firewall NAT dst-nat rule (redirect HTTP to portal)
	natComment := isolationCommentPrefix + "-redirect"
	if !a.firewallRuleExists(router, "/ip/firewall/nat/print", natComment) {
		cmd := []string{
			"/ip/firewall/nat/add",
			"=chain=dstnat",
			"=src-address-list=" + config.AddressList,
			"=dst-port=80",
			"=protocol=tcp",
			"=action=dst-nat",
			"=to-addresses=" + config.PortalIP,
			"=to-ports=" + config.PortalPort,
			"=comment=" + natComment,
		}
		if _, err := client.RunArgs(cmd); err != nil {
			return fmt.Errorf("failed to create NAT redirect rule: %w", err)
		}
	}

	// 3. Create firewall filter rules (order matters)
	filterRules := []struct {
		comment  string
		args     []string
	}{
		{
			comment: isolationCommentPrefix + "-allow-dns-udp",
			args: []string{
				"=chain=forward",
				"=src-address-list=" + config.AddressList,
				"=dst-port=53",
				"=protocol=udp",
				"=action=accept",
			},
		},
		{
			comment: isolationCommentPrefix + "-allow-dns-tcp",
			args: []string{
				"=chain=forward",
				"=src-address-list=" + config.AddressList,
				"=dst-port=53",
				"=protocol=tcp",
				"=action=accept",
			},
		},
		{
			comment: isolationCommentPrefix + "-allow-portal",
			args: []string{
				"=chain=forward",
				"=src-address-list=" + config.AddressList,
				"=dst-address=" + config.PortalIP,
				"=action=accept",
			},
		},
		{
			comment: isolationCommentPrefix + "-drop-rest",
			args: []string{
				"=chain=forward",
				"=src-address-list=" + config.AddressList,
				"=action=drop",
			},
		},
	}

	for _, rule := range filterRules {
		if !a.firewallRuleExists(router, "/ip/firewall/filter/print", rule.comment) {
			cmd := append([]string{"/ip/firewall/filter/add"}, rule.args...)
			cmd = append(cmd, "=comment="+rule.comment)
			if _, err := client.RunArgs(cmd); err != nil {
				return fmt.Errorf("failed to create firewall filter rule %s: %w", rule.comment, err)
			}
		}
	}

	return nil
}

// CheckIsolationSetup checks if the isolation profile and firewall rules exist on the router.
func (a *mikrotikClientAdapter) CheckIsolationSetup(router *model.MikrotikRouter) (bool, error) {
	client, err := a.getClient(router)
	if err != nil {
		return false, err
	}

	// Check if isolation profile exists
	reply, err := client.Run("/ppp/profile/print", "?name=isolir")
	if err != nil {
		return false, fmt.Errorf("failed to check isolation profile: %w", err)
	}
	if len(reply.Re) == 0 {
		return false, nil
	}

	// Check if NAT redirect rule exists
	if !a.firewallRuleExists(router, "/ip/firewall/nat/print", isolationCommentPrefix+"-redirect") {
		return false, nil
	}

	// Check if drop rule exists (last filter rule)
	if !a.firewallRuleExists(router, "/ip/firewall/filter/print", isolationCommentPrefix+"-drop-rest") {
		return false, nil
	}

	return true, nil
}

// firewallRuleExists checks if a firewall rule with the given comment exists.
func (a *mikrotikClientAdapter) firewallRuleExists(router *model.MikrotikRouter, printCmd string, comment string) bool {
	client, err := a.getClient(router)
	if err != nil {
		return false
	}

	reply, err := client.RunArgs([]string{printCmd})
	if err != nil {
		return false
	}

	for _, re := range reply.Re {
		if strings.Contains(re.Map["comment"], comment) {
			return true
		}
	}
	return false
}
