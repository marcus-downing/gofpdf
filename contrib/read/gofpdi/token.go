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
	"os"

	// "github.com/jung-kurt/gofpdf"
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

// PDFTokenReader is a low-level reader for the tokens in a PDF file
// See pdf_parser.php and pdf_context.php
type PDFTokenReader struct {
	file     *os.File
	buffered *bufio.Reader
	scanner  *bufio.Scanner
	stack    [][]byte
}

// NewFileReader constructs a reader for a given file
func NewTokenReader(filename string) (*PDFTokenReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := new(PDFTokenReader)
	reader.file = file
	reader.init()

	// scanner splits the file into tokens
	return reader, nil
}

// Set up a new PDF reader
func (reader *PDFTokenReader) init() {
	reader.buffered = bufio.NewReader(reader.file)
	reader.scanner = bufio.NewScanner(reader.buffered)
	reader.scanner.Split(splitPDF)
	reader.stack = make([][]byte, 0, 100)
}

func splitPDF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// see pdf_parser::_readToken in pdf_parser.php:726

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
		// This is either an array or literal string delimiter, return it
		return 1, b, nil

	case `<`:
	case `>`:
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
		a, token, err := bufio.ScanLines(data, atEOF)
		return a, token, err

	}

	// This is another type of token (probably a dictionary entry or a numeric value). Find the end and return it.
	var nextToken int
	for nextToken = offset; !isPdfWhitespaceOrBreak(data[nextToken]); nextToken++ {
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

// Reset a reader to the start of the file
func (reader *PDFTokenReader) Reset() {
	reader.file.Seek(0, 0)
	reader.init()
}

// Peek ahead a certain number of bytes
func (reader *PDFTokenReader) Peek(length int) ([]byte, error) {
	bytes, err := reader.buffered.Peek(length)
	return bytes, err
}

// ReadToken gets the next PDF token
func (reader *PDFTokenReader) ReadToken() string {
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

// type PdfDictionary
