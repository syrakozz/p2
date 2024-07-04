package erc

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"disruptive/config"
)

type rocketReach struct {
	ctx        context.Context
	logCtx     *log.Entry
	m          map[string]int
	id         string
	inFile     *os.File
	csvInFile  *csv.Writer
	pppFile    *os.File
	csvPppFile *csv.Writer
}

func newRocketReach(ctx context.Context, logCtx *log.Entry, id string) (*rocketReach, error) {
	r := &rocketReach{
		ctx:    ctx,
		logCtx: logCtx,
		id:     id,
	}

	// read ppp headers

	headerFile, err := os.Open(filepath.Join(config.VARS.ERCDataRoot, "ppp", "ppp_00.csv"))
	if err != nil {
		log.WithError(err).Error("unable to read file")
		return nil, err
	}
	defer headerFile.Close()

	reader := csv.NewReader(headerFile)

	header, err := reader.Read()
	if err != nil {
		r.logCtx.WithError(err).Error("unable to read header")
		return nil, err
	}

	r.m = make(map[string]int, len(header))
	for i, h := range header {
		r.m[h] = i
	}

	// create in file

	inFile, err := os.Create(filepath.Join(config.VARS.ERCDataRoot, "process", fmt.Sprintf("%s-rocketreach-in.csv", r.id)))
	if err != nil {
		return nil, err
	}

	csvInFile := csv.NewWriter(inFile)
	if err := csvInFile.Write([]string{
		"ID",
		"Company",
		"Address",
		"City",
		"State",
		"Zip",
		"NAICS",
		"BusinessType",
		"Gender",
		"Veteran",
	}); err != nil {
		return nil, err
	}

	r.inFile = inFile
	r.csvInFile = csvInFile

	// create ppp file

	pppFile, err := os.Create(filepath.Join(config.VARS.ERCDataRoot, "process", fmt.Sprintf("%s-ppp.csv", r.id)))
	if err != nil {
		return nil, err
	}

	csvPppFile := csv.NewWriter(pppFile)
	if err := csvPppFile.Write(header); err != nil {
		return nil, err
	}

	r.pppFile = pppFile
	r.csvPppFile = csvPppFile

	return r, nil
}

func (r *rocketReach) write(num int, row []string) error {
	// write rocketreach infile row

	address := row[r.m["BorrowerAddress"]]
	if address == "N/A" {
		address = ""
	}

	city := row[r.m["BorrowerCity"]]
	if city == "N/A" {
		city = ""
	}

	zip := row[r.m["BorrowerZip"]]
	if strings.ContainsRune(zip, '-') && len(zip) > 5 {
		zip = zip[:5]
	}

	businessType := row[r.m["BusinessType"]]
	switch {
	case businessType == "Limited Liability Company(LLC)" || businessType == "Limited  Liability Company(LLC)":
		businessType = "Limited Liability Company (LLC)"
	case businessType == "Employee Stock Ownership Plan(ESOP)":
		businessType = "Employee Stock Ownership Plan (ESOP)"
	case businessType == "Rollover as Business Start-Ups (ROB":
		businessType = "Rollover as Business Start-Ups (ROB)"
	case strings.HasPrefix(businessType, "501(c)"):
		businessType = "501(c) Non-Profit"
	}

	gender := row[r.m["Gender"]]
	if gender == "Unanswered" {
		gender = ""
	}

	veteran := row[r.m["Veteran"]]
	if veteran == "Unanswered" {
		veteran = ""
	}

	if err := r.csvInFile.Write([]string{
		row[r.m["LoanNumber"]],
		row[r.m["BorrowerName"]],
		address,
		city,
		row[r.m["BorrowerState"]],
		zip,
		row[r.m["NAICSCode"]],
		businessType,
		gender,
		veteran,
	}); err != nil {
		return err
	}

	// write ppp row

	if err := r.csvPppFile.Write(row); err != nil {
		return err
	}

	log.WithFields(log.Fields{"num": num + 1, "LoanNumber": row[r.m["LoanNumber"]], "JobsReported": row[r.m["JobsReported"]]}).Info("RocketReach")

	return nil
}

func (r rocketReach) close() {
	if r.csvInFile != nil {
		r.csvInFile.Flush()
	}

	if r.inFile != nil {
		r.inFile.Close()
	}

	if r.csvPppFile != nil {
		r.csvPppFile.Flush()
	}

	if r.pppFile != nil {
		r.pppFile.Close()
	}
}
