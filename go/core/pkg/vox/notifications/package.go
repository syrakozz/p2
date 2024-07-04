// Package notifications ...
package notifications

import "time"

// Document contains the notification document.
type Document struct {
	ID              string           `firestore:"id" json:"id"`
	Inactive        bool             `firestore:"inactive" json:"inactive"`
	ModerationValue *ModerationValue `firestore:"moderation_value,omitempty" json:"moderation_value,omitempty"`
	Read            bool             `firestore:"read" json:"read"`
	Timestamp       time.Time        `firestore:"timestamp" json:"timestamp"`
	Type            string           `firestore:"type" json:"type"`
}
