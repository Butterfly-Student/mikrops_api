package hotspot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-template/pkg/hotspot/internal"
)

// RecordSale creates sales record as RouterOS system script
func (c *hotspotClient) RecordSale(ctx context.Context, sale *Sale) error {
	// Validate
	if sale.Username == "" || sale.Price <= 0 {
		return NewError("record sale", fmt.Errorf("username and price required"))
	}

	// Set default values
	if sale.Date == "" {
		sale.Date = time.Now().Format("Jan/02/2006")
	}
	if sale.Time == "" {
		sale.Time = time.Now().Format("15:04:05")
	}

	// Build script name
	scriptName := internal.BuildSaleScriptName(
		sale.Date,
		sale.Time,
		sale.Username,
		sale.Price,
		sale.Address,
		sale.Mac,
		sale.Validity,
	)

	_, err := c.execute(ctx, PathSystemScript+"/add",
		"=name="+scriptName,
		"=owner="+ScriptOwnerSales,
		"=policy="+PolicyReadWrite,
	)

	if err != nil {
		return WrapError("record sale", err)
	}

	return nil
}

// GetAllSales retrieves all sales records from RouterOS.
// Sales are stored as system scripts with owner=hotspot-sales.
func (c *hotspotClient) GetAllSales(ctx context.Context, filter *SaleFilter) ([]Sale, error) {
	// Filter by owner to only retrieve sales scripts, not unrelated system scripts
	reply, err := c.execute(ctx, PathSystemScript+"/print", "?owner="+ScriptOwnerSales)
	if err != nil {
		return nil, WrapError("get all sales", err)
	}

	sales := make([]Sale, 0)

	for _, re := range reply.Re {
		scriptName := re.Map["name"]

		// Extra safety: ensure the script name matches the sales format
		if !strings.Contains(scriptName, internal.SaleSeparator) {
			continue
		}

		saleData, err := internal.ParseSaleScriptName(scriptName)
		if err != nil {
			continue
		}

		sale := Sale{
			Date:     saleData.Date,
			Time:     saleData.Time,
			Username: saleData.Username,
			Price:    saleData.Price,
			Address:  saleData.Address,
			Mac:      saleData.Mac,
			Validity: saleData.Validity,
			ScriptID: re.Map[".id"],
		}

		// Apply date filter
		if filter != nil && filter.StartDate != "" && filter.EndDate != "" {
			saleDate, _ := time.Parse("Jan/02/2006", sale.Date)
			startDate, _ := time.Parse("Jan/02/2006", filter.StartDate)
			endDate, _ := time.Parse("Jan/02/2006", filter.EndDate)

			if saleDate.Before(startDate) || saleDate.After(endDate) {
				continue
			}
		}

		// Apply prefix filter
		if filter != nil && filter.Prefix != "" {
			if !strings.HasPrefix(sale.Username, filter.Prefix) {
				continue
			}
		}

		sales = append(sales, sale)
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		sales = applyPaginationSales(sales, filter.Offset, filter.Limit)
	}

	return sales, nil
}

// GetSalesByDateRange retrieves sales within date range
func (c *hotspotClient) GetSalesByDateRange(ctx context.Context, startDate, endDate string) ([]Sale, error) {
	if startDate == "" || endDate == "" {
		return nil, fmt.Errorf("start date and end date are required")
	}
	return c.GetAllSales(ctx, &SaleFilter{
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetSalesByPrefix retrieves sales filtered by username prefix
func (c *hotspotClient) GetSalesByPrefix(ctx context.Context, prefix string) ([]Sale, error) {
	if prefix == "" {
		return nil, fmt.Errorf("prefix is required")
	}
	return c.GetAllSales(ctx, &SaleFilter{Prefix: prefix})
}

// GetTotalRevenue calculates total revenue for a date range
func (c *hotspotClient) GetTotalRevenue(ctx context.Context, startDate, endDate string) (float64, error) {
	sales, err := c.GetSalesByDateRange(ctx, startDate, endDate)
	if err != nil {
		return 0, err
	}

	totalRevenue := 0.0
	for _, sale := range sales {
		totalRevenue += sale.Price
	}

	return totalRevenue, nil
}

// DeleteSale deletes sales record by script ID
func (c *hotspotClient) DeleteSale(ctx context.Context, scriptID string) error {
	if scriptID == "" {
		return NewError("delete sale", fmt.Errorf("script ID is required"))
	}

	_, err := c.execute(ctx, PathSystemScript+"/remove", "=.id="+scriptID)
	if err != nil {
		return WrapError("delete sale", err)
	}

	return nil
}

func applyPaginationSales(sales []Sale, offset, limit int) []Sale {
	start := offset
	if start > len(sales) {
		start = len(sales)
	}
	end := start + limit
	if end > len(sales) {
		end = len(sales)
	}
	return sales[start:end]
}
