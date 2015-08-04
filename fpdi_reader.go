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
	"os"
	"bufio"
)

type PDFReader struct {
	file     *os.File
	buffered *bufio.Reader
	scanner  
	stack    []string
}

func NewFileReader(filename string) *PDFReader {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := &PDFReader{file, buf}
	reader.buffered = bufio.NewReader(file)
	reader.scanner = bufio.NewScanner(reader.buffered)

	reader.scanner.Split(splitPdfToken)
	reader.stack = make([]string, 0, 100)
	return reader
}

func (reader *PDFReader) Reset() {
	file.Seek(0)
	reader.buffered := bufio.NewReader(file)
}

func (reader *PDFReader) ReadToken() string {
	if token, ok := reader.popStack(); ok {
		return token
	}

	if reader.scanner.Scan() {
		return reader.scanner.Text()
	}
	return nil

	/*
	// Strip away any whitespace
	// ???

    // Get the first character in the stream
    b, err := reader.buffered.Read(1)
    c := string(b[:])
    switch c {
    case `[`:
    case `]`:
    case `(`:
    case `)`:
		// this is either an array or literal string delimiter, return it
		return c

	case `<`:
	case `>`:
		// this could be either a hex string of dictionary delimiter. determine which and return the token
		b2, err := reader.buffered.Peek(1)
		c2 := string(b2[:])
		if c2 == c {
			reader.Read(1)
			return c+c2
		} else {
			return c
		}

	case `%`:
		// this is a comment - jump over it!
		// ???
		reader.buffered.ReadBytes("\n")
		return reader.readToken()

	default:
		// this is another type of token. find the end and return it.
		// ???
		token := reader.buffered.ReadString(...)
		return token
    }
    */
}

func splitPdfToken(data []byte, atEOF bool) (advance int, token []byte, err error) {

}

func splitPdfNewline(data []byte, atEOF bool) (advance int, token []byte, err error) {

}

func splitPdfNewline(data []byte, atEOF bool) (advance int, token []byte, err error) {

}

func (parser *PDFParser) popStack() ([]byte, bool) {
	if len(pasrer.stack) == 0 {
		return nil, false
	}

	token := parser.stack[len(parser.stack)-1]
	parser.stack = parser.stack[:len(pasrer.stack)-1]
	return token, true
}

func (parser *PDFParser) pushStack(token []byte) {
	parser.stack = append(parser.stack, token)
}