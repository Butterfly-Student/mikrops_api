package mikrotik_outbound_adapter

import (
	"fmt"
	"strings"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3/proto"
)

// PppoeActive

func (a *mikrotikClientAdapter) ListActiveSessions(router *model.MikrotikRouter) ([]model.PppoeActive, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	reply, err := client.Run("/ppp/active/print")
	if err != nil {
		return nil, err
	}

	var sessions []model.PppoeActive
	for _, re := range reply.Re {
		sessions = append(sessions, *parseActive(re))
	}
	return sessions, nil
}

func (a *mikrotikClientAdapter) GetActiveSession(router *model.MikrotikRouter, id string) (*model.PppoeActive, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	cmd := []string{"/ppp/active/print"}
	if strings.HasPrefix(id, "*") {
		cmd = append(cmd, "?.id="+id)
	} else {
		cmd = append(cmd, "?name="+id)
	}

	reply, err := client.RunArgs(cmd)
	if err != nil {
		return nil, err
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("active session not found")
	}

	return parseActive(reply.Re[0]), nil
}

func (a *mikrotikClientAdapter) RemoveActiveSession(router *model.MikrotikRouter, id string) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	// Active sessions usually identified by .id
	if !strings.HasPrefix(id, "*") {
		// If passed username, find ID first
		reply, err := client.Run("/ppp/active/print", "?name="+id)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return nil
		}
		id = reply.Re[0].Map[".id"]
	}

	_, err = client.Run("/ppp/active/remove", "=.id="+id)
	return err
}

func parseActive(re *proto.Sentence) *model.PppoeActive {
	return &model.PppoeActive{
		ID:        re.Map[".id"],
		Name:      re.Map["name"],
		Service:   re.Map["service"],
		CallerID:  re.Map["caller-id"],
		Address:   re.Map["address"],
		Uptime:    re.Map["uptime"],
		Encoding:  re.Map["encoding"],
		SessionID: re.Map["session-id"],
	}
}
