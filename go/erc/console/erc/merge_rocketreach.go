package erc

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"disruptive/config"
)

type (
	rrLoansProcessed map[string]struct{}
)

func mergeRocketReach(ctx context.Context, basename string, pppHeaders pppHeaderIndexes, pppRows pppRowsByLoanNumber) (rrLoansProcessed, error) {
	logCtx := log.WithField("basename", basename)

	path := filepath.Join(config.VARS.ERCDataRoot, "process", basename+"-rocketreach-out.csv")

	f, err := os.Open(path)
	if err != nil {
		logCtx.WithError(err).Error("unable to open file")
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// read header

	row, err := r.Read()
	if err != nil {
		logCtx.WithError(err).Error("unable to read header")
		return nil, err
	}

	rrHeaders := make(map[string]int, len(row))
	for i, h := range row {
		rrHeaders[h] = i
	}

	// create highlevel file

	highlevelHeaders := []string{
		"First Name",
		"Last Name",
		"Job Title",
		"Phone",
		"Email",
		"Source",
		"Type",
		"Business Name",
		"City",
		"State",
		"Country",
		"Postal Code",
		"Website",
		"Business Phone",
		"LinkedIn",
		"Company Revenue",
		"Company Size",
		"Attributes",
		"Business Type",
		"Industry",
		"PPP Loan Number",
		"PPP Current Approval Amount",
		"PPP Servicing Lender Name",
		"PPP Jobs Reported",
		"ERC Estimate Low",
		"ERC Estimate High",
	}

	highlevelFile, err := os.Create(filepath.Join(config.VARS.ERCDataRoot, "process", fmt.Sprintf("%s-highlevel.csv", basename)))
	if err != nil {
		return nil, err
	}
	defer highlevelFile.Close()

	w := csv.NewWriter(highlevelFile)
	if err := w.Write(highlevelHeaders); err != nil {
		logCtx.WithError(err).Error("unable to write highlevel headers")
		return nil, err
	}
	defer w.Flush()

	// read rows

	rrProcessed := rrLoansProcessed{}

	for {
		select {
		case <-ctx.Done():
			log.Warn("canceled")
			return nil, context.Canceled
		default:
		}

		rrRow, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logCtx.WithError(err).Error("unable to read file")
			return nil, err
		}

		phone := rrRow[rrHeaders["Phone"]]
		if len(phone) > 0 && phone[0] == '#' {
			phone = ""
		}

		i, ok := rrHeaders["disruptive_id"]
		if !ok {
			logCtx.Error("D1srupt1ve ID column in RocketReach file missing")
			return nil, errors.New("D1srupt1ve ID column in RocketReach file missing")
		}
		loanNumber := rrRow[i]

		pppRow := pppRows[loanNumber]
		if len(pppRow) == 0 {
			continue
		}

		country := ""
		if rrRow[rrHeaders["Country"]] == "US" {
			country = "United States"
		}

		revenue := rrRow[rrHeaders["Employer Revenue"]]

		switch {
		case strings.HasSuffix(revenue, " Million"):
			f, err := strconv.ParseFloat(revenue[1:len(revenue)-8], 64)
			if err != nil {
				revenue = ""
			}
			revenue = strconv.Itoa(int(f * 1_000_000))
		case strings.HasSuffix(revenue, " Billion"):
			f, err := strconv.ParseFloat(revenue[1:len(revenue)-8], 64)
			if err != nil {
				revenue = ""
			}
			revenue = strconv.Itoa(int(f * 1_000_000_000))
		}

		attributes := []string{}
		if pppRow[pppHeaders["Gender"]] == "Female Owned" {
			attributes = append(attributes, "Female Owned")
		}
		if pppRow[pppHeaders["NonProfit"]] == "Y" {
			attributes = append(attributes, "Non-Profit")
		}
		if pppRow[pppHeaders["Veteran"]] == "Veteran" {
			attributes = append(attributes, "Veteran")
		}

		businessType := pppRow[pppHeaders["BusinessType"]]
		switch {
		case businessType == "Limited Liability Company(LLC)" || businessType == "Limited  Liability Company(LLC)":
			businessType = "Limited Liability Company (LLC)"
		case businessType == "Employee Stock Ownership Plan(ESOP)":
			businessType = "Employee Stock Ownership Plan (ESOP)"
		case businessType == "Rollover as Business Start-Ups (ROB":
			businessType = "Rollover as Business Start-Ups (ROB)"
		case strings.HasPrefix(businessType, "501(c)"):
			// 501(c) types are simplified for consistent querying
			businessType = "501(c) Non-Profit"
		}

		jobsReported := pppRow[pppHeaders["JobsReported"]]
		jobsReportedNum, err := strconv.Atoi(jobsReported)
		if err != nil {
			return nil, errors.New("invalid Jobs Reproted column in PPP file")
		}

		ercEstimateLow := strconv.Itoa(jobsReportedNum * 10000)
		ercEstimateHigh := strconv.Itoa(jobsReportedNum * 15000)

		w.Write([]string{
			rrRow[rrHeaders["First Name"]],
			rrRow[rrHeaders["Last Name"]],
			rrRow[rrHeaders["Title"]],
			phone,
			rrRow[rrHeaders["Recommended Email"]],
			"PPP",
			"lead",
			pppRow[pppHeaders["BorrowerName"]],
			rrRow[rrHeaders["City"]],
			rrRow[rrHeaders["Region"]],
			country,
			rrRow[rrHeaders["Postal Code"]],
			rrRow[rrHeaders["Employer Website"]],
			rrRow[rrHeaders["Office Phone"]],
			rrRow[rrHeaders["LinkedIn"]],
			revenue,
			rrRow[rrHeaders["Employer Size"]],
			strings.Join(attributes, ","),
			businessType,
			rrRow[rrHeaders["Employer Industry 1"]],
			loanNumber,
			pppRow[pppHeaders["CurrentApprovalAmount"]],
			pppRow[pppHeaders["ServicingLenderName"]],
			jobsReported,
			ercEstimateLow,
			ercEstimateHigh,
		})

		rrProcessed[loanNumber] = struct{}{}

	}

	return rrProcessed, nil
}

func mergeRocketReachNP(ctx context.Context, basename string, header pppHeader, headersIndexes pppHeaderIndexes, rows pppRowsByLoanNumber, loansProcessed rrLoansProcessed) error {
	logCtx := log.WithField("basename", basename)

	np, err := os.Create(filepath.Join(config.VARS.ERCDataRoot, "process", fmt.Sprintf("%s-ppp-np.csv", basename)))
	if err != nil {
		logCtx.WithError(err).Error("unable to create ppp-np file")
		return err
	}
	defer np.Close()

	// write header
	w := csv.NewWriter(np)
	if err := w.Write(header); err != nil {
		logCtx.WithError(err).Error("unable to write ppp-np header row")
		return err
	}
	defer w.Flush()

	// delete rows already processed
	for loan := range loansProcessed {
		delete(rows, loan)
	}

	// write leftover rows
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			logCtx.WithError(err).Error("unable to write ppp-np row")
			return err
		}
	}

	return nil
}
