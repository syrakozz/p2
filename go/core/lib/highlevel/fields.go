package highlevel

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type field struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	FieldKey        string   `json:"fieldKey"`
	PlaceHolder     string   `json:"placeholder"`
	Position        int      `json:"position"`
	DateType        string   `json:"dateType"`
	PickListOptions []string `json:"picklistOptions"`
}

type fields struct {
	CustomFields []field `json:"customFields"`
}

type customFields map[string]field

func getCustomFields(ctx context.Context, logCtx *slog.Logger, apiKey string) (customFields, error) {
	logCtx = logCtx.With("fid", "highlevel.getCustomFields")

	res, err := Resty.R().
		SetContext(ctx).
		SetAuthToken(apiKey).
		SetResult(&fields{}).
		Get(customFieldsEndpoint)

	if err != nil {
		logCtx.Error("highlevel custom-fields endpoint failed", "error", err)
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("highlevel custom-fields endpoint failed", "status", res.StatusCode())
		return nil, errors.New(res.Status())
	}

	cf := customFields{}
	for _, f := range *res.Result().(*customFields) {
		cf[f.ID] = f
	}

	return cf, nil
}
