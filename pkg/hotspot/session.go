package hotspot

import (
	"context"
	"fmt"

	"github.com/go-routeros/routeros/v3/proto"
	"go-template/pkg/hotspot/internal"
)

// GetActiveSessions retrieves all active hotspot sessions
func (c *hotspotClient) GetActiveSessions(ctx context.Context) ([]Session, error) {
	reply, err := c.execute(ctx, PathHotspotActive+"/print")
	if err != nil {
		return nil, WrapError("get active sessions", err)
	}

	sessions := make([]Session, 0, len(reply.Re))
	for _, re := range reply.Re {
		sessions = append(sessions, c.mapReplyToSession(re))
	}

	return sessions, nil
}

// GetSessionsByServer retrieves active sessions for specific server
func (c *hotspotClient) GetSessionsByServer(ctx context.Context, server string) ([]Session, error) {
	if server == "" {
		server = DefaultServer
	}

	reply, err := c.execute(ctx, PathHotspotActive+"/print", "?server="+server)
	if err != nil {
		return nil, WrapError("get sessions by server", err)
	}

	sessions := make([]Session, 0, len(reply.Re))
	for _, re := range reply.Re {
		sessions = append(sessions, c.mapReplyToSession(re))
	}

	return sessions, nil
}

// GetSessionByUsername retrieves active session by username
func (c *hotspotClient) GetSessionByUsername(ctx context.Context, username string) (*Session, error) {
	if username == "" {
		return nil, NewError("get session", fmt.Errorf("username is required"))
	}

	reply, err := c.execute(ctx, PathHotspotActive+"/print", "?user="+username)
	if err != nil {
		return nil, WrapError("get session by username", err)
	}

	if len(reply.Re) == 0 {
		return nil, ErrSessionNotFound
	}

	session := c.mapReplyToSession(reply.Re[0])
	return &session, nil
}

// DisconnectUser disconnects active user session
func (c *hotspotClient) DisconnectUser(ctx context.Context, username string) error {
	if username == "" {
		return NewError("disconnect user", fmt.Errorf("username is required"))
	}

	// Find user in active sessions
	reply, err := c.execute(ctx, PathHotspotActive+"/print", "?user="+username)
	if err != nil {
		return WrapError("disconnect user", err)
	}

	if len(reply.Re) == 0 {
		return ErrSessionNotFound
	}

	sessionID := reply.Re[0].Map[".id"]

	_, err = c.execute(ctx, PathHotspotActive+"/remove", "=.id="+sessionID)
	if err != nil {
		return WrapError("disconnect user", err)
	}

	return nil
}

// GetSessionStats retrieves aggregated session statistics.
// TotalUsers is the count of all registered hotspot users (/ip/hotspot/user).
// ActiveUsers is the count of currently connected sessions (/ip/hotspot/active).
func (c *hotspotClient) GetSessionStats(ctx context.Context) (*SessionStats, error) {
	// Count total registered users
	totalReply, err := c.execute(ctx, PathHotspotUser+"/print")
	if err != nil {
		return nil, WrapError("get session stats total users", err)
	}

	// Get active sessions for traffic stats
	activeReply, err := c.execute(ctx, PathHotspotActive+"/print")
	if err != nil {
		return nil, WrapError("get session stats active sessions", err)
	}

	totalBytesIn := int64(0)
	totalBytesOut := int64(0)

	for _, re := range activeReply.Re {
		totalBytesIn += internal.ParseReplyToInt64(re.Map["bytes-in"])
		totalBytesOut += internal.ParseReplyToInt64(re.Map["bytes-out"])
	}

	return &SessionStats{
		TotalUsers:    len(totalReply.Re),
		ActiveUsers:   len(activeReply.Re),
		TotalBytesIn:  internal.FormatBytes(totalBytesIn),
		TotalBytesOut: internal.FormatBytes(totalBytesOut),
	}, nil
}

func (c *hotspotClient) mapReplyToSession(re *proto.Sentence) Session {
	return Session{
		Name:            re.Map["user"],
		Address:         re.Map["address"],
		MacAddress:      re.Map["mac-address"],
		Uptime:          re.Map["uptime"],
		SessionTimeLeft: re.Map["session-time-left"],
		BytesIn:         re.Map["bytes-in"],
		BytesOut:        re.Map["bytes-out"],
		LoginBy:         re.Map["login-by"],
	}
}
