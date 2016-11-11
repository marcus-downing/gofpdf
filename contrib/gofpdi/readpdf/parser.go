package readpdf

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
	// "math"
	"bytes"
	"strconv"
	// "bufio"
	"errors"
	// "github.com/jung-kurt/gofpdf"
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
	reader, err := NewTokenReader(file, filesize)
	if err != nil {
		return nil, err
	}

	parser := new(PDFParser)
	parser.reader = reader
	// parser.pageNumber = 0
	parser.lastUsedPageBox = DefaultBox

	// read xref data
	offset, err := parser.reader.FindXrefTable()
	if err != nil {
		return nil, err
	}
	parser.readXrefTable(offset)
	// fmt.Printf("Xref trailer: %v\n", parser.xref.trailer)

	// check for encryption
	if _, ok := parser.xref.trailer[EncryptRef]; ok {
		parser.SetError(errors.New("File is encrypted!"))
	}

	// read root object
	rootRef := parser.xref.trailer[RootRef]
	// fmt.Printf("Root ref: %v\n", rootRef)
	rootObj := parser.resolveObject(rootRef)
	if rootObj == nil {
		parser.SetError(errors.New("Cannot load root object!"))
	} else {
		parser.root = *rootObj
	}
	// fmt.Printf("Root: %v\n", parser.root)

	// read pages
	// pagesRef := 
	// pagesDictionary := parser.resolveObject(PageRef)

	if err := parser.Error(); err != nil {
		return nil, err
	}
	return parser, nil
}

