package gofpdi

/*
 * Copyright (c) 2015 Kurt Jung (Gmail: kurt.w.jung),
 *   Marcus Downing, Jan Slabon (Setasign)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

import (
	"fmt"
	// "os"
	// "strings"
	"regexp"
	// "strconv"
	// "bufio"
	"github.com/jung-kurt/gofpdf"
)

const (
	defaultPdfVersion = "1.3"
)

// OpenPDFParser opens an existing PDF file and readies it
func OpenPDFParser(filename string) (*PDFParser, error) {
	reader, err := NewTokenReader(filename)
	if err != nil {
		return nil, err
	}

	parser := new(PDFParser)
	parser.reader = reader
	parser.pageNumber = 0
	parser.lastUsedPageBox = DefaultBox
	// parser.pages = make([]PDFPage, 0)

	parser.pdfVersion = parser.getPdfVersion()
	fmt.Println("Read PDF version:", parser.pdfVersion)

	return parser, nil
}

// PDFParser is a high-level parser for PDF elements
// See fpdf_pdf_parser.php
type PDFParser struct {
	reader          *PDFTokenReader // the underlying token reader
	pageNumber      int             // the current page number
	lastUsedPageBox string          // the most recently used page box
	pages           []PDFPage       // already loaded pages
	pdfVersion      string          // the document's PDF version
}

// load the initial PDF version string, and offset
func (parser *PDFParser) getPdfVersion() string {
	b, err := parser.reader.Peek(16)
	if err != nil {
		return defaultPdfVersion
	}

	re := regexp.MustCompile("\\d\\.\\d")
	match := re.Find(b)
	if match == nil {
		return defaultPdfVersion
	}
	return string(match[:])
}

func (parser *PDFParser) setPageNumber(pageNumber int) {
	parser.pageNumber = pageNumber
}

// PDFPage is a page extracted from an existing PDF document
type PDFPage struct {
	Dictionary
	Number int
}

type Dictionary map[string]Value

type Value interface{}

// PageBoxes is a transient collection of the page boxes used in a document
// The keys are the constants: MediaBox, CrobBox, BleedBox, TrimBox, ArtBox
type PageBoxes struct {
	pageBoxes       map[string]*PageBox
	lastUsedPageBox string
}

// select a
func (boxes PageBoxes) get(boxName string) *PageBox {
	/**
	 * MediaBox
	 * CropBox: Default -> MediaBox
	 * BleedBox: Default -> CropBox
	 * TrimBox: Default -> CropBox
	 * ArtBox: Default -> CropBox
	 */
	if pageBox, ok := boxes.pageBoxes[boxName]; ok {
		boxes.lastUsedPageBox = boxName
		return pageBox
	}
	switch boxName {
	case BleedBox:
	case TrimBox:
	case ArtBox:
		return boxes.get(CropBox)
	case CropBox:
		return boxes.get(MediaBox)
	}
	return nil
}

// PageBox is the bounding box for a page
type PageBox struct {
	gofpdf.PointType
	gofpdf.SizeType
	Lower gofpdf.PointType // llx, lly
	Upper gofpdf.PointType // urx, ury
}

// getPageBoxes gets the all the bounding boxes for a given page
//
// pageNumber is 1-indexed
// k is a scaling factor from user space units to points
func (parser *PDFParser) getPageBoxes(pageNumber int, k float64) PageBoxes {
	if pageNumber >= num
	page := parser.pages[pageNumber]
	boxes := make(map[string]*PageBox, 5)
	if box := parser.getPageBox(page, MediaBox, k); box != nil {
		boxes[MediaBox] = box
	}
	if box := parser.getPageBox(page, CropBox, k); box != nil {
		boxes[CropBox] = box
	}
	if box := parser.getPageBox(page, BleedBox, k); box != nil {
		boxes[BleedBox] = box
	}
	if box := parser.getPageBox(page, TrimBox, k); box != nil {
		boxes[TrimBox] = box
	}
	if box := parser.getPageBox(page, ArtBox, k); box != nil {
		boxes[ArtBox] = box
	}
	return PageBoxes{boxes, ""}
}

// getPageBox reads a bounding box from a page.
//
// page is a /Page dictionary.
//
// k is a scaling factor from user space units to points.
func (parser *PDFParser) getPageBox(page PDFPage, boxIndex string, k float64) *PageBox {
	// page = parser.resolveObject(page);

	/*
	   $page = $this->resolveObject($page);
	   $box = null;
	   if (isset($page[1][1][$boxIndex])) {
	       $box = $page[1][1][$boxIndex];
	   }

	   if (!is_null($box) && $box[0] == pdf_parser::TYPE_OBJREF) {
	       $tmp_box = $this->resolveObject($box);
	       $box = $tmp_box[1];
	   }

	   if (!is_null($box) && $box[0] == pdf_parser::TYPE_ARRAY) {
	       $b = $box[1];
	       return array(
	           'x' => $b[0][1] / $k,
	           'y' => $b[1][1] / $k,
	           'w' => abs($b[0][1] - $b[2][1]) / $k,
	           'h' => abs($b[1][1] - $b[3][1]) / $k,
	           'llx' => min($b[0][1], $b[2][1]) / $k,
	           'lly' => min($b[1][1], $b[3][1]) / $k,
	           'urx' => max($b[0][1], $b[2][1]) / $k,
	           'ury' => max($b[1][1], $b[3][1]) / $k,
	       );
	   } else if (!isset($page[1][1]['/Parent'])) {
	       return false;
	   } else {
	       return $this->_getPageBox($this->resolveObject($page[1][1]['/Parent']), $boxIndex, $k);
	   }
	*/

	// page = resolveObject(page)

	return nil
}
