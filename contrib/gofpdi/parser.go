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
	"os"
	"strings"
	// "regexp"
	"math"
	"bytes"
	"strconv"
	// "bufio"
	"errors"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi/readpdf"
	. "github.com/jung-kurt/gofpdf/contrib/gofpdi/types"
)

// OpenPDFParser opens an existing PDF file and readies it
func OpenPDFParser(filename string) (*PDFParser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// read the last chunk of file
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	filesize := stat.Size()

	// fmt.Println("Opening PDF file:", filename)
	reader, err := readpdf.NewTokenReader(file, filesize)
	if err != nil {
		return nil, err
	}

	parser := new(PDFParser)
	parser.reader = reader
	parser.pageNumber = 0
	parser.lastUsedPageBox = DefaultBox

	// read xref data
	offset, err := parser.reader.FindXrefTable()
	if err != nil {
		return nil, err
	}
	parser.readXrefTable(offset)

	// check for encryption
	// parser.checkEncryption()

	// read root
	// pagesDictionary := parser.resolveObject("/Pages")

	if err := parser.Error(); err != nil {
		return nil, err
	}
	return parser, nil
}

// PDFParser is a high-level parser for PDF elements
// See fpdf_pdf_parser.php
type PDFParser struct {
	reader          *readpdf.PDFTokenReader // the underlying token reader
	pageNumber      int             // the current page number
	lastUsedPageBox string          // the most recently used page box
	pages           []PDFPage       // already loaded pages

	xref struct {
		maxObject    int                 // the highest xref object number
		xrefLocation int64               // the location of the xref table
		xref         map[ObjectRef]int64 // all the xref offsets
		trailer      Dictionary          // trailer data
	}

	currentObject *ObjectDeclaration // the most recently read object
	warnings      []error    // non-fatal errors encountered during reading
	err           error      // the first fatal error encountered
}

// SetWarning notes a non-fatal parser error
func (parser *PDFParser) SetWarning(warning error) {
	parser.warnings = append(parser.warnings, warning)
}

// SetError notes a fatal parser error
func (parser *PDFParser) SetError(err error) {
	if parser.err == nil {
		parser.err = err
	}
}

