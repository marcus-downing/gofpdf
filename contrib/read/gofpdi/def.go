package gofpdi

import (
// "fmt"
// "math"
// "strconv"
// "github.com/jung-kurt/gofpdf"
)

const (
	// MediaBox is a bounding box that includes bleed area and crop marks
	MediaBox = "MediaBox"

	// CropBox is a bounding bos that denotes the area the page should be cropped for display
	CropBox = "CropBox"

	// BleedBox is a bounding box that includes bleed area
	BleedBox = "BleedBox"

	// TrimBox is a bounding box that dentoes the area the page should be trimmed for printing
	TrimBox = "TrimBox"

	// ArtBox is a bounding bxo that denotes an interesting part of the page
	ArtBox = "ArtBox"

	// DefaultBox is the default bounding box to use
	DefaultBox = CropBox
)
