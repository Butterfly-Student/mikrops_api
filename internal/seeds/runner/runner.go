package runner

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
)

type Seeder interface {
	Seed(db *gorm.DB) error
	Name() string
}

var (
	seeders      []Seeder
	seedersMutex sync.Mutex
)

func RegisterSeeder(seeder Seeder) {
	seedersMutex.Lock()
	defer seedersMutex.Unlock()
	seeders = append(seeders, seeder)
}

func GetRegisteredSeeders() []Seeder {
	seedersMutex.Lock()
	defer seedersMutex.Unlock()

	result := make([]Seeder, len(seeders))
	copy(result, seeders)
	return result
}

type SeedResult struct {
	Name    string
	Success bool
	Error   error
}

func Run(db *gorm.DB, seeders []Seeder) ([]SeedResult, error) {
	var results []SeedResult
	var hasError bool

	for _, seeder := range seeders {
		log.Printf("[SEED] Running: %s", seeder.Name())

		result := SeedResult{
			Name: seeder.Name(),
		}

		if err := seeder.Seed(db); err != nil {
			result.Success = false
			result.Error = err
			hasError = true
			log.Printf("[SEED] FAILED: %s - Error: %v", seeder.Name(), err)
		} else {
			result.Success = true
			log.Printf("[SEED] SUCCESS: %s", seeder.Name())
		}

		results = append(results, result)
	}

	if hasError {
		return results, fmt.Errorf("one or more seeders failed")
	}

	return results, nil
}

func RunInTransaction(db *gorm.DB, seeders []Seeder) ([]SeedResult, error) {
	var results []SeedResult

	err := db.Transaction(func(tx *gorm.DB) error {
		innerResults, err := Run(tx, seeders)
		results = innerResults
		return err
	})

	return results, err
}
