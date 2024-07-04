package highlevel

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"disruptive/lib/common"
)

const ()

// Contact contains a single contact.
// CustomField.Value can be an int, string, or slice.
type Contact struct {
	ID           string    `json:"id"`
	LocationID   string    `json:"locationId"`
	LocationName string    `json:"locationName"`
	FirstName    string    `json:"firstName,omitempty"`
	LastName     string    `json:"lastName,omitempty"`
	Name         string    `json:"name,omitempty"`
	Email        string    `json:"email,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Address1     string    `json:"address1,omitempty"`
	City         string    `json:"city,omitempty"`
	Country      string    `json:"country,omitempty"`
	DateAdded    time.Time `json:"dateAdded"`
	State        string    `json:"state,omitempty"`
	PostalCode   string    `json:"postalCode,omitempty"`
	Website      string    `json:"website,omitempty"`
	Timezone     string    `json:"timezone,omitempty"`
	DND          bool      `json:"dnd,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
	Source       string    `json:"source,omitempty"`
	CustomField  []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Value any    `json:"value"`
	} `json:"customField,omitempty"`
}

// Contacts contains contacts and metadata.
// NextPage is empty string when there is no next page, otherwise it's an int.
type Contacts struct {
	Contacts []Contact `json:"contacts"`
	Meta     struct {
		Total        int    `json:"total"`
		NextPageURL  string `json:"nextPageUrl"`
		StartAfterID string `json:"startAfterId"`
		StartAfter   int    `json:"startAfter"`
		CurrentPage  int    `json:"currentPage"`
		NextPage     any    `json:"nextPage"`
		PrevPage     int    `json:"prevPage"`
	} `json:"meta"`
}

// GetContacts returns contacts for a location and optional query.
func GetContacts(ctx context.Context, logCtx *slog.Logger, location, query string) (Contacts, error) {
	logCtx = logCtx.With("fid", "highlevel.GetContacts")

	loc, ok := locationsByID[location]
	if !ok {
		loc, ok = locationsByName[location]
		if !ok {
			return Contacts{}, common.ErrBadRequest{Msg: "invalid location"}
		}
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetAuthToken(loc.APIKey).
		SetQueryParam("query", query).
		SetResult(&Contacts{}).
		Get(contactsEndpoint)

	if err != nil {
		logCtx.Error("highlevel contacts endpoint failed", "error", err)
		return Contacts{}, err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("highlevel contacts endpoint failed", "status", res.Status())
		return Contacts{}, errors.New(res.Status())
	}

	fields, err := getCustomFields(ctx, logCtx, loc.APIKey)
	if err != nil {
		return Contacts{}, err
	}

	contacts := res.Result().(*Contacts)

	for i := 0; i < len(contacts.Contacts); i++ {
		loc := locationsByID[contacts.Contacts[i].LocationID]
		contacts.Contacts[i].LocationName = loc.FriendlyName

		for j := 0; j < len(contacts.Contacts[i].CustomField); j++ {
			f, ok := fields[contacts.Contacts[i].CustomField[j].ID]
			if !ok {
				continue
			}
			contacts.Contacts[i].CustomField[j].Name = f.Name
		}
	}

	return *contacts, nil
}
