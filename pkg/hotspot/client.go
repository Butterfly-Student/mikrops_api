package hotspot

import (
	"context"
	"fmt"

	routeros "github.com/go-routeros/routeros/v3"
)

// hotspotClient is the concrete implementation of the Client interface
type hotspotClient struct {
	routerID uint
	client   RouterOSClient
	config   *Config
}

// NewClient creates a new hotspot client for managing MikroTik hotspot operations.
// It requires a router ID and an existing RouterOS client connection.
//
// Example:
//
//	client := hotspot.NewClient(routerID, mikrotikClient)
func NewClient(routerID uint, mikrotikClient RouterOSClient) Client {
	return &hotspotClient{
		routerID: routerID,
		client:   mikrotikClient,
		config:   &Config{},
	}
}

// NewClientWithConfig creates a new hotspot client with custom configuration.
// Use this when you need to set specific options like hotspot name, currency, etc.
//
// Example:
//
//	config := &hotspot.Config{
//	    RouterID: routerID,
//	    HotspotName: "hs1",
//	    Currency: "IDR",
//	    Debug: true,
//	}
//	client := hotspot.NewClientWithConfig(mikrotikClient, config)
func NewClientWithConfig(mikrotikClient RouterOSClient, config *Config) Client {
	return &hotspotClient{
		routerID: config.RouterID,
		client:   mikrotikClient,
		config:   config,
	}
}

// GetRouterID returns the router ID this client is connected to
func (c *hotspotClient) GetRouterID() uint {
	return c.routerID
}

// GetConfig returns the client configuration
func (c *hotspotClient) GetConfig() *Config {
	return c.config
}

// SetConfig updates the client configuration
func (c *hotspotClient) SetConfig(config *Config) {
	c.config = config
}

// Close closes the underlying RouterOS connection
func (c *hotspotClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// execute runs a RouterOS command with the given arguments.
// It wraps the command execution with error handling and context support.
func (c *hotspotClient) execute(ctx context.Context, cmd string, args ...string) (*routeros.Reply, error) {
	fullArgs := append([]string{cmd}, args...)

	// Check context before executing
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("operation canceled: %w", ctx.Err())
		default:
		}
	}

	reply, err := c.client.Run(fullArgs...)
	if err != nil {
		return nil, fmt.Errorf("command failed: %s %v: %w", cmd, args, err)
	}

	return reply, nil
}

// findResourceID executes a print command and returns the .id field of the first matching result
func (c *hotspotClient) findResourceID(ctx context.Context, path string, filterKey, filterValue string) (string, error) {
	filter := fmt.Sprintf("?%s=%s", filterKey, filterValue)
	reply, err := c.execute(ctx, path+"/print", filter)
	if err != nil {
		return "", WrapError("find resource", err)
	}

	if len(reply.Re) == 0 {
		return "", ErrUserNotFound // Generic not found, should be replaced by caller
	}

	return reply.Re[0].Map[".id"], nil
}

// buildCommandArgs constructs RouterOS command arguments from a map
func buildCommandArgs(baseID string, updates map[string]string) []string {
	args := []string{"=.id=" + baseID}
	for key, value := range updates {
		args = append(args, "="+key+"="+value)
	}
	return args
}
