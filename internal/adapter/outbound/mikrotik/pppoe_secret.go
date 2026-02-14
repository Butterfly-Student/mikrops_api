package mikrotik_outbound_adapter

import (
	"fmt"
	"strings"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3/proto"
)

// PppoeSecret

func (a *mikrotikClientAdapter) CreateSecret(router *model.MikrotikRouter, secret *model.PppoeSecret) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	cmd := []string{
		"/ppp/secret/add",
		"=name=" + secret.Name,
		"=password=" + secret.Password,
		"=service=pppoe",
	}

	if secret.Profile != "" {
		cmd = append(cmd, "=profile="+secret.Profile)
	}
	if secret.LocalAddress != "" {
		cmd = append(cmd, "=local-address="+secret.LocalAddress)
	}
	if secret.RemoteAddress != "" {
		cmd = append(cmd, "=remote-address="+secret.RemoteAddress)
	}
	if secret.Comment != "" {
		cmd = append(cmd, "=comment="+secret.Comment)
	}

	_, err = client.RunArgs(cmd)
	return err
}

func (a *mikrotikClientAdapter) UpdateSecret(router *model.MikrotikRouter, secret *model.PppoeSecret) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	// Identify secret by name or id. Usually .id is preferred but we might not have it.
	// We can use 'set' command with 'numbers' arg if we have ID, or find it first.
	// For simplicity, let's assume we find by name if ID is missing.

	targetID := secret.ID
	if targetID == "" {
		// Find ID by Name
		reply, err := client.Run("/ppp/secret/print", "?name="+secret.Name)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return fmt.Errorf("secret not found")
		}
		targetID = reply.Re[0].Map[".id"]
	}

	cmd := []string{
		"/ppp/secret/set",
		"=.id=" + targetID,
	}

	if secret.Password != "" {
		cmd = append(cmd, "=password="+secret.Password)
	}
	if secret.Profile != "" {
		cmd = append(cmd, "=profile="+secret.Profile)
	}
	if secret.LocalAddress != "" {
		cmd = append(cmd, "=local-address="+secret.LocalAddress)
	}
	if secret.RemoteAddress != "" {
		cmd = append(cmd, "=remote-address="+secret.RemoteAddress)
	}
	if secret.Comment != "" {
		cmd = append(cmd, "=comment="+secret.Comment)
	}

	_, err = client.RunArgs(cmd)
	return err
}

func (a *mikrotikClientAdapter) DeleteSecret(router *model.MikrotikRouter, id string) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	// If id is not internal ID (*...), assume it is name?
	// RouterOS 'remove' command usually takes internal ID (numbers).
	// If the user passes name, we need to find ID first.
	if !strings.HasPrefix(id, "*") {
		reply, err := client.Run("/ppp/secret/print", "?name="+id)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			// Already gone?
			return nil
		}
		id = reply.Re[0].Map[".id"]
	}

	_, err = client.Run("/ppp/secret/remove", "=.id="+id)
	return err
}

func (a *mikrotikClientAdapter) GetSecret(router *model.MikrotikRouter, id string) (*model.PppoeSecret, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	var cmd []string
	if strings.HasPrefix(id, "*") {
		cmd = []string{"/ppp/secret/print", "?.id=" + id}
	} else {
		cmd = []string{"/ppp/secret/print", "?name=" + id}
	}

	reply, err := client.RunArgs(cmd)
	if err != nil {
		return nil, err
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("secret not found")
	}

	return parseSecret(reply.Re[0]), nil
}

func (a *mikrotikClientAdapter) ListSecrets(router *model.MikrotikRouter) ([]model.PppoeSecret, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	reply, err := client.Run("/ppp/secret/print")
	if err != nil {
		return nil, err
	}

	var secrets []model.PppoeSecret
	for _, re := range reply.Re {
		secrets = append(secrets, *parseSecret(re))
	}
	return secrets, nil
}

func parseSecret(re *proto.Sentence) *model.PppoeSecret {
	return &model.PppoeSecret{
		ID:            re.Map[".id"],
		Name:          re.Map["name"],
		Password:      re.Map["password"],
		Service:       re.Map["service"],
		Profile:       re.Map["profile"],
		LocalAddress:  re.Map["local-address"],
		RemoteAddress: re.Map["remote-address"],
		Comment:       re.Map["comment"],
		Disabled:      re.Map["disabled"] == "true",
	}
}
