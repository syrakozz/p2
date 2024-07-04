// Package pinecone integration with pinecone.io
package pinecone

import (
	"disruptive/config"
	"disruptive/lib/common"
	"errors"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL = "https://starter-c85f00b.svc.northamerica-northeast1-gcp.pinecone.io"

	statsStarterEndpoint  = "/describe_index_stats"
	deleteStarterEndpoint = "/vectors/delete"
	fetchStarterEndpoint  = "/vectors/fetch"
	upsertStarterEndpoint = "/vectors/upsert"
	updateStarterEndpoint = "/vectors/update"
	queryStarterEndpoint  = "/query"
)

var (
	// Resty is the shared Resty client for the piencone package.
	Resty *resty.Client
)

// Metadata contains metadata key/value pairs.
// The value is of type any because it can include comparison operations.
type Metadata map[string]string

// Vector contains fields for a single vector.
// SetMetadata is used in update
type Vector struct {
	ID       string    `json:"id"`
	Metadata Metadata  `json:"metadata,omitempty"`
	Values   []float64 `json:"values"`
}

// ErrPinecone contains the error from the pinecone API.
type ErrPinecone struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Details []struct {
		TypeURL string `json:"typeURL"`
		Value   string `json:"value"`
	} `json:"details"`
}

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("Api-Key", config.VARS.PineconeKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetError(&ErrPinecone{}).
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}

func (e ErrPinecone) Error() string {
	return e.Message
}

// Is interface method
func (e ErrPinecone) Is(err error) bool {
	as := ErrPinecone{}
	return errors.As(err, &as)
}
