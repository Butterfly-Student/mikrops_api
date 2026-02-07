package internal

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	joonix "github.com/joonix/log"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	command_inbound_adapter "mikrops/internal/adapter/inbound/command"
	gin_inbound_adapter "mikrops/internal/adapter/inbound/gin"
	rabbitmq_inbound_adapter "mikrops/internal/adapter/inbound/rabbitmq"
	temporal_inbound_adapter "mikrops/internal/adapter/inbound/temporal"
	postgres_outbound_adapter "mikrops/internal/adapter/outbound/postgres"
	rabbitmq_outbound_adapter "mikrops/internal/adapter/outbound/rabbitmq"
	redis_outbound_adapter "mikrops/internal/adapter/outbound/redis"
	temporal_outbound_adapter "mikrops/internal/adapter/outbound/temporal"
	"mikrops/internal/domain"
	_ "mikrops/internal/migration/postgres"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils"
	"mikrops/utils/activity"
	"mikrops/utils/database"
	"mikrops/utils/log"
	"mikrops/utils/rabbitmq"
	"mikrops/utils/redis"
)

var databaseDriverList = []string{"postgres"}
var httpDriverList = []string{"gin"}
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
	domain := domain.NewDomain(
		databaseOutbound(ctx),
		messageOutbound(ctx),
		cacheOutbound(ctx),
		workflowOutbound(ctx),
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

func databaseOutbound(ctx context.Context) outbound_port.DatabasePort {
	if !utils.IsInList(databaseDriverList, outboundDatabaseDriver) {
		log.WithContext(ctx).Fatal("database driver is not supported")
		os.Exit(1)
	}

	switch outboundDatabaseDriver {
	case "postgres":
		db := database.InitGormDatabase(ctx, outboundDatabaseDriver)
		return postgres_outbound_adapter.NewAdapter(db)
	}
	return nil
}

func messageOutbound(ctx context.Context) outbound_port.MessagePort {
	if !utils.IsInList(messageDriverList, outboundMessageDriver) {
		log.WithContext(ctx).Fatal("message driver is not supported")
		os.Exit(1)
	}

	switch outboundMessageDriver {
	case "rabbitmq":
		if err := rabbitmq.InitMessage(); err != nil {
			log.WithContext(ctx).Fatalf("failed to init rabbitmq: %v", err)
		}
		return rabbitmq_outbound_adapter.NewAdapter()
	}
	return nil
}

func cacheOutbound(ctx context.Context) outbound_port.CachePort {
	if !utils.IsInList([]string{"redis"}, outboundCacheDriver) {
		log.WithContext(ctx).Fatal("cache driver is not supported")
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
		log.WithContext(ctx).Fatal("workflow driver is not supported")
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
		log.WithContext(ctx).Fatal("http driver is not supported")
		os.Exit(1)
	}

	switch inboundHttpDriver {
	case "gin":
		router := gin.Default()
		inboundHttpAdapter := gin_inbound_adapter.NewAdapter(a.domain)
		gin_inbound_adapter.InitRoute(ctx, router, inboundHttpAdapter)
		go func() {
			if err := router.Run(":" + os.Getenv("SERVER_PORT")); err != nil {
				log.WithContext(ctx).Fatalf("failed to listen and serve: %+v", err)
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
		log.WithContext(ctx).Fatal("message driver is not supported")
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
		log.WithContext(ctx).Fatal("workflow driver is not supported")
		os.Exit(1)
	}

	switch inboundWorkflowDriver {
	case "temporal":
		inboundWorkflowAdapter := temporal_inbound_adapter.NewAdapter(a.domain)
		temporal_inbound_adapter.InitRoute(ctx, os.Args, inboundWorkflowAdapter)
	}
}

func configureLogging() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(utils.LogrusSourceContextHook{})

	if os.Getenv("APP_MODE") != "release" {
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	} else {
		logrus.SetFormatter(&joonix.FluentdFormatter{})
	}
}
