package outbound_port

import "go-template/pkg/gowa"

// GowaPort provides access to the Gowa WhatsApp gateway client.
type GowaPort interface {
	GetClient() (*gowa.Client, error)
}
