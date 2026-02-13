package mikrotik_outbound_adapter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3"
)

// Streaming Stats

func (a *mikrotikClientAdapter) ListenQueueStats(router *model.MikrotikRouter) (<-chan []model.QueueStats, error) {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router for streaming: %w", err)
	}

	statsCh := make(chan []model.QueueStats)

	go func() {
		defer client.Close()
		defer close(statsCh)

		for {
			// Correct usage of ListenArgs: returns (chan *Reply, error)
			// It accepts variadic string args
			// v3.0.1: func (c *Client) ListenArgs(args ...string) (chan *Reply, error)
			// Wait, if it returns 2 values, where is the error channel?
			// The channel returns *Reply. Error is returned immediately if start fails.
			// Async errors might come in the channel?
			// Checking common patterns: ListenArgs returns a channel that emits replies.
			// When command finishes, channel closes? Or keeps open?

			// Command: /queue/simple/print stats
			// Args: "/queue/simple/print", "stats" -> passed as individual strings

			// Note: "stats" is a parameter without value, usually passed as "stats" in CLI, but in API often just "stats".
			// Or "=stats=". Let's try "stats".

			replyChan, err := client.ListenArgs("/queue/simple/print", "stats")
			if err != nil {
				return
			}

			var batch []model.QueueStats

			for re := range replyChan {
				if re.Done {
					// Command finished
					statsCh <- batch
					batch = nil
					break
				}
				// Reply sentence
				batch = append(batch, *parseQueueStats(re))
			}

			// Poll interval
			time.Sleep(1 * time.Second)
		}
	}()

	return statsCh, nil
}

func parseQueueStats(re *routeros.Reply) *model.QueueStats {
	bytes := parseSlashPair(re.Map["bytes"])
	packets := parseSlashPair(re.Map["packets"])
	rate := parseSlashPair(re.Map["rate"])
	packetRate := parseSlashPair(re.Map["packet-rate"])

	return &model.QueueStats{
		Name:          re.Map["name"],
		BytesIn:       bytes[0],
		BytesOut:      bytes[1],
		PacketsIn:     packets[0],
		PacketsOut:    packets[1],
		RateIn:        rate[0],
		RateOut:       rate[1],
		PacketRateIn:  packetRate[0],
		PacketRateOut: packetRate[1],
	}
}

func parseSlashPair(s string) [2]int64 {
	var res [2]int64
	if s == "" {
		return res
	}
	parts := strings.Split(s, "/")
	if len(parts) >= 1 {
		res[0], _ = strconv.ParseInt(parts[0], 10, 64)
	}
	if len(parts) >= 2 {
		res[1], _ = strconv.ParseInt(parts[1], 10, 64)
	}
	return res
}
