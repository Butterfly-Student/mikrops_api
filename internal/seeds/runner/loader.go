package runner

import (
	"os"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

type SeedConfig struct {
	Env      Environment
	Entities []string
}

func GetEnvironment() Environment {
	env := os.Getenv("SEED_ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
		if env == "" {
			env = os.Getenv("GO_ENV")
		}
	}

	switch strings.ToLower(env) {
	case "production", "prod":
		return Production
	case "staging", "stage":
		return Staging
	case "testing", "test":
		return Testing
	case "development", "dev", "":
		return Development
	default:
		return Development
	}
}

func GetSeedConfig() SeedConfig {
	entities := os.Getenv("SEED_ENTITIES")
	var entityList []string

	if entities == "" || strings.ToLower(entities) == "all" {
		entityList = []string{"all"}
	} else {
		entityList = strings.Split(entities, ",")
		for i := range entityList {
			entityList[i] = strings.TrimSpace(entityList[i])
		}
	}

	return SeedConfig{
		Env:      GetEnvironment(),
		Entities: entityList,
	}
}

func ShouldSeedEntity(entity string, entities []string) bool {
	for _, e := range entities {
		if strings.ToLower(e) == "all" || strings.ToLower(e) == strings.ToLower(entity) {
			return true
		}
	}
	return false
}
