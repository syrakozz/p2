package mailfinder

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// Employees contains employee data.
type Employees struct {
	CompanyNameUsed string `json:"company_name_used"`
	Employees       []struct {
		Name    string
		Title   string
		Company string
	}
	Success bool
}

// PostSearchEmployees returns a company's employees.
func PostSearchEmployees(ctx context.Context, logCtx *slog.Logger, domain, company, title, country string) (Employees, error) {
	logCtx = logCtx.With("fid", "mailfinder.PostSearchEmployees")

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(map[string]string{"domain": domain, "company_name": company, "preferred_title": title, "country_code": country}).
		SetResult(&Employees{}).
		Post(anymailfinderSearchEmployeesEndpoint)

	if err != nil {
		logCtx.Error("anymailfinder search employees endpoint failed", "error", err)
		return Employees{}, err
	}

	if res.StatusCode() == http.StatusNotFound {
		logCtx.Error("unable to retrieve search employees result", "status", res.Status())
		return Employees{}, common.ErrNotFound{Msg: "unable to retrieve search employees result"}
	}

	if res.StatusCode() == http.StatusUnavailableForLegalReasons {
		logCtx.Error("blacklisted", "status", res.Status())
		return Employees{}, common.ErrNotFound{Msg: "blacklisted"}
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("anymailfinder endpoint failed", "status", res.Status())
		return Employees{}, errors.New(res.Status())
	}

	return *res.Result().(*Employees), nil
}
