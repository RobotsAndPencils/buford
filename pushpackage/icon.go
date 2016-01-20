package pushpackage

import "io"

// IconSet is a set of icons. Six images are required.
type IconSet []Icon

// Icon to display to the user (PNG format).
type Icon struct {
	// Name of file to send to Apple (eg. icon_16x16.png)
	Name string
	// Reader for the image data
	Reader io.Reader
}

const iconDirectory = "icon.iconset/"
