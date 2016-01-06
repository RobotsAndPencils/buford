package payload

import "encoding/json"

// Browser for Safari Push Notifications
type Browser struct {
	// Alert dictionary.
	Alert   BrowserAlert
	URLArgs []string
}

// BrowserAlert for Safari Push Notifications
type BrowserAlert struct {
	// Title and Body are required
	Title string `json:"title"`
	Body  string `json:"body"`
	// Action button label (defaults to "Show")
	Action string `json:"action,omitempty"`
}

// MarshalJSON allows you to json.Marshal(browser) directly.
func (b Browser) MarshalJSON() ([]byte, error) {
	aps := map[string]interface{}{"alert": b.Alert, "url-args": b.URLArgs}
	return json.Marshal(map[string]interface{}{"aps": aps})
}
