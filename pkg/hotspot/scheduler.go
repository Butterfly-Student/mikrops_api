package hotspot

import (
	"context"
	"fmt"

	"github.com/go-routeros/routeros/v3/proto"
	"go-template/pkg/hotspot/internal"
)

// CreateExpiryScheduler creates scheduler for monitoring expired users
func (c *hotspotClient) CreateExpiryScheduler(ctx context.Context, profileName string) error {
	if profileName == "" {
		return NewError("create scheduler", fmt.Errorf("profile name is required"))
	}

	schedulerName := internal.BuildSchedulerName(profileName)

	// Build scheduler script
	schedulerScript := internal.BuildExpirySchedulerScript(profileName)

	_, err := c.execute(ctx, PathSystemScheduler+"/add",
		"=name="+schedulerName,
		"=interval="+DefaultSchedulerInterval,
		"=start-time=startup",
		"=policy="+PolicyReadWrite,
		"=on-event="+schedulerScript,
	)

	if err != nil {
		return WrapError("create scheduler", err)
	}

	return nil
}

// RemoveExpiryScheduler removes expiry monitoring scheduler
func (c *hotspotClient) RemoveExpiryScheduler(ctx context.Context, profileName string) error {
	if profileName == "" {
		return NewError("remove scheduler", fmt.Errorf("profile name is required"))
	}

	schedulerName := internal.BuildSchedulerName(profileName)

	// Find scheduler
	reply, err := c.execute(ctx, PathSystemScheduler+"/print", "?name="+schedulerName)
	if err != nil {
		return WrapError("remove scheduler", err)
	}

	if len(reply.Re) == 0 {
		// Scheduler not found, ignore
		return nil
	}

	schedulerID := reply.Re[0].Map[".id"]

	// Remove scheduler
	_, err = c.execute(ctx, PathSystemScheduler+"/remove", "=.id="+schedulerID)
	if err != nil {
		return WrapError("remove scheduler", err)
	}

	return nil
}

// GetAllSchedulers retrieves all schedulers
func (c *hotspotClient) GetAllSchedulers(ctx context.Context) ([]Scheduler, error) {
	reply, err := c.execute(ctx, PathSystemScheduler+"/print")
	if err != nil {
		return nil, WrapError("get all schedulers", err)
	}

	schedulers := make([]Scheduler, 0, len(reply.Re))
	for _, re := range reply.Re {
		scheduler := c.mapReplyToScheduler(re)
		schedulers = append(schedulers, *scheduler)
	}

	return schedulers, nil
}

// GetSchedulerByName retrieves scheduler by name
func (c *hotspotClient) GetSchedulerByName(ctx context.Context, name string) (*Scheduler, error) {
	if name == "" {
		return nil, NewError("get scheduler", fmt.Errorf("name is required"))
	}

	reply, err := c.execute(ctx, PathSystemScheduler+"/print", "?name="+name)
	if err != nil {
		return nil, WrapError("get scheduler", err)
	}

	if len(reply.Re) == 0 {
		return nil, ErrSchedulerNotFound
	}

	return c.mapReplyToScheduler(reply.Re[0]), nil
}

func (c *hotspotClient) mapReplyToScheduler(re *proto.Sentence) *Scheduler {
	return &Scheduler{
		Name:      re.Map["name"],
		Interval:  re.Map["interval"],
		StartTime: re.Map["start-time"],
		Policy:    re.Map["policy"],
		OnEvent:   re.Map["on-event"],
		// RouterOS uses "yes"/"no" for boolean fields, not "true"/"false"
		Enabled: re.Map["disabled"] != "yes",
	}
}
