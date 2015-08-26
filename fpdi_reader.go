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
	"bufio"
	"os"
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

// PDFReader is a low-level reader for the tokens in a PDF file
// See pdf_parser.php and pdf_context.php
type PDFReader struct {
	file     *os.File
	buffered *bufio.Reader
	scanner  *bufio.Scanner
	stack    []string
}

// NewFileReader constructs a reader for a given file
func NewFileReader(filename string) (*PDFReader, err) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := &PDFReader{file, buf}
	reader.buffered = bufio.NewReader(file)

	// scanner splits the file into tokens
	reader.scanner = bufio.NewScanner(reader.buffered)
	reader.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// see pdf_parser::_readToken in pdf_parser.php:726

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Strip away any whitespace
		var offset int
		for offset = 0; isPdfWhitespace(data[offset]); offset++ {
		}

		// Get the first character in the stream
		b = data[offset:1]
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
				reader.Read(1)
				return 2, data[offset:2], nil
			}
			return 1, c, nil

		case `%`:
			// This is a comment - jump over it!
			a, token, err := bufio.ScanLines(data, atEOF)
			return a, token, err

		default:
			// This is another type of token (probably a dictionary entry or a numeric value). Find the end and return it.
			var nextToken int
			for nextToken = offset; !isPdfWhitespaceOrBreak(data[nextToken]); nextToken++ {
			}
			return nextToken, data[offset : nextToken-offset], nil
		}
	})
	reader.stack = make([]string, 0, 100)
	return reader, nil
}

func isPdfWhitespace(b byte) bool {
	// \x20\x0A\x0C\x0D\x09\x00
	switch b {
	case 0x20:
	case 0x0A:
	case 0x0C:
	case 0x0D:
	case 0x09:
	case 0x00:
		return true
	default:
		return false
	}
}

func isPdfWhitespaceOrBreak(b byte) bool {
	// %[]<>()/
	switch b {
	case 0x20:
	case 0x0A:
	case 0x0C:
	case 0x0D:
	case 0x09:
	case 0x00:
	case `%`:
	case `[`:
	case `]`:
	case `<`:
	case `>`:
	case `(`:
	case `)`:
		return true
	default:
		return false
	}
}

// Reset a reader to the start of the file
func (reader *PDFReader) Reset() {
	file.Seek(0)
	reader.buffered = bufio.NewReader(file)
}

// ReadToken gets the next PDF token
func (reader *PDFReader) ReadToken() string {
	if token, ok := reader.popStack(); ok {
		return token
	}

	if reader.scanner.Scan() {
		return reader.scanner.Text()
	}
	return nil
}

//
func (parser *PDFParser) popStack() ([]byte, bool) {
	if len(pasrer.stack) == 0 {
		return nil, false
	}

	token := parser.stack[len(parser.stack)-1]
	parser.stack = parser.stack[:len(pasrer.stack)-1]
	return token, true
}

// push a token onto the stack so it
func (parser *PDFParser) pushStack(token []byte) {
	parser.stack = append(parser.stack, token)
}

// type PdfDictionary