// Error returns the internal parser error; this will be nil if no error has occurred.
func (parser *PDFParser) Error() error {
	return parser.err
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

	// page = parser.resolveObject(page)

	if box, ok := page.Dictionary[boxIndex]; ok {
		if boxRef, ok := box.(ObjectRef); ok {
			box = parser.resolveObject(boxRef)
		}

		if arr, ok := box.(Array); ok {
			x := arr[0].ToReal().ToFloat64()
			y := arr[1].ToReal().ToFloat64()
			x2 := arr[2].ToReal().ToFloat64()
			y2 := arr[3].ToReal().ToFloat64()
			w := math.Abs(x2 - x)
			h := math.Abs(y2 - y)
			llx := math.Min(x, x2)
			lly := math.Min(y, y2)
			urx := math.Max(x, x2)
			ury := math.Max(y, y2)

			return &PageBox{
				gofpdf.PointType{x / k, y / k},
				gofpdf.SizeType{w / k, h / k},
				gofpdf.PointType{llx / k, lly / k},
				gofpdf.PointType{urx / k, ury / k},
			}
		}

	}
	if parent, ok := page.Dictionary["/Parent"]; ok {
		parent := parser.resolveObject(parent)
		parentPage := PDFPage{parent, 0}
		return parser.getPageBox(parentPage, boxIndex, k)
	}
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

func (parser *PDFParser) readXrefTable(offset int64) {
	// fmt.Println("Reading xref table at", offset)

	// offset, err := parser.reader.checkXrefTable(offset)
	// if err != nil {
	// 	return err
	// }

	// first read in the Xref table data and the trailer dictionary
	if _, err := parser.reader.Seek(offset, 0); err != nil {
		fmt.Println("Reading xref table at", offset)
		parser.SetError(err)
	}
	lines, ok := parser.reader.ReadLinesToToken(Token("trailer"))
	if !ok {
		fmt.Println("Reading xref table at", offset)
		// fi, _ := parser.reader.FileStat()
		// fmt.Printf("The file is %d bytes long\n", fi.Size())

		// fmt.Println("Read lines to token 'trailer'", lines)
		err := errors.New("Cannot read end of xref table")
		parser.SetError(err)
	}

	// read the lines, store the xref table data
	start := 1
	if parser.xref.xrefLocation == 0 {
		parser.xref.maxObject = 0
		parser.xref.xrefLocation = offset
		parser.xref.xref = make(map[ObjectRef]int64, len(lines))
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
				// if _, ok := parser.xref.xref[start]; !ok {
				// 	parser.xref.xref[start] = make(map[int]int, len(lines))
				// }
				xr, _ := strconv.ParseInt(pieces[0], 10, 64)
				gen, _ := strconv.Atoi(pieces[1])

				ref := ObjectRef{start, gen}
				if _, ok := parser.xref.xref[ref]; !ok {
					if pieces[2] == "n" {
						parser.xref.xref[ref] = xr
					} else {
						// xref[ref] = nil // ???
					}
				}
				start++
			default:
				err := errors.New("Unexpected data in xref table: '" + line + "'")
				parser.SetError(err)
			}
		}
	}

	// read the trailer dictionary
	trailerPeekTokens := parser.reader.PeekTokens(16)
	parser.reader.SkipToken()
	// trailerPeek := parser.reader.Peek(100)
	trailerValue := parser.readValue()

	trailer, ok := trailerValue.(Dictionary)
	if !ok {
		// fmt.Println("Trailer peek:", string(trailerPeek))
		fmt.Println("Trailer peek tokens:", trailerPeekTokens)
		fmt.Println("Trailer:", trailerValue)
		err := errors.New("Trailer not a dictionary")
		parser.SetError(err)
	}

	// process the trailer
	if parser.xref.trailer == nil {
		// fmt.Println("Storing trailer data:", trailer)
		parser.xref.trailer = trailer
	}

	// fmt.Println("Xref table:", fmt.Sprintf("%v", parser.xref))

	// return nil
}

// readValue reads the next value from the PDF
func (parser *PDFParser) readValue() Value {
	token := parser.reader.ReadToken()
	if token == nil {
		return nil
	}

	str := token.String()
	switch str {
	case "<":
		// This is a hex value
		// Read the value, then the terminator
		bytes, _ := parser.reader.ReadBytesToToken(Token(">"))
		fmt.Println("Read hex:", bytes)
		return Hex(bytes)

	case "<<":
		// This is a dictionary.
		// Recurse into this function until we reach
		// the end of the dictionary.
		result := make(map[string]Value, 32)
		for key := parser.reader.ReadToken(); !key.Equals(Token(">>")); key = parser.reader.ReadToken() {
			if key == nil {
				return nil // ?
			}
			value := parser.readValue()
			if value == nil {
				return nil // ?
			}
			// Catch missing value
			if value.Equals(Token(">>")) {
				result[key.String()] = value
				break
			}
			result[key.String()] = value
		}
		return Dictionary(result)

	case "[":
		// This is an array.
		// Recurse into this function until we reach
		// the end of the array.
		result := make([]Value, 0, 32)
		for {
			value := parser.readValue()
			if value.Equals(Token("]")) {
				break
			}
			result = append(result, value)
		}
		return Array(result)

	case "(":
		// This is a string
		openBrackets := 1
		buf := bytes.NewBuffer([]byte{})
		for openBrackets > 0 {
			b, ok := parser.reader.ReadByte()
			if !ok {
				break
			}
			switch b {
			case 0x28: // (
				openBrackets++
			case 0x29: // )
				openBrackets++
			case 0x5C: // \
				b, ok = parser.reader.ReadByte()
				if !ok {
					break
				}
			}
			buf.WriteByte(b)
		}
		return String(buf.Bytes())

	case "stream":
		// ensure line breaks in front of the stream
		peek := parser.reader.Peek(32)
		for _, c := range peek {
			if !readpdf.IsPdfWhitespace(c) {
				break
			}
			parser.reader.ReadByte()
		}

		// TODO get the stream length
		lengthObj := parser.currentObject.GetParam("/Length")
		if lengthObj.Type() == TypeObjRef {
			lengthObj = parser.resolveObject(lengthObj)
		}
		length := int(lengthObj.ToNumeric().ToInt64()) // lengthObj[1] ???

		stream, _ := parser.reader.ReadBytes(length)

		if endstream := parser.reader.ReadToken(); endstream.Equals(Token("endstream")) {
			// We don't throw an error here because the next
			// round trip will start at a new offset
		}

		return Stream{parser.currentObject.GetDictionary(), stream}
	}

	if real, err := strconv.ParseFloat(str, 64); err != nil {
		return Real(real)
	}
	if number, err := strconv.Atoi(str); err != nil {
		// A numeric token. Make sure that
		// it is not part of something else.
		if moreTokens := parser.reader.PeekTokens(2); len(moreTokens) == 2 {
			if number2, err := strconv.Atoi(string(moreTokens[0])); err != nil {
				// Two numeric tokens in a row.
				// In this case, we're probably in
				// front of either an object reference
				// or an object specification.
				// Determine the case and return the data
				switch string(moreTokens[1]) {
				case "obj":
					parser.reader.ReadTokens(2)
					objectRef := ObjectRef{number, number2}
					values := make([]Value, 0, 2) // ???
					// values := Dictionary{}
					return ObjectDeclaration{objectRef, values}
				case "R":
					parser.reader.ReadTokens(2)
					return ObjectRef{number, number2}
				}
			}
		}

		return Numeric(number)
	}
	if str == "true" {
		return Boolean(true)
	}
	if str == "false" {
		return Boolean(false)
	}
	if str == "null" {
		return Null(struct{}{})
	}
	// Just a token. Return it.
	return token
}

