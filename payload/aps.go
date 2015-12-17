// Package payload
package payload

import (
	"encoding/json"

	"github.com/RobotsAndPencils/buford/payload/badge"
)

// APS is Apple's reserved namespace.
type APS struct {
	// Alert dictionary.
	Alert Alert

	// Badge to display on the app icon.
	// Set to badge.Preserve (default), badge.Clear
	// or a specific value with badge.New(n).
	Badge badge.Badge

	// The name of a sound file to play as an alert.
	Sound string

	// Content available apps launched in the background or resumed...
	ContentAvailable bool

	// Category identifier for custom actions in iOS 8 or newer.
	Category string
}

// Alert dictionary
type Alert struct {
	// Title is a short string shown briefly on Apple Watch in iOS 8.2 or newer.
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	Action       string   `json:"action,omitempty"`
	LocKey       string   `json:"loc-key,omitempty"`
	LocArgs      []string `json:"loc-args,omitempty"`
	ActionLocKey string   `json:"action-loc-key,omitempty"`
	LaunchImage  string   `json:"launch-image,omitempty"`
}

// isSimple alert with only Body set.
func (a *Alert) isSimple() bool {
	return len(a.Title) == 0 && len(a.Action) == 0 && len(a.LocKey) == 0 && len(a.LocArgs) == 0 && len(a.ActionLocKey) == 0 && len(a.LaunchImage) == 0
}

// isZero if no Alert fields are set.
func (a *Alert) isZero() bool {
	return a.isSimple() && len(a.Body) == 0
}

// Map returns the APS payload as a map that you can customize
// before serializing it to JSON.
func (a *APS) Map() map[string]interface{} {
	aps := make(map[string]interface{}, 4)

	if !a.Alert.isZero() {
		if a.Alert.isSimple() {
			aps["alert"] = a.Alert.Body
		} else {
			aps["alert"] = a.Alert
		}
	}
	if n, ok := a.Badge.Number(); ok {
		aps["badge"] = n
	}
	if a.Sound != "" {
		aps["sound"] = a.Sound
	}
	if a.ContentAvailable {
		aps["content-available"] = 1
	}

	// wrap in "aps" to form final payload
	return map[string]interface{}{"aps": aps}
}

// MarshalJSON allows you to json.Marshal(aps) directly.
func (a APS) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Map())
}
