package outbound_port

import (
	"go-template/internal/model"
	"go-template/pkg/hotspot"
)

// HotspotPort provides a hotspot client connected to a specific MikroTik router.
// The returned client wraps pkg/hotspot and communicates directly with RouterOS.
type HotspotPort interface {
	GetHotspotClient(router *model.MikrotikRouter) (hotspot.Client, error)
}
