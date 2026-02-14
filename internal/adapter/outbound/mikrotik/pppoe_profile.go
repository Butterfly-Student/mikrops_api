package mikrotik_outbound_adapter

import (
	"fmt"
	"strings"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3/proto"
)

// PppoeProfile

func (a *mikrotikClientAdapter) CreateProfile(router *model.MikrotikRouter, profile *model.PppoeProfile) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	cmd := []string{
		"/ppp/profile/add",
		"=name=" + profile.Name,
	}

	if profile.LocalAddress != "" {
		cmd = append(cmd, "=local-address="+profile.LocalAddress)
	}
	if profile.RemoteAddress != "" {
		cmd = append(cmd, "=remote-address="+profile.RemoteAddress)
	}
	if profile.RateLimit != "" {
		cmd = append(cmd, "=rate-limit="+profile.RateLimit)
	}
	if profile.DNSServer != "" {
		cmd = append(cmd, "=dns-server="+profile.DNSServer)
	}
	if profile.OnlyOne != "" {
		cmd = append(cmd, "=only-one="+profile.OnlyOne)
	}

	_, err = client.RunArgs(cmd)
	return err
}

func (a *mikrotikClientAdapter) UpdateProfile(router *model.MikrotikRouter, profile *model.PppoeProfile) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	id := profile.ID
	if id == "" {
		reply, err := client.Run("/ppp/profile/print", "?name="+profile.Name)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return fmt.Errorf("profile not found")
		}
		id = reply.Re[0].Map[".id"]
	}

	cmd := []string{
		"/ppp/profile/set",
		"=.id=" + id,
	}

	if profile.LocalAddress != "" {
		cmd = append(cmd, "=local-address="+profile.LocalAddress)
	}
	if profile.RemoteAddress != "" {
		cmd = append(cmd, "=remote-address="+profile.RemoteAddress)
	}
	if profile.RateLimit != "" {
		cmd = append(cmd, "=rate-limit="+profile.RateLimit)
	}
	if profile.DNSServer != "" {
		cmd = append(cmd, "=dns-server="+profile.DNSServer)
	}

	_, err = client.RunArgs(cmd)
	return err
}

func (a *mikrotikClientAdapter) DeleteProfile(router *model.MikrotikRouter, id string) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(id, "*") {
		reply, err := client.Run("/ppp/profile/print", "?name="+id)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return nil // Already deleted
		}
		id = reply.Re[0].Map[".id"]
	}

	_, err = client.Run("/ppp/profile/remove", "=.id="+id)
	return err
}

func (a *mikrotikClientAdapter) GetProfile(router *model.MikrotikRouter, id string) (*model.PppoeProfile, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	var cmd []string
	if strings.HasPrefix(id, "*") {
		cmd = []string{"/ppp/profile/print", "?.id=" + id}
	} else {
		cmd = []string{"/ppp/profile/print", "?name=" + id}
	}

	reply, err := client.RunArgs(cmd)
	if err != nil {
		return nil, err
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("profile not found")
	}

	return parseProfile(reply.Re[0]), nil
}

func (a *mikrotikClientAdapter) ListProfiles(router *model.MikrotikRouter) ([]model.PppoeProfile, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	reply, err := client.Run("/ppp/profile/print")
	if err != nil {
		return nil, err
	}

	var profiles []model.PppoeProfile
	for _, re := range reply.Re {
		profiles = append(profiles, *parseProfile(re))
	}
	return profiles, nil
}

func parseProfile(re *proto.Sentence) *model.PppoeProfile {
	return &model.PppoeProfile{
		ID:            re.Map[".id"],
		Name:          re.Map["name"],
		LocalAddress:  re.Map["local-address"],
		RemoteAddress: re.Map["remote-address"],
		RateLimit:     re.Map["rate-limit"],
		OnlyOne:       re.Map["only-one"],
		DNSServer:     re.Map["dns-server"],
		Comment:       re.Map["comment"],
	}
}
