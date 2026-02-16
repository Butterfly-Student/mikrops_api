package temporal_inbound_adapter

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/log"
)

func InitRoute(
	ctx context.Context,
	args []string,
	port inbound_port.WorkflowPort,
) {
	if len(args) > 2 {
		switch args[2] {
		case "upsert_client":
			port.Client().Upsert()
			return
		case "start-workers":
			startAllWorkers(ctx, port)
			return
		case "billing-worker":
			startBillingWorkers(ctx, port)
			return
		case "isolation-worker":
			startIsolationWorkers(ctx, port)
			return
		default:
			log.WithContext(ctx).Info("command not found")
			printUsage()
		}
	} else {
		log.WithContext(ctx).Info("command not found")
		printUsage()
	}
}

func printUsage() {
	ctx := context.Background()
	log.WithContext(ctx).Info("Available commands:")
	log.WithContext(ctx).Info("  upsert_client              - Upsert client data (one-time workflow)")
	log.WithContext(ctx).Info("  start-workers              - Start all temporal workers with schedulers")
	log.WithContext(ctx).Info("  billing-worker             - Start billing workers (invoice generation + payment processing)")
	log.WithContext(ctx).Info("  isolation-worker           - Start isolation worker (isolation + reactivation)")
}

// startAllWorkers starts all temporal workers with their schedulers
func startAllWorkers(ctx context.Context, port inbound_port.WorkflowPort) {
	log.WithContext(ctx).Info("Starting all temporal workers...")

	// Start workers
	port.Billing().StartInvoiceGenerationWorker(ctx)
	port.Billing().StartPaymentProcessingWorker(ctx)
	port.Isolation().StartIsolationWorker(ctx)

	// Give workers time to start
	time.Sleep(2 * time.Second)

	log.WithContext(ctx).Info("All workers started successfully")

	// Start schedulers
	startBillingScheduler(ctx, port)
	startIsolationScheduler(ctx, port)

	// Wait for interrupt signal
	waitForInterrupt(ctx)
}

// startBillingWorkers starts only billing workers with scheduler
func startBillingWorkers(ctx context.Context, port inbound_port.WorkflowPort) {
	log.WithContext(ctx).Info("Starting billing workers...")

	port.Billing().StartInvoiceGenerationWorker(ctx)
	port.Billing().StartPaymentProcessingWorker(ctx)

	// Give workers time to start
	time.Sleep(2 * time.Second)

	log.WithContext(ctx).Info("Billing workers started successfully")

	// Start billing scheduler
	startBillingScheduler(ctx, port)

	// Wait for interrupt signal
	waitForInterrupt(ctx)
}

// startIsolationWorkers starts only isolation worker with scheduler
func startIsolationWorkers(ctx context.Context, port inbound_port.WorkflowPort) {
	log.WithContext(ctx).Info("Starting isolation worker...")

	port.Isolation().StartIsolationWorker(ctx)

	// Give worker time to start
	time.Sleep(2 * time.Second)

	log.WithContext(ctx).Info("Isolation worker started successfully")

	// Start isolation scheduler
	startIsolationScheduler(ctx, port)

	// Wait for interrupt signal
	waitForInterrupt(ctx)
}

// startBillingScheduler starts the billing scheduler
func startBillingScheduler(ctx context.Context, port inbound_port.WorkflowPort) {
	port.Billing().StartScheduler(ctx)

	log.WithContext(ctx).Info("Billing scheduler started successfully")
	log.WithContext(ctx).Info("  - Monthly invoice generation: Scheduled")
	log.WithContext(ctx).Info("  - Daily overdue check: Scheduled")
	log.WithContext(ctx).Info("  - Daily reminders: Scheduled")
}

// startIsolationScheduler starts the isolation scheduler
func startIsolationScheduler(ctx context.Context, port inbound_port.WorkflowPort) {
	port.Isolation().StartScheduler(ctx)

	log.WithContext(ctx).Info("Isolation scheduler started successfully")
	log.WithContext(ctx).Info("  - Daily isolation check: Scheduled")
}

// waitForInterrupt waits for SIGINT or SIGTERM signal
func waitForInterrupt(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.WithContext(ctx).Info(fmt.Sprintf("Received signal: %v. Shutting down workers...", sig))

	// Additional cleanup can be done here if needed
	log.WithContext(ctx).Info("Workers stopped")
}
