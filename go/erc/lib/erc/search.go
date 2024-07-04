package erc

import (
	"context"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SearchCriteria contains search filters for the SQL query
type SearchCriteria struct {
	Status                   string `json:"status"`
	LoanNumber               string `json:"loan_number"`
	BorrowerName             string `json:"borrower_name"`
	CurrentApprovalAmountMin int    `json:"current_approval_amount_min"`
	CurrentApprovalAmountMax int    `json:"current_approval_amount_max"`
	ServicingLenderName      string `json:"servicing_lender_name"`
	ProjectCity              string `json:"project_city"`
	ProjectState             string `json:"project_state"`
	JobsReported             int    `json:"jobs_reported"`
	NaicsCode                string `json:"naics_code"`
	BusinessType             string `json:"business_type"`
	FemaleOwned              bool   `json:"female_owned"`
	Veteran                  bool   `json:"veteran"`
	NonProfit                bool   `json:"non_profit"`
	Limit                    int    `json:"limit"`
}

// GetSearch retrieves PPP data from Spanner
func GetSearch(ctx context.Context, logCtx *log.Entry, criteria SearchCriteria) ([]PPPData, error) {
	_ = logCtx.WithField("fid", "erc.GetSearch")

	sql := strings.Builder{}
	sql.WriteString("SELECT * FROM PPP")

	w := make([]string, 0, 10)

	if criteria.Status != "" {
		w = append(w, "status = @status")
	}

	if criteria.LoanNumber != "" {
		w = append(w, "loan_number = @loan_number")
	}

	if criteria.BorrowerName != "" {
		w = append(w, "borrower_name = @borrower_name")
	}

	if criteria.CurrentApprovalAmountMin > 0 {
		w = append(w, "current_approval_amount >= @current_approval_amount_min")
	}

	if criteria.CurrentApprovalAmountMax > 0 {
		w = append(w, "current_approval_amount <= @current_approval_amount_max")
	}

	if criteria.ServicingLenderName != "" {
		w = append(w, "servicing_lender_name = @servicing_lender_name")
	}

	if criteria.ProjectCity != "" {
		w = append(w, "project_city = @project_city")
	}

	if criteria.ProjectState != "" {
		w = append(w, "project_state = @project_state")
	}

	if criteria.JobsReported > 0 {
		w = append(w, "jobs_reported = @jobs_reported")
	}

	if criteria.NaicsCode != "" {
		w = append(w, "naics_code LIKE @naics_code")
	}

	if criteria.BusinessType != "" {
		w = append(w, "business_type = @busisenss_type")
	}

	if criteria.FemaleOwned {
		w = append(w, "female_owned = TRUE")
	}

	if criteria.Veteran {
		w = append(w, "veteran = TRUE")
	}

	if criteria.NonProfit {
		w = append(w, "non_profit = TRUE")
	}

	if where := strings.Join(w, " AND "); where != "" {
		sql.WriteString(" WHERE ")
		sql.WriteString(where)
	}

	if criteria.Limit > 0 {
		sql.WriteString(" LIMIT ")
		sql.WriteString(strconv.Itoa(criteria.Limit))
	}

	sql.WriteRune(';')

	results := []PPPData{}
	results = append(results, PPPData{LoanNumber: "123456"})

	return results, nil
}
