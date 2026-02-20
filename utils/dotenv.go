package utils

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnvFile reads a .env file and sets environment variables that are not already set.
// Lines starting with '#' and empty lines are ignored.
// Variables already set in the environment take precedence (won't be overwritten).
func LoadEnvFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		// If file doesn't exist, silently skip
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '=' only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if any
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		// Only set if not already defined in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
