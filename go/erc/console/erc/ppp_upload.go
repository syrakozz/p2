package erc

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	log "github.com/sirupsen/logrus"

	"disruptive/lib/erc"
)

// Upload the ppp csv file.
func Upload(ctx context.Context, file string) {
	logCtx := log.WithField("fid", "console.erc.Upload")

	logCtx.Infof("Uploading: %s", file)

	f, err := os.Open(file)
	if err != nil {
		log.WithError(err).Error("unable to read file")
		return
	}
	defer f.Close()

	r := csv.NewReader(f)

	// read header

	header, err := r.Read()
	if err != nil {
		logCtx.WithError(err).Error("unable to read line")
		return
	}

	m := make(map[string]int, len(header))
	for i, h := range header {
		m[h] = i
	}

	// read rows

	cleanString := strings.NewReplacer(
		"&amp;amp;", "&",
		"&amp;apos;", "'",
		"&amp;", "&",
		"  ", " ",
	)

	n := 1

	for {
		select {
		case <-ctx.Done():
			log.Warn("canceled")
			return
		default:
		}

		row, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logCtx.WithError(err).Error("unable to read file")
			return
		}

		// clean row values

		for i := 0; i < len(row); i++ {
			s := strings.TrimSpace(row[i])
			s = strings.Join(strings.Fields(s), " ")
			row[i] = cleanString.Replace(s)
		}

		// The following conversions ignore errors.
		// If an error does occur, the default zero value is used.

		dateApproved, _ := time.Parse("01/02/2006", row[m["DateApproved"]])
		loanStatusDate, _ := time.Parse("01/02/2006", row[m["LoanStatusDate"]])
		term, _ := strconv.Atoi(row[m["Term"]])
		sbaGuarantyPercentage, _ := strconv.Atoi(row[m["SBAGuarantyPercentage"]])
		currentApprovalAmount, _ := strconv.ParseFloat(row[m["CurrentApprovalAmount"]], 64)
		undisbursedAmount, _ := strconv.ParseFloat(row[m["UndisbursedAmount"]], 64)
		jobsReported, _ := strconv.Atoi(row[m["JobsReported"]])
		forgivenessAmount, _ := strconv.ParseFloat(row[m["ForgivenessAmount"]], 64)
		forgivenessDate, _ := time.Parse("01/02/2006", row[m["ForgivenessDate"]])

		businessTypeFull := ""
		businessType := row[m["BusinessType"]]
		switch {
		case businessType == "Limited Liability Company(LLC)":
			businessType = "Limited Liability Company (LLC)"
		case businessType == "Employee Stock Ownership Plan(ESOP)":
			businessType = "Employee Stock Ownership Plan (ESOP)"
		case businessType == "Rollover as Business Start-Ups (ROB":
			businessType = "Rollover as Business Start-Ups (ROB)"
		case strings.HasPrefix(businessType, "501(c)"):
			// 501(c) types are simplified for consistent querying
			// Save full version in businessTypeFull
			businessTypeFull = businessType
			businessType = "501(c) Non-Profit"

		}

		ppp := erc.PPPData{
			LoanNumber:                  row[m["LoanNumber"]],
			DateApproved:                civil.DateOf(dateApproved),
			SBAOfficeCode:               row[m["SBAOfficeCode"]],
			ProcessingMethod:            strings.ToUpper(row[m["ProcessingMethod"]]),
			BorrowerName:                strings.ToUpper(row[m["BorrowerName"]]),
			BorrowerAddress:             row[m["BorrowerAddress"]],
			BorrowerCity:                strings.ToUpper(row[m["BorrowerCity"]]),
			BorrowerState:               strings.ToUpper(row[m["BorrowerState"]]),
			BorrowerZip:                 row[m["BorrowerZip"]],
			LoanStatusDate:              civil.DateOf(loanStatusDate),
			LoanStatus:                  row[m["LoanStatus"]],
			Term:                        term,
			SBAGuarantyPercentage:       sbaGuarantyPercentage,
			CurrentApprovalAmount:       int(currentApprovalAmount),
			UndisbursedAmount:           int(undisbursedAmount),
			FranchiseName:               row[m["FranchiseName"]],
			ServicingLenderLocationID:   row[m["ServicingLenderLocationID"]],
			ServicingLenderName:         strings.ToUpper(row[m["ServicingLenderName"]]),
			ServicingLenderAddress:      row[m["ServicingLenderAddress"]],
			ServicingLenderCity:         strings.ToUpper(row[m["ServicingLenderCity"]]),
			ServicingLenderState:        strings.ToUpper(row[m["ServicingLenderState"]]),
			ServicingLenderZip:          row[m["ServicingLenderZip"]],
			RuralUrbanIndicator:         strings.ToUpper(row[m["RuralUrbanIndicator"]]),
			HubzoneIndicator:            row[m["HubzoneIndicator"]] == "Y",
			LMIIndicator:                row[m["LMIIndicator"]] == "Y",
			BusinessAgeDescription:      row[m["BusinessAgeDescription"]],
			ProjectCity:                 strings.ToUpper(row[m["ProjectCity"]]),
			ProjectCounty:               strings.ToUpper(row[m["ProjectCounty"]]),
			ProjectState:                strings.ToUpper(row[m["ProjectState"]]),
			ProjectZip:                  row[m["ProjectZip"]],
			CD:                          strings.ToUpper(row[m["CD"]]),
			JobsReported:                jobsReported,
			NAICSCode:                   row[m["NAICSCode"]],
			NAICSName:                   erc.Naics[row[m["NAICSCode"]]],
			Race:                        row[m["Race"]],
			Ethnicity:                   row[m["Ethnicity"]],
			BusinessType:                businessType,
			BusinessTypeFull:            businessTypeFull,
			OriginatingLenderLocationID: row[m["OriginatingLenderLocationID"]],
			OriginatingLenderName:       row[m["OriginatingLender"]],
			OriginatingLenderCity:       strings.ToUpper(row[m["OriginatingLenderCity"]]),
			OriginatingLenderState:      strings.ToUpper(row[m["OriginatingLenderState"]]),
			FemaleOwned:                 row[m["Gender"]] == "Female Owned",
			Veteran:                     row[m["Veteran"]] == "Veteran",
			NonProfit:                   row[m["NonProfit"]] == "Y",
			ForgivenessAmount:           int(forgivenessAmount),
			ForgivenessDate:             civil.DateOf(forgivenessDate),
		}

		if err := erc.UploadPPP(ctx, logCtx, ppp); err != nil {
			log.WithError(err).Error("unable to upload ppp data")
		}

		log.WithFields(log.Fields{"LoanNumber": ppp.LoanNumber, "BorrowerName": ppp.BorrowerName}).Infof("Processing %d", n)
		n++
	}

}
