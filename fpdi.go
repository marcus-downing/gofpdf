package gofpdf

//
//  GoFPDI
//
//    Copyright 2015 Marcus Downing
//
//  FPDI - Version 1.5.2
//
//    Copyright 2004-2014 Setasign - Jan Slabon
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

import (
	"fmt"
	"math"
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
func OpenFile(filename string) (*Fpdi, error) {
	parser, err := openParser(filename)
	if err != nil {
		return nil, err
	}

	fpdi := new(Fpdi)
	fpdi.parser = parser

	version := math.Max(1.3, fpdi.parser.getPdfVersion())
	fpdi.pdfVersion = fmt.Sprintf("%.1f", version)

	return fpdi, nil
}

// Fpdi represents a PDF file parser that can be used to load templates to use in other documents
type Fpdi struct {
	numPages        int    // the number of pages in the PDF cocument
	lastUsedPageBox string // the most recently used value of boxName
	parser          *parser
	pdfVersion      string
}

// CountPages returns the number of pages in this source document
func (i *Fpdi) CountPages() int {
	return i.numPages
}

// ImportPage imports a single page of the source document to use as a template in another document
func (i *Fpdi) ImportPage(pageNumber int, boxName string, groupXObject bool) Template {
	if boxName == "" {
		boxName = DefaultBox
	}
	i.parser.setPageNumber(pageNumber)

	t := new(fpdiTemplate)
	t.id = GenerateTemplateID()
	// t.pageSize =
	return t
}

// GetLastUsedPageBox returns the last used page boundary box.
func (i *Fpdi) GetLastUsedPageBox() string {
	return i.lastUsedPageBox
}

// Close releases references and closes the file handler of the parser
func (i *Fpdi) Close() {

}
