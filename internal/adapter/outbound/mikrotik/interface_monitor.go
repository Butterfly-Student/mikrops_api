package mikrotik_outbound_adapter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-template/internal/model"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
)

// MonitorAllInterfaces streams traffic stats for all interfaces
func (a *mikrotikClientAdapter) MonitorAllInterfaces(ctx context.Context, router *model.MikrotikRouter) (<-chan []model.InterfaceStats, error) {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router for interface monitoring: %w", err)
	}

	statsCh := make(chan []model.InterfaceStats, 10)

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
				// Execute monitor-traffic command once
				reply, err := client.Run("/interface/monitor-traffic", "=once=", "=.proplist=name,rx-bits-per-second,tx-bits-per-second,rx-packets-per-second,tx-packets-per-second,rx-drops-per-second,tx-drops-per-second,rx-errors-per-second,tx-errors-per-second,fp-rx-bits-per-second,fp-tx-bits-per-second,fp-rx-packets-per-second,fp-tx-packets-per-second")
				if err != nil {
					return
				}

				var batch []model.InterfaceStats
				for _, re := range reply.Re {
					stats := parseInterfaceStats(re)
					if stats != nil {
						batch = append(batch, *stats)
					}
				}

				if len(batch) > 0 {
					select {
					case statsCh <- batch:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return statsCh, nil
}

// MonitorInterface streams traffic stats for a specific interface by name
func (a *mikrotikClientAdapter) MonitorInterface(ctx context.Context, router *model.MikrotikRouter, interfaceName string) (<-chan model.InterfaceStats, error) {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router for interface monitoring: %w", err)
	}

	statsCh := make(chan model.InterfaceStats, 10)

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
				// Execute monitor-traffic for specific interface
				reply, err := client.Run(
					"/interface/monitor-traffic",
					"=interface="+interfaceName,
					"=once=",
					"=.proplist=name,rx-bits-per-second,tx-bits-per-second,rx-packets-per-second,tx-packets-per-second,rx-drops-per-second,tx-drops-per-second,rx-errors-per-second,tx-errors-per-second,fp-rx-bits-per-second,fp-tx-bits-per-second,fp-rx-packets-per-second,fp-tx-packets-per-second",
				)
				if err != nil {
					return
				}

				if len(reply.Re) > 0 {
					stats := parseInterfaceStats(reply.Re[0])
					if stats != nil {
						select {
						case statsCh <- *stats:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return statsCh, nil
}

// parseInterfaceStats converts RouterOS response to InterfaceStats model
func parseInterfaceStats(re *proto.Sentence) *model.InterfaceStats {
	return &model.InterfaceStats{
		Name:            re.Map["name"],
		RxBitsPerS:      parseInt64(re.Map["rx-bits-per-second"]),
		TxBitsPerS:      parseInt64(re.Map["tx-bits-per-second"]),
		RxPacketsPerS:   parseInt64(re.Map["rx-packets-per-second"]),
		TxPacketsPerS:   parseInt64(re.Map["tx-packets-per-second"]),
		RxDropsPerS:     parseInt64(re.Map["rx-drops-per-second"]),
		TxDropsPerS:     parseInt64(re.Map["tx-drops-per-second"]),
		RxErrorsPerS:    parseInt64(re.Map["rx-errors-per-second"]),
		TxErrorsPerS:    parseInt64(re.Map["tx-errors-per-second"]),
		FpRxBitsPerS:    parseInt64(re.Map["fp-rx-bits-per-second"]),
		FpTxBitsPerS:    parseInt64(re.Map["fp-tx-bits-per-second"]),
		FpRxPacketsPerS: parseInt64(re.Map["fp-rx-packets-per-second"]),
		FpTxPacketsPerS: parseInt64(re.Map["fp-tx-packets-per-second"]),
	}
}

// parseInt64 safely parses string to int64, returns 0 on error
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}
