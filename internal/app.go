package internal

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	command_inbound_adapter "go-template/internal/adapter/inbound/command"
	gin_inbound_adapter "go-template/internal/adapter/inbound/gin"
	rabbitmq_inbound_adapter "go-template/internal/adapter/inbound/rabbitmq"
	temporal_inbound_adapter "go-template/internal/adapter/inbound/temporal"
	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	rabbitmq_outbound_adapter "go-template/internal/adapter/outbound/rabbitmq"
	redis_outbound_adapter "go-template/internal/adapter/outbound/redis"
	temporal_outbound_adapter "go-template/internal/adapter/outbound/temporal"
	"go-template/internal/domain"
	_ "go-template/internal/migration/postgres"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils"
	"go-template/utils/activity"
	"go-template/utils/database"
	"go-template/utils/log"
	"go-template/utils/rabbitmq"
	"go-template/utils/redis"

	"github.com/casbin/casbin/v2"
)

var databaseDriverList = []string{"postgres"}
var httpDriverList = []string{"fiber"}
var messageDriverList = []string{"rabbitmq"}
var workflowDriverList = []string{"temporal"}
var outboundDatabaseDriver string
var outboundMessageDriver string
var outboundCacheDriver string
var outboundWorkflowDriver string
var inboundHttpDriver string
var inboundMessageDriver string
var inboundWorkflowDriver string

type App struct {
	ctx    context.Context
	domain domain.Domain
}

func NewApp() *App {
	ctx := activity.NewContext("init")
	ctx = activity.WithClientID(ctx, "system")
	_ = godotenv.Load(".env")
	configureLogging()
	outboundDatabaseDriver = os.Getenv("OUTBOUND_DATABASE_DRIVER")
	outboundMessageDriver = os.Getenv("OUTBOUND_MESSAGE_DRIVER")
	outboundCacheDriver = os.Getenv("OUTBOUND_CACHE_DRIVER")
	outboundWorkflowDriver = os.Getenv("OUTBOUND_WORKFLOW_DRIVER")
	inboundHttpDriver = os.Getenv("INBOUND_HTTP_DRIVER")
	inboundMessageDriver = os.Getenv("INBOUND_MESSAGE_DRIVER")
	inboundWorkflowDriver = os.Getenv("INBOUND_WORKFLOW_DRIVER")
	dbPort, enforcer := databaseOutbound(ctx)
	domain := domain.NewDomain(
		dbPort,
		messageOutbound(ctx),
		cacheOutbound(ctx),
		workflowOutbound(ctx),
		enforcer,
	)

	return &App{
		ctx:    ctx,
		domain: domain,
	}
}

func (a *App) Run(option string) {
	switch option {
	case "http":
		a.httpInbound()
	case "message":
		a.messageInbound()
	case "workflow":
		a.workflowInbound()
	default:
		a.commandInbound()
	}
}

func databaseOutbound(ctx context.Context) (outbound_port.DatabasePort, *casbin.Enforcer) {
	if !utils.IsInList(databaseDriverList, outboundDatabaseDriver) {
		log.WithContext(ctx).Error("database driver is not supported")
		os.Exit(1)
	}
	db := database.InitDatabase(ctx, outboundDatabaseDriver)

	switch outboundDatabaseDriver {
	case "postgres":
		return postgres_outbound_adapter.NewAdapter(db), postgres_outbound_adapter.InitCasbin(db)
	}
	return nil, nil
}

func messageOutbound(ctx context.Context) outbound_port.MessagePort {
	if !utils.IsInList(messageDriverList, outboundMessageDriver) {
		log.WithContext(ctx).Error("message driver is not supported")
		os.Exit(1)
	}

	switch outboundMessageDriver {
	case "rabbitmq":
		if err := rabbitmq.InitMessage(); err != nil {
			log.WithContext(ctx).Error("failed to init rabbitmq", err)
			os.Exit(1)
		}
		return rabbitmq_outbound_adapter.NewAdapter()
	}
	return nil
}

func cacheOutbound(ctx context.Context) outbound_port.CachePort {
	if !utils.IsInList([]string{"redis"}, outboundCacheDriver) {
		log.WithContext(ctx).Error("cache driver is not supported")
		os.Exit(1)
	}

	switch outboundCacheDriver {
	case "redis":
		redis.InitDatabase()
		return redis_outbound_adapter.NewAdapter()
	}
	return nil
}

func workflowOutbound(ctx context.Context) outbound_port.WorkflowPort {
	if !utils.IsInList([]string{"temporal"}, outboundWorkflowDriver) {
		log.WithContext(ctx).Error("workflow driver is not supported")
		os.Exit(1)
	}

	switch outboundWorkflowDriver {
	case "temporal":
		return temporal_outbound_adapter.NewAdapter()
	}
	return nil
}

func (a *App) httpInbound() {
	ctx := a.ctx
	if !utils.IsInList(httpDriverList, inboundHttpDriver) {
		log.WithContext(ctx).Error("http driver is not supported")
		os.Exit(1)
	}

	switch inboundHttpDriver {
	case "gin":
		app := gin.Default()
		inboundHttpAdapter := gin_inbound_adapter.NewAdapter(a.domain)
		gin_inbound_adapter.InitRoute(ctx, app, inboundHttpAdapter)
		go func() {
			if err := app.Run(":" + os.Getenv("SERVER_PORT")); err != nil {
				log.WithContext(ctx).Error("failed to listen and serve", err)
				os.Exit(1)
			}
		}()
	}

	ctx, shutdown := context.WithTimeout(ctx, 5*time.Second)
	defer shutdown()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)
	<-quit

	log.WithContext(ctx).Info("http server stopped")
}

func (a *App) messageInbound() {
	ctx := a.ctx
	if !utils.IsInList(messageDriverList, inboundMessageDriver) {
		log.WithContext(ctx).Error("message driver is not supported")
		os.Exit(1)
	}

	switch inboundMessageDriver {
	case "rabbitmq":
		inboundMessageAdapter := rabbitmq_inbound_adapter.NewAdapter(a.domain)
		rabbitmq_inbound_adapter.InitRoute(ctx, os.Args, inboundMessageAdapter)
	}
}

func (a *App) commandInbound() {
	ctx := a.ctx
	inboundCommandAdapter := command_inbound_adapter.NewAdapter(a.domain)
	command_inbound_adapter.InitRoute(ctx, os.Args, inboundCommandAdapter)
}

func (a *App) workflowInbound() {
	ctx := a.ctx
	if !utils.IsInList(workflowDriverList, inboundWorkflowDriver) {
		log.WithContext(ctx).Error("workflow driver is not supported")
		os.Exit(1)
	}

	switch inboundWorkflowDriver {
	case "temporal":
		inboundWorkflowAdapter := temporal_inbound_adapter.NewAdapter(a.domain)
		temporal_inbound_adapter.InitRoute(ctx, os.Args, inboundWorkflowAdapter)
	}
}

func configureLogging() {
	// Zap logger is initialized in log package init()
	// No additional configuration needed here; it auto-detects APP_MODE
	defer log.Sync()
}
