package buford

import "encoding/json"

// Payload to be serialized to JSON.
type Payload struct {
	Alert Alert `json:"alert"`

	// Badge to display on the app icon.
	// See PreserveBadge (default), ClearBadge, NewBadge.
	Badge Badge

	// The name of a sound file to play as an alert.
	Sound string

	// Content available apps launched in the background or resumed...
	ContentAvailable bool

	// Category identifier for custom actions in iOS 8 or newer.
	Category string
}

// MDMPayload for mobile device management.
type MDMPayload struct {
	Token string `json:"mdm"`
}

// Alert dictionary
type Alert struct {
	// Title is a short string shown briefly on Apple Watch in iOS 8.2.
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

// MarshalJSON does marshal
func (p Payload) MarshalJSON() ([]byte, error) {
	aps := make(map[string]interface{}, 4)

	if !p.Alert.isZero() {
		if p.Alert.isSimple() {
			aps["alert"] = p.Alert.Body
		} else {
			aps["alert"] = p.Alert
		}
	}
	if n, ok := p.Badge.Number(); ok {
		aps["badge"] = n
	}
	if p.Sound != "" {
		aps["sound"] = p.Sound
	}
	if p.ContentAvailable {
		aps["content-available"] = 1
	}

	// wrap in "aps" to form final payload
	payload := struct {
		APS map[string]interface{} `json:"aps"`
	}{APS: aps}

	return json.Marshal(payload)
}
