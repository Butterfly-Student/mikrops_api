package fixtures

import (
	"fmt"

	"go-template/internal/model"

	"github.com/google/uuid"
)

type PppoeTestData struct{}

func NewPppoeTestData() *PppoeTestData {
	return &PppoeTestData{}
}

// ValidPppoeSecret returns a valid PPPoE secret for testing
func (p *PppoeTestData) ValidPppoeSecret() model.PppoeSecret {
	return model.PppoeSecret{
		ID:            "1",
		Name:          "test-secret",
		Password:      "secret-password",
		Service:       "pppoe",
		CallerID:      "AA:BB:CC:DD:EE:FF",
		Profile:       "default",
		LocalAddress:  "192.168.1.1",
		RemoteAddress: "192.168.1.2",
		Routes:        "0.0.0.0/0",
		Comment:       "Test PPPoE Secret",
		Disabled:      false,
	}
}

// ValidPppoeProfile returns a valid PPPoE profile for testing
func (p *PppoeTestData) ValidPppoeProfile() model.PppoeProfile {
	return model.PppoeProfile{
		ID:            "1",
		Name:          "default",
		LocalAddress:  "192.168.1.1",
		RemoteAddress: "192.168.1.0/24",
		RateLimit:     "10M/20M",
		OnlyOne:       "default",
		DNSServer:     "8.8.8.8",
		Comment:       "Default PPPoE Profile",
	}
}

// ValidPppoeActive returns a valid active PPPoE session for testing
func (p *PppoeTestData) ValidPppoeActive() model.PppoeActive {
	return model.PppoeActive{
		ID:            "1",
		Name:          "test-user",
		Service:       "pppoe",
		CallerID:      "AA:BB:CC:DD:EE:FF",
		Address:       "192.168.1.100",
		Uptime:        "00:05:30",
		Encoding:      "CHAP",
		SessionID:     "1234567890",
		LimitBytesIn:  1000000000,
		LimitBytesOut: 500000000,
		Radius:        false,
	}
}

// ValidPppoeCallbackData returns a valid PPPoE callback data for testing
func (p *PppoeTestData) ValidPppoeCallbackData() model.PppoeCallbackData {
	return model.PppoeCallbackData{
		User:       "test-user",
		IP:         "192.168.1.100",
		CallerID:   "AA:BB:CC:DD:EE:FF",
		SessionID:  "1234567890",
		Interface:  "pppoe-out0",
		Uptime:     "00:05:30",
		BytesIn:    1024000,
		BytesOut:   512000,
		PacketsIn:  10000,
		PacketsOut: 5000,
		RouterID:   1,
	}
}

// MultiplePppoeSecrets returns multiple PPPoE secrets for testing
func (p *PppoeTestData) MultiplePppoeSecrets(count int) []model.PppoeSecret {
	secrets := make([]model.PppoeSecret, count)

	for i := 0; i < count; i++ {
		secrets[i] = model.PppoeSecret{
			ID:            fmt.Sprintf("%d", i+1),
			Name:          fmt.Sprintf("secret-%d", i+1),
			Password:      fmt.Sprintf("password-%d", i+1),
			Service:       "pppoe",
			CallerID:      fmt.Sprintf("AA:BB:CC:DD:EE:%02X", i+1),
			Profile:       "default",
			LocalAddress:  "192.168.1.1",
			RemoteAddress: fmt.Sprintf("192.168.1.%d", i+2),
			Routes:        "0.0.0.0/0",
			Comment:       fmt.Sprintf("PPPoE Secret %d", i+1),
			Disabled:      false,
		}
	}

	return secrets
}

// MultiplePppoeProfiles returns multiple PPPoE profiles for testing
func (p *PppoeTestData) MultiplePppoeProfiles(count int) []model.PppoeProfile {
	profiles := make([]model.PppoeProfile, count)

	for i := 0; i < count; i++ {
		profiles[i] = model.PppoeProfile{
			ID:            fmt.Sprintf("%d", i+1),
			Name:          fmt.Sprintf("profile-%d", i+1),
			LocalAddress:  "192.168.1.1",
			RemoteAddress: fmt.Sprintf("192.168.%d.0/24", i+1),
			RateLimit:     fmt.Sprintf("%dM/%dM", (i+1)*5, (i+1)*10),
			OnlyOne:       "default",
			DNSServer:     "8.8.8.8",
			Comment:       fmt.Sprintf("PPPoE Profile %d", i+1),
		}
	}

	return profiles
}

// MultiplePppoeActive returns multiple active PPPoE sessions for testing
func (p *PppoeTestData) MultiplePppoeActive(count int) []model.PppoeActive {
	sessions := make([]model.PppoeActive, count)

	for i := 0; i < count; i++ {
		sessions[i] = model.PppoeActive{
			ID:            fmt.Sprintf("%d", i+1),
			Name:          fmt.Sprintf("user-%d", i+1),
			Service:       "pppoe",
			CallerID:      fmt.Sprintf("AA:BB:CC:DD:EE:%02X", i+1),
			Address:       fmt.Sprintf("192.168.1.%d", i+100),
			Uptime:        "00:05:30",
			Encoding:      "CHAP",
			SessionID:     fmt.Sprintf("SESSION-%d", i+1),
			LimitBytesIn:  1000000000,
			LimitBytesOut: 500000000,
			Radius:        i%2 == 0,
		}
	}

	return sessions
}

// PppoeSecretWithName returns a secret with specific name
func (p *PppoeTestData) PppoeSecretWithName(name string) model.PppoeSecret {
	secret := p.ValidPppoeSecret()
	secret.Name = name
	return secret
}

// PppoeProfileWithName returns a profile with specific name
func (p *PppoeTestData) PppoeProfileWithName(name string) model.PppoeProfile {
	profile := p.ValidPppoeProfile()
	profile.Name = name
	return profile
}

// PppoeActiveWithUser returns an active session for specific user
func (p *PppoeTestData) PppoeActiveWithUser(username string) model.PppoeActive {
	active := p.ValidPppoeActive()
	active.Name = username
	return active
}

// PppoeCallbackDataWithUser returns callback data for specific user
func (p *PppoeTestData) PppoeCallbackDataWithUser(username string) model.PppoeCallbackData {
	data := p.ValidPppoeCallbackData()
	data.User = username
	data.SessionID = uuid.New().String()
	return data
}

// WebSocketMessageSessionUp returns a session up message
func (p *PppoeTestData) WebSocketMessageSessionUp(username string) model.WebSocketMessage {
	data := p.PppoeCallbackDataWithUser(username)
	return model.WebSocketMessage{
		Event: "session_up",
		Data:  data,
	}
}

// WebSocketMessageSessionDown returns a session down message
func (p *PppoeTestData) WebSocketMessageSessionDown(username string) model.WebSocketMessage {
	data := p.PppoeCallbackDataWithUser(username)
	return model.WebSocketMessage{
		Event: "session_down",
		Data:  data,
	}
}