func (parser *PDFParser) resolveObject(spec Value) Dictionary {
	// Exit if we get invalid data
	if spec == nil {
		return nil
	}

	if objRef, ok := spec.(ObjectRef); ok {
		// This is a reference, resolve it
		if offset, ok := parser.xref.xref[objRef]; ok {
			parser.reader.Seek(offset, 0)
			header := parser.readValue()
			if header != objRef {
				toSearchFor := Token(fmt.Sprintf("%d %d obj", objRef.Obj, objRef.Gen))
				if parser.reader.SkipToToken(toSearchFor) {
					parser.reader.SkipBytes(len(toSearchFor))
				} else {
					// Unable to find object
					return nil
				}
			}

			if headerRef, ok := header.(ObjectRef); ok {
				// If we're being asked to store all the information
				// about the object, we add the object ID and generation
				// number for later use
				result := ObjectDeclaration{headerRef, make([]Value, 0, 2)}
				parser.currentObject = &result

				// Now simply read the object data until
				// we encounter an end-of-object marker
				for {
					value := parser.readValue()
					if value == nil || len(result.Values) > 1 { // ???
						// in this case the parser couldn't find an "endobj" so we break here
						break
					}

					if value.Equals(Token("endobj")) {
						break
					}

					result.Values = append(result.Values, value)
				}
			} else {
				// ?
			}

			// Reset to the start
			// parser.reader.Seek(???)

		} else {
			// Unable to find object
			return nil
		}
	}

	if obj, ok := spec.(Dictionary); ok {
		return obj
	}

	// Er, it's a what now?
	parser.SetWarning(fmt.Errorf("Attempt to resolve unknown value as object spec: %v", spec))
	return nil
}

func (parser *PDFParser) unfilterStream(s Stream) Stream {
	filters := make([]StreamFilter, 8)
	if filter, ok := s.Parameters["/Filter"]; ok {
		if objRef, ok := filter.(ObjectRef); ok {
			filter = parser.resolveObject(objRef)
		}

		if filterName, ok := filter.(Token); ok {
			streamFilter := GetStreamFilter(filterName.String())
			filters = append(filters, streamFilter)
		} else if array, ok := filter.(Array); ok {
			filterName := array[0].ToString().String()
			streamFilter := GetStreamFilter(filterName)
			filters = append(filters, streamFilter)
		}
		// return s
	}
	// I appear to be missing something...

	var stream Stream = s
	for _, filter := range filters {
		stream = filter(stream)
	}
	return s
}
