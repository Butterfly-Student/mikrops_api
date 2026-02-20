package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go-template/internal/seeds"
	"go-template/utils"
	"go-template/utils/database"

	"gorm.io/gorm"
)

func main() {
	// Load .env file so DATABASE_* variables are available even when
	// running via `make seed` without manually exporting environment variables.
	if err := utils.LoadEnvFile(".env"); err != nil {
		log.Printf("[WARN] Could not load .env file: %v", err)
	}

	clean := flag.Bool("clean", false, "Clean seed data instead of seeding")
	flag.Parse()

	ctx := context.Background()
	db := database.InitDatabase(ctx, "postgres")

	if *clean {
		if err := cleanSeedData(db); err != nil {
			log.Fatalf("[ERROR] Failed to clean seed data: %v", err)
		}
		fmt.Println("[SUCCESS] Seed data cleaned successfully")
		return
	}

	if err := seeds.Run(db); err != nil {
		log.Fatalf("[ERROR] Failed to seed data: %v", err)
	}

	fmt.Println("[SUCCESS] Seed data loaded successfully")
}

func cleanSeedData(db *gorm.DB) error {
	fmt.Println("[CLEAN] Cleaning seed data...")

	tables := []string{
		"casbin_rule",
		"mikrotik_routers",
		"users",
		"clients",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			log.Printf("[WARNING] Failed to clean table %s: %v", table, err)
		} else {
			fmt.Printf("[CLEAN] Cleaned table: %s\n", table)
		}
	}

	if err := db.Exec("ALTER SEQUENCE casbin_rule_id_seq RESTART WITH 1").Error; err == nil {
		fmt.Println("[CLEAN] Reset casbin_rule_id_seq sequence")
	}

	return nil
}
