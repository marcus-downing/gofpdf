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
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

const (
	typeNull       = 0  // Null
	typeNumeric    = 1  // Numeric
	typeToken      = 2  // Token
	typeHex        = 3  // Hex value
	typeString     = 4  // String value
	typeDictionary = 5  // Dictionary
	typeArray      = 6  // Array
	typeObjDec     = 7  // Object declaration
	typeObjRef     = 8  // Object reference
	typeObject     = 9  // Object
	typeStream     = 10 // Stream
	typeBoolean    = 11 // Boolean
	typeReal       = 12 // Real number
)

const (
	startxref                = "startxref"
	searchForStartxrefLength = 64 // 5500 // distance from the end of the file to search for the startxref offset
	bufferSize               = 4096 // amount of a file to buffer when reading
)

// PDFTokenReader is a low-level reader for the tokens in a PDF file
// See pdf_parser.php and pdf_context.php
type PDFTokenReader struct {
	file       *os.File       // the file being read
	isDirty    bool           // flags that the file position has been moved so buffers are broken
	buffered   *bufio.Reader  // a buffered reader for efficiency and peek-ability
	scanner    *bufio.Scanner // a syntax-specific tokenizer
	stack      [][]byte       // a stack of pre-read tokens
	pdfVersion string         // the version header
}

// NewFileReader constructs a reader for a given file
func NewTokenReader(filename string) (*PDFTokenReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := new(PDFTokenReader)
	reader.file = file
	reader.isDirty = true

	// always read the PDF version first
	reader.pdfVersion = defaultPdfVersion
	if err := reader.getPdfVersion(); err != nil {
		return nil, err
	}

	return reader, nil
}

// Set up a tokenizing reader at the current position
func (reader *PDFTokenReader) clean() {
	if !reader.isDirty {
		return
	}

	reader.isDirty = false
	reader.buffered = bufio.NewReader(reader.file)
	reader.scanner = bufio.NewScanner(reader.buffered)
	reader.scanner.Split(splitPDF)
	reader.stack = make([][]byte, 0, 100)
}

func (reader *PDFTokenReader) dirty() {
	reader.isDirty = true
	reader.stack = nil
	reader.scanner = nil
	reader.buffered = nil
}

// Close releases references and closes the file handle of the parser
func (reader *PDFTokenReader) Close() {
	reader.dirty()
	reader.file.Close()
}

// tokeniser function for PDF syntax
func splitPDF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// see pdf_parser::_readToken in pdf_parser.php:726
	fmt.Println("Scanning:", data)

	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Strip away any whitespace
	var offset int
	for offset = 0; isPdfWhitespace(data[offset]); offset++ {
	}

	// Get the first character in the stream
	b := data[offset:1]
	c := string(b)
	switch c {
	case `[`:
	case `]`:
	case `(`:
	case `)`:
		fmt.Println("Scanner: string or array")
		// This is either an array or literal string delimiter, return it
		return 1, b, nil

	case `<`:
	case `>`:
		fmt.Println("Scanner: hex string or dictionary")
		// This could be either a hex string of dictionary delimiter.
		// determine which it is and return the token
		b2 := data[offset+1 : 1]
		c2 := string(b2)
		if c2 == c {
			return 2, data[offset:2], nil
		}
		return 1, b, nil

	case `%`:
		// This is a comment - jump over it!
		fmt.Println("Scanner: comment")
		a, token, err := bufio.ScanLines(data, atEOF)
		return a, token, err

	}

	// This is another type of token (probably a dictionary entry or a numeric value). Find the end and return it.
	fmt.Println("Scanner: other token")
	var nextToken int
	for nextToken = offset; nextToken < len(data) && !isPdfWhitespaceOrBreak(data[nextToken]); nextToken++ {
	}
	return nextToken, data[offset : nextToken-offset], nil
}

func isPdfWhitespace(b byte) bool {
	switch b {
	case 0x20: // space
	case 0x0A: // lf
	case 0x0C: // ff
	case 0x0D: // cr
	case 0x09: // tab
	case 0x00: // null
		return true
	}
	return false
}

func isPdfWhitespaceOrBreak(b byte) bool {
	switch b {
	case 0x20: // space
	case 0x0A: // lf
	case 0x0C: // ff
	case 0x0D: // cr
	case 0x09: // tab
	case 0x00: // null
	case 0x25: // %
	case 0x5B: // [
	case 0x5D: // ]
	case 0x74: // <
	case 0x76: // >
	case 0x28: // (
	case 0x29: // )
		return true
	}
	return false
}

