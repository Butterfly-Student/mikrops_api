package mikrotik_outbound_adapter

import (
	"fmt"
	"strings"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3"
)

// Queue Management

func (a *mikrotikClientAdapter) CreateQueue(router *model.MikrotikRouter, queue *model.PppoeQueue) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	cmd := []string{
		"/queue/simple/add",
		"=name=" + queue.Name,
		"=target=" + queue.Target,
	}

	if queue.MaxLimit != "" {
		cmd = append(cmd, "=max-limit="+queue.MaxLimit)
	}
	if queue.BurstLimit != "" {
		cmd = append(cmd, "=burst-limit="+queue.BurstLimit)
	}
	if queue.BurstThreshold != "" {
		cmd = append(cmd, "=burst-threshold="+queue.BurstThreshold)
	}
	if queue.BurstTime != "" {
		cmd = append(cmd, "=burst-time="+queue.BurstTime)
	}
	if queue.Priority != "" {
		cmd = append(cmd, "=priority="+queue.Priority)
	}
	if queue.Parent != "" {
		cmd = append(cmd, "=parent="+queue.Parent)
	}
	if queue.Comment != "" {
		cmd = append(cmd, "=comment="+queue.Comment)
	}

	_, err = client.RunArgs(cmd)
	return err
}

func (a *mikrotikClientAdapter) UpdateQueue(router *model.MikrotikRouter, queue *model.PppoeQueue) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	id := queue.ID
	if id == "" {
		reply, err := client.Run("/queue/simple/print", "?name="+queue.Name)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return fmt.Errorf("queue not found")
		}
		id = reply.Re[0].Map[".id"]
	}

	cmd := []string{
		"/queue/simple/set",
		"=.id=" + id,
	}

	if queue.Target != "" {
		cmd = append(cmd, "=target="+queue.Target)
	}
	if queue.MaxLimit != "" {
		cmd = append(cmd, "=max-limit="+queue.MaxLimit)
	}
	if queue.BurstLimit != "" {
		cmd = append(cmd, "=burst-limit="+queue.BurstLimit)
	}
	if queue.BurstThreshold != "" {
		cmd = append(cmd, "=burst-threshold="+queue.BurstThreshold)
	}
	if queue.BurstTime != "" {
		cmd = append(cmd, "=burst-time="+queue.BurstTime)
	}
	if queue.Priority != "" {
		cmd = append(cmd, "=priority="+queue.Priority)
	}
	if queue.Parent != "" {
		cmd = append(cmd, "=parent="+queue.Parent)
	}
	if queue.Comment != "" {
		cmd = append(cmd, "=comment="+queue.Comment)
	}

	_, err = client.RunArgs(cmd)
	return err
}

func (a *mikrotikClientAdapter) DeleteQueue(router *model.MikrotikRouter, id string) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(id, "*") {
		reply, err := client.Run("/queue/simple/print", "?name="+id)
		if err != nil {
			return err
		}
		if len(reply.Re) == 0 {
			return nil
		}
		id = reply.Re[0].Map[".id"]
	}

	_, err = client.Run("/queue/simple/remove", "=.id="+id)
	return err
}

func (a *mikrotikClientAdapter) GetQueue(router *model.MikrotikRouter, id string) (*model.PppoeQueue, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	cmd := []string{"/queue/simple/print"}
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
		return nil, fmt.Errorf("queue not found")
	}

	return parseQueue(reply.Re[0]), nil
}

func (a *mikrotikClientAdapter) ListQueues(router *model.MikrotikRouter) ([]model.PppoeQueue, error) {
	client, err := a.getClient(router)
	if err != nil {
		return nil, err
	}

	reply, err := client.Run("/queue/simple/print")
	if err != nil {
		return nil, err
	}

	var queues []model.PppoeQueue
	for _, re := range reply.Re {
		queues = append(queues, *parseQueue(re))
	}
	return queues, nil
}

func parseQueue(re *routeros.ReplyPair) *model.PppoeQueue {
	return &model.PppoeQueue{
		ID:             re.Map[".id"],
		Name:           re.Map["name"],
		Target:         re.Map["target"],
		MaxLimit:       re.Map["max-limit"],
		BurstLimit:     re.Map["burst-limit"],
		BurstThreshold: re.Map["burst-threshold"],
		BurstTime:      re.Map["burst-time"],
		Priority:       re.Map["priority"],
		Parent:         re.Map["parent"],
		Comment:        re.Map["comment"],
	}
}
