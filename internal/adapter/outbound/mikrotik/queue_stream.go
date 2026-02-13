package mikrotik_outbound_adapter

import (
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
			cmd := []string{"/queue/simple/print", "stats"}

			listenReply, err := client.ListenArgs(cmd)
			if err != nil {
				return
			}

			var batch []model.QueueStats
			replyChan := listenReply.Chan()

			for sen := range replyChan {
				stats := parseQueueStatsSentence(sen)
				batch = append(batch, *stats)
			}

			statsCh <- batch
			time.Sleep(1 * time.Second)
		}
	}()

	return statsCh, nil
}

func parseQueueStatsSentence(sen *proto.Sentence) *model.QueueStats {
	bytes := parseSlashPair(sen.Map["bytes"])
	packets := parseSlashPair(sen.Map["packets"])
	rate := parseSlashPair(sen.Map["rate"])
	packetRate := parseSlashPair(sen.Map["packet-rate"])

	return &model.QueueStats{
		Name:          sen.Map["name"],
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
