package model

import "go-template/pkg/hotspot"

// Hotspot type aliases — all hotspot types come from pkg/hotspot.
// Using type aliases means model.HotspotProfile and hotspot.Profile are identical types.
type (
	HotspotProfile         = hotspot.Profile
	HotspotProfileUpdate   = hotspot.ProfileUpdate
	HotspotUser            = hotspot.User
	HotspotUserUpdate      = hotspot.UserUpdate
	HotspotUserFilter      = hotspot.UserFilter
	HotspotSession         = hotspot.Session
	HotspotSessionStats    = hotspot.SessionStats
	HotspotSale            = hotspot.Sale
	HotspotSaleFilter      = hotspot.SaleFilter
	HotspotVoucherGenerator = hotspot.VoucherGenerator
	HotspotVoucherResult   = hotspot.VoucherResult
	HotspotScheduler       = hotspot.Scheduler
)
