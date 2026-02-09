package command_inbound_adapter

import (
	"context"

	"mikrops/internal/domain"
	"mikrops/utils/log"
)

// CutoffCommand handles daily cutoff cron job
type CutoffCommand struct {
	domain domain.Domain
}

func NewCutoffCommand(d domain.Domain) *CutoffCommand {
	return &CutoffCommand{domain: d}
}

// Run executes the daily cutoff job
func (c *CutoffCommand) Run() error {
	ctx := context.Background()
	log.WithContext(ctx).Info("Starting daily cutoff job")
	
	// Execute the daily cutoff logic from the domain
	err := c.domain.Cutoff().RunDailyCutoff(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("Daily cutoff job failed: %v", err)
		return err
	}
	
	log.WithContext(ctx).Info("Daily cutoff job completed successfully")
	return nil
}
