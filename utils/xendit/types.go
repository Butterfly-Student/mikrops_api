package xendit

// XenditStatus represents status from Xendit
const (
	StatusPending   = "PENDING"
	StatusPaid      = "PAID"
	StatusExpired   = "EXPIRED"
	StatusCancelled = "CANCELLED"
)

// PaymentMethod represents available payment methods
const (
	PaymentMethodVA       = "VA"
	PaymentMethodEwallet  = "EWALLET"
	PaymentMethodCard     = "CARD"
	PaymentMethodQRIS     = "QRIS"
	PaymentMethodRetail   = "RETAIL_OUTLET"
)

// EwalletType represents e-wallet types
const (
	EwalletOVO    = "OVO"
	EwalletGopay  = "GOPAY"
	EwalletDana   = "DANA"
	EwalletLinkAja = "LINKAJA"
	EwalletShopeePay = "SHOPEEPAY"
)

// BankCode represents bank codes for VA
const (
	BankBCA     = "BCA"
	BankBNI     = "BNI"
	BankBRI     = "BRI"
	BankMandiri = "MANDIRI"
	BankPermata = "PERMATA"
	BankCIMB    = "CIMB"
)