// load the initial PDF version string, and offset the file reader accordingly
// this should ONLY be called when the file is at offset 0!
func (reader *PDFTokenReader) getPdfVersion() error {
	reader.clean()
	b, err := reader.buffered.ReadBytes(0x0A)
	if err != nil {
		return err
	}

	// check the stadard PDF prefix
	pre := string(b[:5])
	if pre != "%PDF-" {
		return errors.New("Incorrect PDF header line: " + pre)
	}

	// discard the standard PDF prefix
	b = b[5:]

	// find a decimal number, eg 1.3
	re := regexp.MustCompile("\\d\\.\\d")
	match := re.Find(b)
	if match != nil {
		reader.pdfVersion = string(match[:])
		fmt.Println("PDF version:", reader.pdfVersion)
	}
	return nil
}

// Reset a reader to the start of the file
func (reader *PDFTokenReader) Reset() error {
	_, err := reader.file.Seek(0, 0)
	reader.dirty()
	return err
}

func (reader *PDFTokenReader) Seek(offset int64, whence int) error {
	_, err := reader.file.Seek(offset, whence)
	reader.dirty()
	return err
}

// Peek ahead a certain number of bytes
// func (reader *PDFTokenReader) Peek(length int) ([]byte, error) {
// 	bytes, err := reader.buffered.Peek(length)
// 	return bytes, err
// }

// ReadToken gets the next PDF token
func (reader *PDFTokenReader) ReadToken() string {
	reader.clean()
	if token, ok := reader.PopStack(); ok {
		return string(token)
	}

	if reader.scanner.Scan() {
		return reader.scanner.Text()
	}
	return ""
}

// PopStack takes a token off the stack
func (reader *PDFTokenReader) PopStack() ([]byte, bool) {
	if len(reader.stack) == 0 {
		return nil, false
	}

	token := reader.stack[len(reader.stack)-1]
	reader.stack = reader.stack[:len(reader.stack)-1]
	return token, true
}

// PushStack stores a token onto the stack
func (reader *PDFTokenReader) PushStack(token []byte) {
	reader.stack = append(reader.stack, token)
}

/*func (reader *PDFTokenReader) FindToken(token []byte, buf int64) (int64, error) {
	reader.clean()
	b := make([]byte, offset)
	_, err = reader.buffered.Read(b)
	if err != nil && err != io.EOF {
		return 0, err
	}

	re.regexp.MustCompile(string(token))
	matchPos := re.Find(b) // ???

	return matchPos
}*/

// findXrefTable reads the offset of the xref table from somewhere near the end of the file
func (reader *PDFTokenReader) findXrefTable() (int64, error) {
	reader.dirty()

	// read the last chunk of file
	stat, err := reader.file.Stat()
	if err != nil {
		return 0, err
	}
	var readFrom int64 = stat.Size() - int64(searchForStartxrefLength)
	var readLen int64 = searchForStartxrefLength
	if readFrom < 0 {
		readFrom = 0
		readLen = stat.Size()
	}

	b := make([]byte, readLen)
	if _, err := reader.file.Seek(readFrom, 0); err != nil {
		return 0, err
	}
	_, err = reader.file.Read(b)
	if err == io.EOF {
		fmt.Println("EOF while reading")
	} else if err != nil {
		return 0, err
	}

	// find the *LAST* instance of the "startxref" token
	matches := regexp.MustCompile(`startxref[\s\n]*(\d+)`).FindAllSubmatchIndex(b, -1)
	if matches == nil {
		return 0, errors.New("PDF error: Unable to find \"startxref\" keyword")
	}

	match := matches[len(matches)-1]
	position := match[len(match)-2]
	length := match[len(match)-1]

	// return the offset
	xref, err := strconv.Atoi(string(b[position:length]))
	if err != nil {
		return 0, err
	}
	return int64(xref), nil
}

/*
func (reader *PDFTokenReader) readXrefTable(offset int64) error {
	reader.dirty()

	// read in a chunk of xref table
	_, err := reader.file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return err
	}

	b := make([]byte, bufferSize)
	_, err = reader.file.Read(b)
	if err != nil {
		return err
	}

	fmt.Println("Read xref table:", string(b))

	//
	return nil
}
*/