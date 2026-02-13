package postgres_outbound_adapter

import (
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

func InitCasbin(db *gorm.DB) *casbin.Enforcer {
	// Initialize Gorm adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("failed to initialize casbin adapter: %v", err)
	}

	// Define RBAC model
	text := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(text)
	if err != nil {
		log.Fatalf("failed to create casbin model: %v", err)
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatalf("failed to create casbin enforcer: %v", err)
	}

	// Load policies from DB
	if err := e.LoadPolicy(); err != nil {
		log.Fatalf("failed to load policies: %v", err)
	}

	return e
}
