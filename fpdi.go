package gofpdf

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
	// "fmt"
	// "math"
	// "strconv"
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

// OpenFile opens an existing PDF file for reading
func OpenFile(filename string) (*TemplateDocument, error) {
	parser, err := openParser(filename)
	if err != nil {
		return nil, err
	}

	td := new(TemplateDocument)
	td.parser = parser
	td.pdfVersion = td.parser.getPdfVersion()

	// ???

	return td, nil
}

// TemplateDocument represents a PDF file parser that can be used to load templates to use in other documents
type TemplateDocument struct {
	numPages        int     // the number of pages in the PDF cocument
	lastUsedPageBox string  // the most recently used value of boxName
	parser          *PDFParser // the actual document reader
	pdfVersion      string
}

// CountPages returns the number of pages in this source document
func (td *TemplateDocument) CountPages() int {
	return td.numPages
}

// ImportPage imports a single page of the source document to use as a template in another document
func (td *TemplateDocument) ImportPage(pageNumber int, boxName string, groupXObject bool) Template {
	if boxName == "" {
		boxName = DefaultBox
	}
	td.parser.setPageNumber(pageNumber)

	t := new(TemplateDocumentTemplate)
	t.id = GenerateTemplateID()

	pageBoxes := td.parser.getPageBoxes(pageNumber, td.scaleFactor)
	pageBox := pageBoxes.get(boxName)
	td.lastUsedPageBox = pageBoxes.lastUsedPageBox

	return t
}

// GetLastUsedPageBox returns the last used page boundary box.
func (td *TemplateDocument) GetLastUsedPageBox() string {
	return td.lastUsedPageBox
}

// Close releases references and closes the file handler of the parser
func (td *TemplateDocument) Close() {

}
