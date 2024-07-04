package erc

import (
	"context"

	"cloud.google.com/go/civil"
	log "github.com/sirupsen/logrus"
)

// PPPData contains one row of the raw PPP CSV data.
type PPPData struct {
	LoanNumber                  string
	DateApproved                civil.Date
	SBAOfficeCode               string
	ProcessingMethod            string
	BorrowerName                string
	BorrowerAddress             string
	BorrowerCity                string
	BorrowerState               string
	BorrowerZip                 string
	LoanStatusDate              civil.Date
	LoanStatus                  string
	Term                        int
	SBAGuarantyPercentage       int
	CurrentApprovalAmount       int
	UndisbursedAmount           int
	FranchiseName               string
	ServicingLenderLocationID   string
	ServicingLenderName         string
	ServicingLenderAddress      string
	ServicingLenderCity         string
	ServicingLenderState        string
	ServicingLenderZip          string
	RuralUrbanIndicator         string
	HubzoneIndicator            bool
	LMIIndicator                bool
	BusinessAgeDescription      string
	ProjectCity                 string
	ProjectCounty               string
	ProjectState                string
	ProjectZip                  string
	CD                          string
	JobsReported                int
	NAICSCode                   string
	NAICSName                   string
	Race                        string
	Ethnicity                   string
	BusinessType                string
	BusinessTypeFull            string
	OriginatingLenderLocationID string
	OriginatingLenderName       string
	OriginatingLenderCity       string
	OriginatingLenderState      string
	FemaleOwned                 bool
	Veteran                     bool
	NonProfit                   bool
	ForgivenessAmount           int
	ForgivenessDate             civil.Date
}

// UploadPPP sends PPP data to the database.
func UploadPPP(ctx context.Context, logCtx *log.Entry, ppp PPPData) error {
	_ = logCtx.WithField("fid", "erc.UploadPPP")

	return nil
}
