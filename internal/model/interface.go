package model

// InterfaceStats represents real-time network interface statistics from MikroTik
type InterfaceStats struct {
	Name            string `json:"name"`
	RxBitsPerS      int64  `json:"rx-bits-per-second"`
	TxBitsPerS      int64  `json:"tx-bits-per-second"`
	RxPacketsPerS   int64  `json:"rx-packets-per-second"`
	TxPacketsPerS   int64  `json:"tx-packets-per-second"`
	RxDropsPerS     int64  `json:"rx-drops-per-second"`
	TxDropsPerS     int64  `json:"tx-drops-per-second"`
	RxErrorsPerS    int64  `json:"rx-errors-per-second"`
	TxErrorsPerS    int64  `json:"tx-errors-per-second"`
	FpRxBitsPerS    int64  `json:"fp-rx-bits-per-second"`
	FpTxBitsPerS    int64  `json:"fp-tx-bits-per-second"`
	FpRxPacketsPerS int64  `json:"fp-rx-packets-per-second"`
	FpTxPacketsPerS int64  `json:"fp-tx-packets-per-second"`
}