// PDFParser is a high-level parser for PDF elements
// See fpdf_pdf_parser.php
type PDFParser struct {
	reader          *PDFTokenReader // the underlying token reader
	// pageNumber      int             // the current page number
	lastUsedPageBox string          // the most recently used page box
	// pages           []*PDFPageParser       // already loaded pages
	root            ObjectDeclaration      // the root object

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

func (parser *PDFParser) GetPDFVersion() string {
	return parser.reader.PdfVersion
}

// Close releases references and closes the file handle of the parser
func (parser *PDFParser) Close() {
	parser.reader.Close()
}

func (parser *PDFParser) getRootObject() *ObjectDeclaration {
	return &parser.root
}

func (parser *PDFParser) getCatalogObject() *ObjectDeclaration {
	return &parser.root
	// if catalogRef, ok := parser.root.Get(CatalogRef); ok {
	// 	return parser.resolveObject(catalogRef)
	// } else {
	// 	fmt.Println("getCatalogObject: no catalog value in root:", parser.root)
	// }
	// return nil
}

func (parser *PDFParser) getPagesObject() *ObjectDeclaration{
	if catalog := parser.getCatalogObject(); catalog != nil {
		if pagesRef, ok := catalog.Get(PagesRef); ok {
			return parser.resolveObject(pagesRef)
		}
	} else {
		fmt.Println("getPagesObject: no catalog")
	}
	return nil
}

func (parser *PDFParser) CountPages() int {
	if pagesObj := parser.getPagesObject(); pagesObj != nil {
		if count, ok := pagesObj.Get(CountRef); ok && count != nil {
			return int(count.ToNumeric().ToInt64())
		} else {
			fmt.Println("CountPages: no count value")
		}
	} else {
		fmt.Println("CountPages: no pages object")
	}
	parser.SetError(errors.New("Unable to read page count"))
	return 0
}

// GetPageParser constucts a PDFPageParser for the given page number
func (parser *PDFParser) GetPageParser(number int) *PDFPageParser {
	// find the object associated with this page

	if pagesObj := parser.getPagesObject(); pagesObj != nil {
		if kids, ok := pagesObj.Get(KidsRef); ok && kids != nil {
			if kidsArray, ok := kids.(Array); ok {
				if number <= len(kidsArray) {
					pageRef := kidsArray[number - 1]
					pageObj := parser.resolveObject(pageRef)
					if pageObj != nil {
						page := PDFPageParser{pageObj, parser, number}
						return &page
					}
				}
			}
		}
	}

	parser.SetError(errors.New("Unable to read page object"))
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
		// fmt.Println("Read hex:", bytes)
		return Hex(bytes)

	case "<<":
		// This is a dictionary.
		// Recurse into this function until we reach
		// the end of the dictionary.
		result := make(map[string]Value, 32)
		// fmt.Println("Reading dictionary")
		for key := parser.reader.ReadToken(); !key.Equals(Token(">>")); key = parser.reader.ReadToken() {
			if key == nil {
				return nil // ?
			}
			// fmt.Println("Next dictionary value peek:", string(parser.reader.Peek(20)))
			value := parser.readValue()
			if value == nil {
				return nil // ?
			}
			// fmt.Printf("Storing dictionary value: %s = %v\n", key.String(), value)
			// Catch missing value
			if value.Equals(Token(">>")) {
				result[key.String()] = value
				break
			}
			result[key.String()] = value
		}
		// fmt.Println("Dictionary:", result)
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
			if !IsPdfWhitespace(c) {
				break
			}
			parser.reader.ReadByte()
		}

		// TODO get the stream length
		var length int = 0
		if lengthObj, ok := parser.currentObject.Get(LengthRef); ok && lengthObj.Type() == TypeObjRef {
			lengthObj = parser.resolveObject(lengthObj)
			length = int(lengthObj.ToNumeric().ToInt64()) // lengthObj[1] ???
		} else {
			parser.SetError(errors.New("Stream has no length"))
		}
		stream, _ := parser.reader.ReadBytes(length)

		if endstream := parser.reader.ReadToken(); endstream.Equals(Token("endstream")) {
			// We don't throw an error here because the next
			// round trip will start at a new offset
		}

		return Stream{parser.currentObject.GetDictionary(), stream}
	}

	// fmt.Println("Parsing token in case it's a number:", str)
	if number, err := strconv.Atoi(str); err == nil {
		// fmt.Println("First number token:", number)
		// fmt.Println("More tokens:", parser.reader.PeekTokens(2))
		// A numeric token. Make sure that
		// it is not part of something else.
		if moreTokens := parser.reader.PeekTokens(2); len(moreTokens) == 2 {
			if number2, err := strconv.Atoi(string(moreTokens[0])); err == nil {
				// fmt.Println("Second number token:", number2)
				// fmt.Println("Third token", moreTokens[1])
				// Two numeric tokens in a row.
				// In this case, we're probably in
				// front of either an object reference
				// or an object specification.
				// Determine the case and return the data
				switch string(moreTokens[1]) {
				case "obj":
					parser.reader.ReadTokens(2)
					objectRef := ObjectRef{number, number2}
					// fmt.Println("Reading values for object:", objectRef)
					values := make([]Value, 0, 2) // ???
					// dictionary := parser.readValue()
					// values := Dictionary{}
					for i := 1; i < 10; i++ {
						value := parser.readValue()
						// fmt.Println("Value:", value)
						if value == nil {
							break
						}
						if tokenValue, ok := value.(Token); ok && tokenValue.String() == "endobj" {
							break
						}
						values = append(values, value)
					}
					return ObjectDeclaration{objectRef, values}
				case "R":
					parser.reader.ReadTokens(2)
					return ObjectRef{number, number2}
				}
			}
		}

		// fmt.Println("Numeric value:", number)
		return Numeric(number)
	}
	if real, err := strconv.ParseFloat(str, 64); err == nil {
		// fmt.Println("Real value:", real)
		return Real(real)
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
	// fmt.Println("Token value", token)
	return token
}

func (parser *PDFParser) resolveObject(spec Value) *ObjectDeclaration {
	// Exit if we get invalid data
	if spec == nil {
		return nil
	}

	if objRef, ok := spec.(ObjectRef); ok {
		// This is a reference, resolve it
		if offset, ok := parser.xref.xref[objRef]; ok {
			// fmt.Println("Resolve ref at offset:", offset)
			parser.reader.Seek(offset, 0)
			// fmt.Println("Peek:", string(parser.reader.Peek(100)))
			obj := parser.readValue()
			// fmt.Println("Reader object value:", obj)
			if objDecl, ok := obj.(ObjectDeclaration); ok {
				return &objDecl
			}

			// so it isn't an object declaration? now what?

			/*
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
*/
			// Reset to the start
			// parser.reader.Seek(???)

		} else {
			// Unable to find object
			return nil
		}
	}

	if obj, ok := spec.(ObjectDeclaration); ok {
		return &obj
	}

	// Er, it's a what now?
	parser.SetWarning(fmt.Errorf("Attempt to resolve unknown value as object spec: %v", spec))
	return nil
}

func (parser *PDFParser) unfilterStream(s Stream) Stream {
	filters := make([]StreamFilter, 8)
	if filter, ok := s.Parameters[FilterRef]; ok {
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
	}

	var stream Stream = s
	for _, filter := range filters {
		stream = filter(stream)
	}
	return s
}
