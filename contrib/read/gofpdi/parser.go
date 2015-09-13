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
	"strings"
	// "regexp"
	// "bytes"
	"strconv"
	// "bufio"
	"errors"
	// "github.com/jung-kurt/gofpdf"
)

const (
	defaultPdfVersion = "1.3"
)

// OpenPDFParser opens an existing PDF file and readies it
func OpenPDFParser(filename string) (*PDFParser, error) {
	// fmt.Println("Opening PDF file:", filename)
	reader, err := NewTokenReader(filename)
	if err != nil {
		return nil, err
	}

	parser := new(PDFParser)
	parser.reader = reader
	parser.pageNumber = 0
	parser.lastUsedPageBox = DefaultBox

	// read xref data
	// xrefOffset, err := reader.findXrefTable()
	// if err != nil {
	// 	return nil, err
	// }
	offset, err := parser.reader.findXrefTable()
	if err != nil {
		return nil, err
	}
	err = parser.readXrefTable(offset)
	if err != nil {
		return nil, err
	}

	// check for encryption
	// ...

	// read root
	// pagesDictionary := parser.resolveObject("/Pages")

	return parser, nil
}

// PDFParser is a high-level parser for PDF elements
// See fpdf_pdf_parser.php
type PDFParser struct {
	reader          *PDFTokenReader // the underlying token reader
	pageNumber      int             // the current page number
	lastUsedPageBox string          // the most recently used page box
	pages           []PDFPage       // already loaded pages

	xref struct {
		maxObject    int                 // the highest xref object number
		xrefLocation int64               // the location of the xref table
		xref         map[int]map[int]int // all the xref offsets
	}
}

func (parser *PDFParser) setPageNumber(pageNumber int) {
	parser.pageNumber = pageNumber
}

// Close releases references and closes the file handle of the parser
func (parser *PDFParser) Close() {
	parser.reader.Close()
}

// PDFPage is a page extracted from an existing PDF document
type PDFPage struct {
	Dictionary
	Number int
}

// GetPageBoxes gets the all the bounding boxes for a given page
//
// pageNumber is 1-indexed
// k is a scaling factor from user space units to points
func (parser *PDFParser) GetPageBoxes(pageNumber int, k float64) PageBoxes {
	boxes := make(map[string]*PageBox, 5)
	if pageNumber >= len(parser.pages) {
		return PageBoxes{boxes, DefaultBox}
	}

	page := parser.pages[pageNumber]
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
	return PageBoxes{boxes, DefaultBox}
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

func (parser *PDFParser) checkXrefTableOffset(offset int64) (int64, error) {
	// if the file is corrupt, it may not line up correctly
	// token := parser.reader.ReadToken()
	// if !bytes.Equal(token, Token("xref")) {
	// 	// bad PDF file! no cookie for you
	// 	// look to see if we can find the xref table nearby
	// 	fmt.Println("Corrupt PDF. Scanning for xref table")
	// 	parser.reader.Seek(-20, 1)
	// 	parser.reader.SkipToToken(Token("xref"))
	// 	token = parser.reader.ReadToken()
	// 	if !bytes.Equal(token, Token("xref")) {
	// 		return errors.New("Corrupt PDF: Could not find xref table")
	// 	}
	// }

	return offset, nil
}

func (parser *PDFParser) readXrefTable(offset int64) error {
	// fmt.Println("Reading xref table at", offset)

	// offset, err := parser.reader.checkXrefTable(offset)
	// if err != nil {
	// 	return err
	// }

	// first read in the Xref table data and the trailer dictionary
	if _, err := parser.reader.Seek(offset, 0); err != nil {
		return err
	}
	lines, err := parser.reader.ReadLinesToToken(Token("trailer"))
	if err != nil {
		fmt.Println("Error reading to end of xref table")
		return err
	}

	// trailer, err := parser.ReadValue().AsDictionary()

	// read the lines, store the xref table data
	start := 1
	if parser.xref.xrefLocation == 0 {
		parser.xref.maxObject = 0
		parser.xref.xrefLocation = offset
		parser.xref.xref = make(map[int]map[int]int, len(lines))
	}
	for _, lineBytes := range lines {
		// fmt.Println("Xref table line:", lineBytes)
		line := strings.TrimSpace(string(lineBytes))
		// fmt.Println("Reading xref table line:", line)
		if line != "" {
			if line == "xref" {
				continue
			}
			pieces := strings.Split(line, " ")
			switch len(pieces) {
			case 0:
				continue
			case 2:
				start, _ = strconv.Atoi(pieces[0])
				end, _ := strconv.Atoi(pieces[1])
				if end > parser.xref.maxObject {
					parser.xref.maxObject = end
				}
			case 3:
				if _, ok := parser.xref.xref[start]; !ok {
					parser.xref.xref[start] = make(map[int]int, len(lines))
				}
				xr, _ := strconv.Atoi(pieces[0])
				gen, _ := strconv.Atoi(pieces[1])

				if _, ok := parser.xref.xref[start][gen]; !ok {
					if pieces[2] == "n" {
						parser.xref.xref[start][gen] = xr
					} else {
						// xref[start][gen] = nil // ???
					}
				}
				start++
			default:
				return errors.New("Unexpected data in xref table: '" + line + "'")
			}
		}
	}

	// process the trailer
	// if parser.xref.trailer == nil {
	// 	parser.xref.trailer = trailer
	// }

	// fmt.Println("Xref table:", fmt.Sprintf("%v", parser.xref))

	return nil
}
