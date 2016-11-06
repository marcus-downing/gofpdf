package types

import (
	// "fmt"
	// "math"
	// "strconv"
	"github.com/jung-kurt/gofpdf"
)


const (
	// DefaultPdfVersion is the format version to assume files are until told otherwise
	DefaultPdfVersion = "1.3"
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


// PageBox is the bounding box for a page
type PageBox struct {
	gofpdf.PointType
	gofpdf.SizeType
	Lower gofpdf.PointType // llx, lly
	Upper gofpdf.PointType // urx, ury
}

// PageBoxes is a transient collection of the page boxes used in a document
// The keys are the constants: MediaBox, CrobBox, BleedBox, TrimBox, ArtBox
type PageBoxes struct {
	PageBoxes       map[string]*PageBox
	LastUsedPageBox string
}

// select a
func (boxes PageBoxes) Get(boxName string) *PageBox {
	/**
	 * MediaBox
	 * CropBox: Default -> MediaBox
	 * BleedBox: Default -> CropBox
	 * TrimBox: Default -> CropBox
	 * ArtBox: Default -> CropBox
	 */
	if pageBox, ok := boxes.PageBoxes[boxName]; ok {
		boxes.LastUsedPageBox = boxName
		return pageBox
	}
	switch boxName {
	case BleedBox:
	case TrimBox:
	case ArtBox:
		return boxes.Get(CropBox)
	case CropBox:
		return boxes.Get(MediaBox)
	}
	return nil
}
