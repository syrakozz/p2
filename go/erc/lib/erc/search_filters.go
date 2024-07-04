package erc

type filterItem struct {
	Display string `json:"display"`
	Value   string `json:"value"`
}

// SearchFilters contain list data values for the UI to display.
type SearchFilters struct {
	Statuses      []filterItem `json:"statuses"`
	States        []filterItem `json:"states"`
	JobsReported  []filterItem `json:"jobs_reported"`
	Naics         []NaicsItem  `json:"naics"`
	BusinessTypes []filterItem `json:"business_types"`
	Genders       []filterItem `json:"genders"`
	Limits        []filterItem `json:"limits"`
}

// GetSearchFilters returns list data value for the UI to display.
func GetSearchFilters() SearchFilters {
	f := SearchFilters{}

	f.Statuses = []filterItem{
		{Display: "Not Started", Value: "Not Started"},
		{Display: "Edit", Value: "Edit"},
		{Display: "Ready", Value: "Ready"},
		{Display: "Sent", Value: "Sent"},
	}

	f.States = []filterItem{
		{Display: "Alabama", Value: "AL"},
		{Display: "Alaska", Value: "AK"},
		{Display: "Arizona", Value: "AZ"},
		{Display: "Arkansas", Value: "AR"},
		{Display: "California", Value: "CA"},
		{Display: "Colorado", Value: "CO"},
		{Display: "Connecticut", Value: "CT"},
		{Display: "Deleware", Value: "DE"},
		{Display: "District of Columbia", Value: "DC"},
		{Display: "Florida", Value: "FL"},
		{Display: "Georgia", Value: "GA"},
		{Display: "Hawaii", Value: "HI"},
		{Display: "Idaho", Value: "ID"},
		{Display: "Illinois", Value: "IL"},
		{Display: "Indiana", Value: "IN"},
		{Display: "Iowa", Value: "IA"},
		{Display: "Kansas", Value: "KS"},
		{Display: "Kentucky", Value: "KY"},
		{Display: "Louisiana", Value: "LA"},
		{Display: "Maine", Value: "ME"},
		{Display: "Maryland", Value: "MD"},
		{Display: "Massachusetts", Value: "MA"},
		{Display: "Michigan", Value: "MI"},
		{Display: "Minnesota", Value: "MN"},
		{Display: "Mississippi", Value: "MS"},
		{Display: "Missouri", Value: "MO"},
		{Display: "Montana", Value: "MT"},
		{Display: "Nebraska", Value: "NE"},
		{Display: "Nevada", Value: "NV"},
		{Display: "New Hampshire", Value: "NH"},
		{Display: "New Jersey", Value: "NJ"},
		{Display: "New Mexico", Value: "NM"},
		{Display: "New York", Value: "NY"},
		{Display: "North Carolina", Value: "NC"},
		{Display: "North Dakota", Value: "ND"},
		{Display: "Ohio", Value: "OH"},
		{Display: "Oklahoma", Value: "OK"},
		{Display: "Oregon", Value: "OR"},
		{Display: "Pennsylvania", Value: "PA"},
		{Display: "Rhode Island", Value: "RI"},
		{Display: "South Carolina", Value: "SC"},
		{Display: "South Dakota", Value: "SD"},
		{Display: "Tennessee", Value: "TN"},
		{Display: "Texas", Value: "TX"},
		{Display: "Utah", Value: "UT"},
		{Display: "Vermont", Value: "VT"},
		{Display: "Virginia", Value: "VA"},
		{Display: "Washington", Value: "WA"},
		{Display: "West Virginia", Value: "WV"},
		{Display: "Wisconsin", Value: "WI"},
		{Display: "Wyoming", Value: "WY"},
	}

	f.JobsReported = []filterItem{
		{Display: "1 - 10", Value: "10"},
		{Display: "11 - 50", Value: "50"},
		{Display: "51 - 200", Value: "200"},
		{Display: "201 - 500", Value: "500"},
		{Display: "501 - 1000", Value: "1000"},
		{Display: "1001 - 5000", Value: "5000"},
		{Display: "5001 - 10000", Value: "10000"},
		{Display: "10001+", Value: "max"},
	}

	f.Naics = GetNaics(3)

	f.BusinessTypes = []filterItem{
		{Display: "501(c) Non-Profit", Value: "501(c) Non-Profit"},
		{Display: "Cooperation", Value: "Cooperation"},
		{Display: "Corporation", Value: "Corporation"},
		{Display: "Employee Stock Ownership Plan (ESOP)", Value: "Employee Stock Ownership Plan (ESOP)"},
		{Display: "Housing Co-op", Value: "Housing Co-op"},
		{Display: "Independent Contrators", Value: "Independent Contrators"},
		{Display: "Joint Venture", Value: "Joint Venture"},
		{Display: "Limited Liability Company (LLC)", Value: "Limited Liability Company (LLC)"},
		{Display: "Limited Liability Partnership", Value: "Limited Liability Partnership"},
		{Display: "Non-Profit Childcare Center", Value: "Non-Profit Childcare Center"},
		{Display: "Non-Profit Organization", Value: "Non-Profit Organization"},
		{Display: "Partnership", Value: "Partnership"},
		{Display: "Professional Association", Value: "Professional Association"},
		{Display: "Qualified Joint-Venture (spouses)", Value: "Qualified Joint-Venture (spouses)"},
		{Display: "Rollover as Business Start-ups (ROB)", Value: "Rollover as Business Start-ups (ROB)"},
		{Display: "Self-Employed Individuals", Value: "Self-Employed Individuals"},
		{Display: "Single Member LLC", Value: "Single Member LLC"},
		{Display: "Sole Proprietorship", Value: "Sole Proprietorship"},
		{Display: "Subchapter S Corporation", Value: "Subchapter S Corporation"},
		{Display: "Tenant in Common", Value: "Tenant in Common"},
		{Display: "Tribal Concerns", Value: "Tribal Concerns"},
		{Display: "Trust", Value: "Trust"},
	}

	f.Limits = []filterItem{
		{Display: "10", Value: "10"},
		{Display: "50", Value: "50"},
		{Display: "100", Value: "100"},
		{Display: "200", Value: "200"},
		{Display: "500", Value: "500"},
	}

	return f
}
