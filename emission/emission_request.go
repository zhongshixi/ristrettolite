package emission

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

type EmissionRequestRow struct {
	RowIdentifier string `json:"rowIdentifier,omitempty"`
	Impressions   int    `json:"impressions"`
	UtcDatetime   string `json:"utcDatetime"`
	InventoryID   string `json:"inventoryId"`
	Country       string `json:"country"`
	Channel       string `json:"channel"`

	Cost int `json:"cost,omitempty"`
}

func (e *EmissionRequestRow) String() string {
	return fmt.Sprintf("Impressions: %d, UtcDatetime: %s, InventoryID: %s, Country: %s, Channel: %s",
		e.Impressions, e.UtcDatetime, e.InventoryID, e.Country, e.Channel)
}

type EmissionRequestPayload struct {
	Rows []EmissionRequestRow `json:"rows"`
}

// GenerateEmissionRequestPayload generates a payload with N rows.
func GenerateEmissionRequestPayload(N int) EmissionRequestPayload {
	// Valid inventory IDs and country codes
	inventoryIDs := []string{"nytimes.com", "fandom.com"}
	countryCodes := []string{"US", "CA", "UK", "DE", "FR", "JP", "AU", "IN"}

	rows := make([]EmissionRequestRow, N)

	// Generate rows
	for i := 0; i < N; i++ {
		// Generate a unique row identifier
		rowID := fmt.Sprintf("row-%d", i+1)

		// Generate a unique UTC datetime (incrementing day for each row)
		date := time.Now().AddDate(0, 0, i).Format("2006-01-02")

		// Randomly select an inventory ID and country
		inventoryID := inventoryIDs[rand.Intn(len(inventoryIDs))]
		country := countryCodes[rand.Intn(len(countryCodes))]

		// Populate the row
		rows[i] = EmissionRequestRow{
			RowIdentifier: rowID,
			Impressions:   1000,
			UtcDatetime:   date,
			InventoryID:   inventoryID,
			Country:       country,
			Channel:       "web",
			// Randomly assign a cost from
			Cost: rand.Intn(10) + 1,
		}
	}

	return EmissionRequestPayload{Rows: rows}
}
