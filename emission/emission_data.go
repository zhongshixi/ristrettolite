package emission

type AdSelectionBreakdown struct {
	AdPlatform   EmissionDetails `json:"adPlatform"`
	DataTransfer EmissionDetails `json:"dataTransfer"`
}

type CompensatedBreakdown struct {
	Compensation CompensationDetails `json:"compensation"`
}

type CreativeDeliveryBreakdown struct {
	AdPlatform   EmissionDetails `json:"adPlatform"`
	DataTransfer EmissionDetails `json:"dataTransfer"`
}

type MediaDistributionBreakdown struct {
	Corporate    EmissionDetails `json:"corporate"`
	DataTransfer EmissionDetails `json:"dataTransfer"`
}

type EmissionsBreakdown struct {
	AdSelection       AdSelectionDetails       `json:"adSelection"`
	Compensated       CompensatedDetails       `json:"compensated"`
	CreativeDelivery  CreativeDeliveryDetails  `json:"creativeDelivery"`
	MediaDistribution MediaDistributionDetails `json:"mediaDistribution"`
}

type EmissionsBreakdownWrapper struct {
	Breakdown EmissionsBreakdown `json:"breakdown"`
	Framework string             `json:"framework"`
}

type EmissionDetails struct {
	Emissions float64 `json:"emissions"`
}

type CompensationDetails struct {
	Emissions float64 `json:"emissions"`
	Provider  string  `json:"provider"`
}

type AdSelectionDetails struct {
	Breakdown AdSelectionBreakdown `json:"breakdown"`
	Total     float64              `json:"total"`
}

type CompensatedDetails struct {
	Breakdown CompensatedBreakdown `json:"breakdown"`
	Total     float64              `json:"total"`
}

type CreativeDeliveryDetails struct {
	Breakdown CreativeDeliveryBreakdown `json:"breakdown"`
	Total     float64                   `json:"total"`
}

type MediaDistributionDetails struct {
	Breakdown MediaDistributionBreakdown `json:"breakdown"`
	Total     float64                    `json:"total"`
}

type InternalData struct {
	CountryRegionGCO2PerKwh float64 `json:"countryRegionGCO2PerKwh"`
	CountryRegionCountry    string  `json:"countryRegionCountry"`
	Channel                 string  `json:"channel"`
	DeviceType              string  `json:"deviceType"`
	PropertyId              int     `json:"propertyId"`
	PropertyInventoryType   string  `json:"propertyInventoryType"`
	PropertyName            string  `json:"propertyName"`
	BenchmarkPercentile     int     `json:"benchmarkPercentile"`
	IsMFA                   bool    `json:"isMFA"`
}

type EmissionDataRow struct {
	EmissionsBreakdown EmissionsBreakdownWrapper `json:"emissionsBreakdown"`
	InventoryCoverage  string                    `json:"inventoryCoverage"`
	RowIdentifier      string                    `json:"rowIdentifier,omitempty"`
	TotalEmissions     float64                   `json:"totalEmissions"`
	Internal           InternalData              `json:"internal"`
}

type EmissionResponsePayload struct {
	Rows []EmissionDataRow `json:"rows"`
}
