package http_outbound_adapter

import (
	"fmt"

	"go-template/pkg/gowa"
	outbound_port "go-template/internal/port/outbound"
)

type gowaAdapter struct {
	client *gowa.Client
	err    error
}

// NewGowaAdapter creates a Gowa outbound adapter using environment variables.
// If GOWA_USERNAME or GOWA_PASSWORD are not set, the adapter is created without a client.
// Calling GetClient() on an unconfigured adapter will return an error.
func NewGowaAdapter() outbound_port.GowaPort {
	client, err := gowa.NewFromEnv()
	return &gowaAdapter{client: client, err: err}
}

func (a *gowaAdapter) GetClient() (*gowa.Client, error) {
	if a.err != nil {
		return nil, fmt.Errorf("gowa client not configured: %w", a.err)
	}
	return a.client, nil
}
