package mikrotik_outbound_adapter

import (
	"fmt"
	"strings"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

// CreateIpPool creates a new IP pool on the MikroTik router
func (a *mikrotikClientAdapter) CreateIpPool(router *model.MikrotikRouter, pool *model.IpPool) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	cmd := []string{"/ip/pool/add"}

	// Required fields
	cmd = append(cmd, "=name="+pool.Name)
	cmd = append(cmd, "=ranges="+pool.Ranges)

	// Optional fields
	if pool.NextPool != "" {
		cmd = append(cmd, "=next-pool="+pool.NextPool)
	}
	if pool.Comment != "" {
		cmd = append(cmd, "=comment="+pool.Comment)
	}

	_, err = client.RunArgs(cmd)
	return err
}

// UpdateIpPool updates an existing IP pool
func (a *mikrotikClientAdapter) UpdateIpPool(router *model.MikrotikRouter, pool *model.IpPool) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	// If ID is empty, try to find by name
	id := pool.ID
	if id == "" || !strings.HasPrefix(id, "*") {
		reply, err := client.Run("/ip/pool/print", "?name="+pool.Name)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return fmt.Errorf("IP pool not found: %s", pool.Name)
		}
		id = reply.Re[0].Map[".id"]
	}

	cmd := []string{"/ip/pool/set", "=.id=" + id}

	// Update fields
	if pool.Ranges != "" {
		cmd = append(cmd, "=ranges="+pool.Ranges)
	}
	if pool.NextPool != "" {
		cmd = append(cmd, "=next-pool="+pool.NextPool)
	}
	if pool.Comment != "" {
		cmd = append(cmd, "=comment="+pool.Comment)
	}

	_, err = client.RunArgs(cmd)
	return err
}

// DeleteIpPool deletes an IP pool by ID or name
func (a *mikrotikClientAdapter) DeleteIpPool(router *model.MikrotikRouter, id string) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	// If not an ID (doesn't start with *), find by name
	if !strings.HasPrefix(id, "*") {
		reply, err := client.Run("/ip/pool/print", "?name="+id)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			// Already deleted, return success
			return nil
		}
		id = reply.Re[0].Map[".id"]
	}

	_, err = client.Run("/ip/pool/remove", "=.id="+id)
	return err
}

// GetIpPool retrieves a single IP pool by ID or name
func (a *mikrotikClientAdapter) GetIpPool(router *model.MikrotikRouter, id string) (*model.IpPool, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	var reply *routeros.Reply
	if strings.HasPrefix(id, "*") {
		// Search by internal ID
		reply, err = client.Run("/ip/pool/print", "?.id="+id)
	} else {
		// Search by name
		reply, err = client.Run("/ip/pool/print", "?name="+id)
	}

	if err != nil {
		return nil, err
	}

	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("IP pool not found")
	}

	return parseIpPool(reply.Re[0]), nil
}

// ListIpPools retrieves all IP pools
func (a *mikrotikClientAdapter) ListIpPools(router *model.MikrotikRouter) ([]model.IpPool, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	reply, err := client.Run("/ip/pool/print")
	if err != nil {
		return nil, err
	}

	pools := make([]model.IpPool, 0, len(reply.Re))
	for _, re := range reply.Re {
		pools = append(pools, *parseIpPool(re))
	}

	return pools, nil
}

// parseIpPool converts RouterOS response to IpPool model
func parseIpPool(re *proto.Sentence) *model.IpPool {
	return &model.IpPool{
		ID:       re.Map[".id"],
		Name:     re.Map["name"],
		Ranges:   re.Map["ranges"],
		NextPool: re.Map["next-pool"],
		Comment:  re.Map["comment"],
	}
}
