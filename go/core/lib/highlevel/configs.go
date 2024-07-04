package highlevel

import (
	"context"
	"errors"
	"log/slog"

	"disruptive/lib/configs"
)

type location struct {
	APIKey       string
	LocationID   string
	FriendlyName string
	Name         string
}

var (
	locationsByName = map[string]location{}
	locationsByID   = map[string]location{}
)

// LoadConfigs loads highlevel config files from firestore.
func LoadConfigs(ctx context.Context, logCtx *slog.Logger) error {
	logCtx = logCtx.With("fid", "highlevel.GetConfigs")
	config, err := configs.Get(ctx, logCtx, "highlevel_locations")
	if err != nil {
		return err
	}

	for k, v := range config {
		entry, ok := v.(map[string]any)
		if !ok {
			return errors.New("unable to parse highlevel_locations")
		}

		apiKey, ok := entry["api_key"].(string)
		if !ok {
			return errors.New("unable to parse highlevel_locations")
		}

		locationID, ok := entry["location_id"].(string)
		if !ok {
			return errors.New("unable to parse highlevel_locations")
		}

		friendlyName, ok := entry["friendly_name"].(string)
		if !ok {
			return errors.New("unable to parse highlevel_locations")
		}

		loc := location{
			Name:         k,
			FriendlyName: friendlyName,
			LocationID:   locationID,
			APIKey:       apiKey,
		}

		locationsByName[loc.Name] = loc
		locationsByID[loc.LocationID] = loc
	}

	return nil
}
