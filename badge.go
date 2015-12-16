package buford

// Badge number to display on the App icon
type Badge struct {
	number uint
	isSet  bool
}

// PreserveBadge to not change the current badge (default behavior)
var PreserveBadge = Badge{}

// ClearBadge can be used in a payload to clear the badge.
var ClearBadge = Badge{number: 0, isSet: true}

// NewBadge to set the badge to a specific value
func NewBadge(number uint) Badge {
	return Badge{number: number, isSet: true}
}

// Number to display on the App Icon and if should be changed.
func (b *Badge) Number() (uint, bool) {
	return b.number, b.isSet
}
