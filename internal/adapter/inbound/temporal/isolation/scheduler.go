package isolation_workflow

import (
	"context"
	"fmt"
	"time"

	"go-template/internal/domain"
	"go-template/utils/log"
)

type IsolationScheduler struct {
	domain domain.Domain
}

func NewIsolationScheduler(domain domain.Domain) *IsolationScheduler {
	return &IsolationScheduler{
		domain: domain,
	}
}

// Start starts all isolation schedulers
func (s *IsolationScheduler) Start(ctx context.Context) {
	go s.scheduleDailyIsolationCheck(ctx)
}

// scheduleDailyIsolationCheck schedules daily isolation check for expired customers
func (s *IsolationScheduler) scheduleDailyIsolationCheck(ctx context.Context) {
	log.WithContext(ctx).Info("Starting daily isolation check scheduler")

	// Calculate initial delay to run at midnight
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, 1) // Next midnight
	duration := time.Until(nextRun)

	log.WithContext(ctx).Info(fmt.Sprintf("First isolation check scheduled at %v (in %v)", nextRun, duration))

	// Wait for first run
	select {
	case <-time.After(duration):
		s.triggerIsolationCheck(ctx)
	case <-ctx.Done():
		log.WithContext(ctx).Info("Daily isolation check scheduler stopped before first run")
		return
	}

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.triggerIsolationCheck(ctx)
		case <-ctx.Done():
			log.WithContext(ctx).Info("Daily isolation check scheduler stopped")
			return
		}
	}
}

// triggerIsolationCheck triggers the isolation check workflow
func (s *IsolationScheduler) triggerIsolationCheck(ctx context.Context) {
	log.WithContext(ctx).Info("Triggering isolation check workflow")

	gracePeriodDays := s.getGracePeriodDays()

	err := s.domain.Workflow().Isolation().TriggerIsolateExpiredCustomers(ctx, gracePeriodDays)
	if err != nil {
		log.WithContext(ctx).Error("Failed to trigger isolation check", err)
	} else {
		log.WithContext(ctx).Info(fmt.Sprintf("Isolation check triggered successfully with grace period: %d days", gracePeriodDays))
	}
}

// getGracePeriodDays gets grace period days from system settings
func (s *IsolationScheduler) getGracePeriodDays() int {
	setting, err := s.domain.SystemSetting().GetSetting(context.Background(), "invoice.grace_period_days")
	if err != nil || setting == nil || setting.Value == nil || *setting.Value == "" {
		return 3 // Default grace period
	}

	var days int
	if _, err := fmt.Sscanf(*setting.Value, "%d", &days); err != nil {
		return 3
	}

	if days < 1 {
		days = 1
	} else if days > 30 {
		days = 30
	}

	return days
}
