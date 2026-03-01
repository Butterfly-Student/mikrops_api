package inbound_port

// MikrotikSyncMessagePort defines the interface for consuming MikroTik sync messages
type MikrotikSyncMessagePort interface {
	// ProcessPPPoESync processes PPPoE sync messages from the queue
	// Returns true if message should be acknowledged, false for retry/nack
	ProcessPPPoESync(data []byte) bool
	
	// ProcessHotspotSync processes Hotspot sync messages from the queue
	// Returns true if message should be acknowledged, false for retry/nack
	ProcessHotspotSync(data []byte) bool
}
