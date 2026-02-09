package xendit_adapter

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/palantir/stacktrace"
	xendit "github.com/xendit/xendit-go/v6"
	invoice "github.com/xendit/xendit-go/v6/invoice"
	payment_request "github.com/xendit/xendit-go/v6/payment_request"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct {
	client *xendit.APIClient
}

func NewAdapter() outbound_port.XenditPort {
	apiKey := os.Getenv("XENDIT_API_KEY")
	client := xendit.NewClient(apiKey)
	return &adapter{
		client: client,
	}
}

func (a *adapter) CreateInvoice(ctx context.Context, inv model.Invoice, customer model.Customer) (string, string, error) {
	amount := float64(inv.TotalAmount)
	currency := "IDR"
	description := fmt.Sprintf("Invoice #%s", inv.InvoiceNumber)

	// Calculate duration
	invoiceDuration := "172800" // Default 2 days
	if !inv.DueDate.IsZero() {
		duration := time.Until(inv.DueDate)
		seconds := int64(duration.Seconds())
		if seconds < 0 {
			seconds = 86400 // Default to 1 day if overdue
		}
		invoiceDuration = strconv.FormatInt(seconds, 10)
	}

	// Prepare invoice request
	createInvoiceRequest := invoice.CreateInvoiceRequest{
		ExternalId:      inv.InvoiceNumber,
		Amount:          amount,
		Description:     &description,
		InvoiceDuration: &invoiceDuration,
		Customer: &invoice.CustomerObject{
			GivenNames:   *invoice.NewNullableString(&customer.FullName),
			Email:        *invoice.NewNullableString(&customer.Email),
			MobileNumber: *invoice.NewNullableString(&customer.Phone),
		},
		Currency: &currency,
	}

	resp, _, err := a.client.InvoiceApi.CreateInvoice(ctx).CreateInvoiceRequest(createInvoiceRequest).Execute()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "failed to create xendit invoice")
	}

	return *resp.Id, resp.InvoiceUrl, nil
}

func (a *adapter) CreateVirtualAccount(ctx context.Context, inv model.Invoice, customer model.Customer, bankCode string) (string, error) {
	amount := float64(inv.TotalAmount)
	currency := payment_request.PAYMENTREQUESTCURRENCY_IDR

	// Channel Code
	channelCode := payment_request.VirtualAccountChannelCode(bankCode)

	valOneTime := payment_request.PAYMENTMETHODREUSABILITY_ONE_TIME_USE

	prRequest := payment_request.PaymentRequestParameters{
		ReferenceId: &inv.InvoiceNumber,
		Amount:      &amount,
		Currency:    currency,
		PaymentMethod: &payment_request.PaymentMethodParameters{
			Type:        "VIRTUAL_ACCOUNT",
			Reusability: valOneTime,
			VirtualAccount: *payment_request.NewNullableVirtualAccountParameters(&payment_request.VirtualAccountParameters{
				ChannelCode: channelCode,
				ChannelProperties: payment_request.VirtualAccountChannelProperties{
					CustomerName: customer.FullName,
					ExpiresAt:    &inv.DueDate,
				},
			}),
		},
	}

	resp, _, err := a.client.PaymentRequestApi.CreatePaymentRequest(ctx).PaymentRequestParameters(prRequest).Execute()
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to create payment request VA")
	}

	if resp.PaymentMethod.VirtualAccount.IsSet() {
		va := resp.PaymentMethod.VirtualAccount.Get()
		if va != nil {
			// va.ChannelProperties is likely just a struct or pointer, not Nullable, based on previous error
			// Error: "type ... has no field Get".
			// Let's assume it's just a struct field `ChannelProperties`
			props := va.ChannelProperties
			if props.VirtualAccountNumber != nil {
				return *props.VirtualAccountNumber, nil
			}
		}
	}

	return "", stacktrace.NewError("VA number not found in response")
}

func (a *adapter) GetInvoiceStatus(ctx context.Context, externalID string) (string, error) {
	// Get invoice by ID? But we have ExternalID.
	// Xendit Get Invoice usually supports ID. If ExternalID, we might need to List with filter.
	// However, usually we store Xendit ID (the 'id' returned from create).
	// The prompt says GetInvoiceStatus(externalID string).
	// If externalID refers to our DB ID, we should have stored the Xendit ID.
	// But if implementation requires Xendit ID, I will assume the `externalID` passed here IS the Xendit ID.
	// Or if it IS the client's external ID, we need to use `GetInvoices` with external_id filter.

	// Let's assume `externalID` passed to this function is the Invoice Number (Client External ID).

	resp, _, err := a.client.InvoiceApi.GetInvoices(ctx).ExternalId(externalID).Execute()
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to get invoice status")
	}

	if len(resp) == 0 {
		return "", stacktrace.NewError("invoice not found")
	}

	// Return the status of the first match (should be unique per active invoice usually)
	return string(resp[0].Status), nil
}
