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
	// "math"
	"errors"
	// "strconv"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi/readpdf"
	. "github.com/jung-kurt/gofpdf/contrib/gofpdi/types"
)

// Open makes an existing PDF file usable for templates
func Open(filename string) (*Fpdi, error) {
	parser, err := readpdf.OpenPDFParser(filename)
	if err != nil {
		return nil, err
	}

	fpdi := new(Fpdi)
	fpdi.parser = parser
	fpdi.pdfVersion = fpdi.parser.GetPDFVersion()
	fpdi.k = 1.0
	fpdi.numPages = parser.CountPages()
	fmt.Println("Num pages:", fpdi.numPages)
	// td.k = ???
	// ???

	return fpdi, nil
}

// Fpdi represents a PDF file parser which can load templates to use in other documents
type Fpdi struct {
	numPages        int                // the number of pages in the PDF cocument
	lastUsedPageBox string             // the most recently used value of boxName
	parser          *readpdf.PDFParser // the actual document reader
	pdfVersion      string             // the PDF version
	k               float64            // default scale factor (number of points in user unit)
}

// Error returns the internal parser error; this will be nil if no error has occurred.
func (fpdi *Fpdi) Error() error {
	return fpdi.parser.Error()
}

// CountPages returns the number of pages in this source document
func (fpdi *Fpdi) CountPages() int {
	return fpdi.numPages
}

// Page imports a single page of the source document using default settings
func (fpdi *Fpdi) Page(pageNumber int) gofpdf.Template {
	return fpdi.ImportPage(pageNumber, DefaultBox, false)
}

// ImportPage imports a single page of the source document to use as a template in another document
func (fpdi *Fpdi) ImportPage(pageNumber int, boxName string, groupXObject bool) gofpdf.Template {
	if pageNumber > fpdi.numPages {
		fmt.Println("Page", pageNumber, "does not exist")
		fpdi.parser.SetError(errors.New("Page does not exist"))
		return nil
	}
	if boxName == "" {
		boxName = DefaultBox
	}

	page := fpdi.parser.GetPageParser(pageNumber)

	t := new(TemplatePage)
	t.id = gofpdf.GenerateTemplateID()

	pageBoxes := page.GetPageBoxes(fpdi.k)
	// fmt.Println("Page boxes:", pageBoxes)
	pageBox := pageBoxes.Get(boxName)
	t.pageSize = pageBox.SizeType
	// fmt.Println("FPDI: Template size:", t.pageSize)
	fpdi.lastUsedPageBox = pageBoxes.LastUsedPageBox

	// get the actual page bytes verbatim
	t.bytes = page.Bytes()
	// fmt.Println("FPDI: Template bytes:", string(t.bytes))

	// load the fonts used on this page
	// these will be resolved by the template system
	t.fonts = page.Fonts()
	fmt.Println("FPDI: Template fonts:", t.fonts)

	t.images = page.Images()
	t.templates = page.Templates()

	return t
}

// GetLastUsedPageBox returns the last used page boundary box.
func (fpdi *Fpdi) GetLastUsedPageBox() string {
	return fpdi.lastUsedPageBox
}

// Close releases references and closes the file handle of the parser
func (fpdi *Fpdi) Close() {
	fpdi.parser.Close()
}
