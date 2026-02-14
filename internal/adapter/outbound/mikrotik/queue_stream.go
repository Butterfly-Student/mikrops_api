package mikrotik_outbound_adapter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
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
			// ListenArgs accepts sentence []string
			listenReply, err := client.ListenArgs([]string{"/queue/simple/print", "stats"})
			if err != nil {
				return
			}

			var batch []model.QueueStats
			replyChan := listenReply.Chan()

			for re := range replyChan {
				// Collect all reply sentences
				// Channel closes when command completes
				batch = append(batch, *parseQueueStats(re))
			}

			// Send batch when channel closes
			statsCh <- batch

			// Poll interval
			time.Sleep(1 * time.Second)
		}
	}()

	return statsCh, nil
}

func parseQueueStats(re *proto.Sentence) *model.QueueStats {
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

// ListenQueueStatsWithContext streams all queue stats with context support
func (a *mikrotikClientAdapter) ListenQueueStatsWithContext(ctx context.Context, router *model.MikrotikRouter) (<-chan []model.QueueStats, error) {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router for streaming: %w", err)
	}

	statsCh := make(chan []model.QueueStats, 10)

	go func() {
		defer client.Close()
		defer close(statsCh)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				listenReply, err := client.ListenArgs([]string{"/queue/simple/print", "stats"})
				if err != nil {
					return
				}

				var batch []model.QueueStats
				replyChan := listenReply.Chan()

				for re := range replyChan {
					batch = append(batch, *parseQueueStats(re))
				}

				select {
				case statsCh <- batch:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return statsCh, nil
}

// ListenQueueStatsByName streams specific queue stats by name with context support
func (a *mikrotikClientAdapter) ListenQueueStatsByName(ctx context.Context, router *model.MikrotikRouter, queueName string) (<-chan model.QueueStats, error) {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router for streaming: %w", err)
	}

	statsCh := make(chan model.QueueStats, 10)

	go func() {
		defer client.Close()
		defer close(statsCh)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Filter by queue name
				listenReply, err := client.ListenArgs([]string{
					"/queue/simple/print",
					"stats",
					fmt.Sprintf("?name=%s", queueName),
				})
				if err != nil {
					return
				}

				replyChan := listenReply.Chan()

				for re := range replyChan {
					stats := parseQueueStats(re)
					select {
					case statsCh <- *stats:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return statsCh, nil
}
