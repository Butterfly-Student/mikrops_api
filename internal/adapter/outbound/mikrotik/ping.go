package mikrotik_outbound_adapter

import (
	"context"
	"fmt"
	"strconv"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

// Ping executes ping command with streaming results
func (a *mikrotikClientAdapter) Ping(ctx context.Context, router *model.MikrotikRouter, req model.PingRequest) (<-chan model.PingResult, error) {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router for ping: %w", err)
	}

	resultCh := make(chan model.PingResult, 10)

	go func() {
		defer client.Close()
		defer close(resultCh)

		// Build ping command
		cmd := []string{"/ping", "=address=" + req.Address}

		// Add optional parameters
		if req.Count > 0 {
			cmd = append(cmd, fmt.Sprintf("=count=%d", req.Count))
		} else {
			// Continuous ping - set high count
			cmd = append(cmd, "=count=999999")
		}

		if req.Size > 0 {
			cmd = append(cmd, fmt.Sprintf("=size=%d", req.Size))
		}

		if req.Interval != "" {
			cmd = append(cmd, "=interval="+req.Interval)
		}

		// Execute ping with streaming
		listenReply, err := client.ListenArgs(cmd)
		if err != nil {
			return
		}

		replyChan := listenReply.Chan()
		seq := 0

		for {
			select {
			case <-ctx.Done():
				return
			case re, ok := <-replyChan:
				if !ok {
					return
				}

				result := parsePingResult(re, req.Address, seq)
				seq++

				select {
				case resultCh <- *result:
				case <-ctx.Done():
					return
				}

				// If count is set and reached, stop
				if req.Count > 0 && seq >= req.Count {
					return
				}
			}
		}
	}()

	return resultCh, nil
}

// parsePingResult converts RouterOS ping response to PingResult model
func parsePingResult(re *proto.Sentence, host string, seq int) *model.PingResult {
	result := &model.PingResult{
		Host: host,
		Seq:  seq,
	}

	// Parse size
	if size := re.Map["size"]; size != "" {
		result.Size, _ = strconv.Atoi(size)
	}

	// Parse TTL
	if ttl := re.Map["ttl"]; ttl != "" {
		result.TTL, _ = strconv.Atoi(ttl)
	}

	// Parse time
	result.Time = re.Map["time"]

	// Parse sent/received time
	result.SentTime = re.Map["sent"]
	result.ReceivedTime = re.Map["received"]

	// Determine status
	if re.Map["timeout"] != "" {
		result.Status = "timeout"
	} else if re.Map["time"] != "" {
		result.Status = "success"
	} else {
		result.Status = "error"
	}

	return result
}
