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
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

const (
	startxref                = "startxref"
	searchForStartxrefLength = 1024 // 5500 // distance from the end of the file to search for the startxref offset
)

// PDFTokenReader is a low-level reader for the tokens in a PDF file
// See pdf_parser.php and pdf_context.php
type PDFTokenReader struct {
	file       *os.File       // the file being read
	scanner    *bufio.Scanner // a syntax-specific tokenizer
	stack      []Token        // a stack of pre-read tokens
	lineEnding []byte         // the type of line ending this file uses
	pdfVersion string         // the version header
}

// NewTokenReader constructs a low level reader for a PDF file
func NewTokenReader(filename string) (*PDFTokenReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := new(PDFTokenReader)
	reader.file = file
	reader.calibrateLineEndings()
	reader.scanner = bufio.NewScanner(reader.file)
	reader.lineEnding = []byte("\n")

	// always read the PDF version first
	reader.pdfVersion = defaultPdfVersion
	if err := reader.getPdfVersion(); err != nil {
		return nil, err
	}

	return reader, nil
}

// Close releases references and closes the file handle of the parser
func (reader *PDFTokenReader) Close() {
	reader.scanner = nil
	reader.file.Close()
}

func (reader *PDFTokenReader) splitOnLines() {
	reader.scanner.Split(splitLines(reader.lineEnding))
}

func (reader *PDFTokenReader) splitPeek(n int, into *[]byte) {
	reader.scanner.Split(splitPeek(n, into))
}

func (reader *PDFTokenReader) splitNext(n int) {
	reader.scanner.Split(splitNext(n))
}

func (reader *PDFTokenReader) splitUntil(token Token) {
	reader.scanner.Split(splitUntil(token))
}

func (reader *PDFTokenReader) splitPeekTokens(n int, into *[]Token) {
	reader.scanner.Split(splitPeekTokens(n, into))
}

func (reader *PDFTokenReader) splitOnTokens() {
	reader.scanner.Split(splitTokens)
}

func (reader *PDFTokenReader) splitOnBytes() {
	reader.scanner.Split(splitBytes)
}

// calibrateLineEndings uses some sample data to determine which line-ending is used by this file
func (reader *PDFTokenReader) calibrateLineEndings() {
	sample := make([]byte, 100)
	reader.file.Read(sample)
	rx := regexp.MustCompile("\r\n|\n|\r")
	match := rx.Find(sample)
	if match != nil {
		reader.lineEnding = match
	}
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// splitLines is like bufio.SplitLines, but uses our calibrated line ending
func splitLines(lineEnding []byte) func([]byte, bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, io.EOF
		}
		if i := bytes.Index(data, lineEnding); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, dropCR(data[0:i]), nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), dropCR(data), nil
		}
		// Request more data.
		return 0, nil, nil
	}
}

// splitPeek reads up to a fixed number of bytes, but does not progress the buffer
func splitPeek(n int, into *[]byte) func([]byte, bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, io.EOF
		}

		// make sure we have enough loaded
		if len(data) < n && !atEOF {
			// request more data
			return 0, nil, nil
		}

		// return as much as we can
		if n > len(data) {
			n = len(data)
		}
		result := make([]byte, n)
		copy(result, data)
		*into = result
		return 0, nil, io.EOF
	}
}

// splitNext reads up to a fixed number of bytes.
func splitNext(n int) func([]byte, bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, io.EOF
		}

		// make sure we have enough loaded
		if len(data) < n && !atEOF {
			// request more data
			return 0, nil, nil
		}

		// return as much as we can
		if n > len(data) {
			n = len(data)
		}
		return n, data[0:n], io.EOF
	}
}

// splitUntil returns variable sized chunks until it reaches the desired token
func splitUntil(token Token) func([]byte, bool) (advance int, token []byte, err error) {
	tokenRx := regexp.MustCompile(token.ToRegex())
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// not enough data
		if atEOF && len(data) == 0 {
			// os.Stderr.WriteString("splitUntil: the end\n")
			return 0, nil, io.EOF
		}
		if len(data) < len(token) {
			if atEOF {
				// os.Stderr.WriteString("splitUntil: EOF\n")
				return 0, nil, io.EOF
			}
			// os.Stderr.WriteString("splitUntil: more please\n")
			return 0, nil, nil
		}
		// os.Stderr.WriteString(fmt.Sprintf("splitUntil: looking in: %s (atEOF %v)\n", string(data), atEOF))
		if match := tokenRx.FindIndex(data); match != nil {
			// os.Stderr.WriteString(fmt.Sprintf("Found: %v\n", match))
			if match[0] == 0 {
				// found = true
				// os.Stderr.WriteString("splitUntil: done\n")
				return 0, nil, io.EOF
			}
			// os.Stderr.WriteString("splitUntil: nearly there\n")
			return match[0], data[0:match[0]], nil
		}
		// os.Stderr.WriteString("splitUntil: proceed\n")
		shift := len(data) - len(token)
		return shift, data[0:shift], nil
	}
}

// splitBytes returns every single byte
func splitBytes(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, io.EOF
	}
	return 1, data[0:1], nil
}

// splitPeekTokens returns
func splitPeekTokens(n int, into *[]Token) func([]byte, bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, io.EOF
		}

		// split tokens until we have the right number or run out of data
		result := make([]Token, 0, n)
		for len(result) < n {
			padvance, ptoken, perr := splitTokens(data, atEOF)
			if padvance == 0 && ptoken == nil && perr == nil {
				// request more data
				return 0, nil, nil
			}

			if ptoken != nil {
				result = append(result, ptoken)
			}

			data = data[padvance:]
			if atEOF && len(data) == 0 {
				return 0, nil, io.EOF
			}

			if perr == io.EOF {
				// stop reading
				break
			}
		}
		*into = result

		return 0, nil, io.EOF
	}
}

// splitTokens returns one token at a time
func splitTokens(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// see pdf_parser::_readToken in pdf_parser.php:726

	if atEOF && len(data) == 0 {
		return 0, nil, io.EOF
	}

	// Strip away any whitespace
	var offset int
	for offset = 0; offset < len(data) && isPdfWhitespace(data[offset]); offset++ {
	}

	// if we're at the end of the data, signal that with an EOF
	if offset >= len(data) {
		if atEOF {
			return offset, nil, io.EOF
		}
		return offset, nil, nil
	}
	// os.Stderr.WriteString(fmt.Sprintf("Data length: %d bytes, offset: %d, data: %v\n", len(data), offset, data[0:6]))

	// Get the first character in the stream
	b := data[offset]
	switch b {
	case 0x28, // (
		0x29, // )
		0x5B, // [
		0x5D: // ]
		// os.Stderr.WriteString("Scanner: string or array\n")
		// This is either an array or literal string delimiter, return it
		return 1, []byte{b}, nil

	case 0x3C, // <
		0x3E: // >
		// os.Stderr.WriteString("Scanner: hex string or dictionary\n")
		// This could be either a hex string of dictionary delimiter.
		// determine which it is and return the token
		b2 := data[offset+1]
		if b2 == b {
			return 2, data[offset:2], nil
		}
		return 1, []byte{b}, nil

	case 0x25: // %
		// This is a comment - jump over it!
		// os.Stderr.WriteString("Scanner: comment\n")
		a, token, err := bufio.ScanLines(data, atEOF)
		return a, token, err
	}

	// This is another type of token (probably a dictionary entry or a numeric value). Find the end and return it.
	// os.Stderr.WriteString("Scanner: other token\n")
	var nextToken int
	for nextToken = offset; nextToken < len(data) && !isPdfWhitespaceOrBreak(data[nextToken]); nextToken++ {
		// os.Stderr.WriteString(fmt.Sprintf("Skipped char: %v\n", data[nextToken]))
	}
	token = data[offset:nextToken]
	// os.Stderr.WriteString(fmt.Sprintf("Token: advance = %d, token = %s, err = %v\n", nextToken, string(token), err))
	return nextToken, token, nil
}

func isPdfWhitespace(b byte) bool {
	switch b {
	case 0x00, // null
		0x09, // tab
		0x0A, // lf
		0x0C, // ff
		0x0D, // cr
		0x20: // space
		return true
	}
	return false
}

func isPdfWhitespaceOrBreak(b byte) bool {
	switch b {
	case 0x00, // null
		0x09, // tab
		0x0A, // lf
		0x0C, // ff
		0x0D, // cr
		0x20, // space
		0x25, // %
		0x28, // (
		0x29, // )
		0x5B, // [
		0x5D, // ]
		0x3C, // <
		0x3E: // >
		return true
	}
	return false
}

// load the initial PDF version string, and offset the file reader accordingly
func (reader *PDFTokenReader) getPdfVersion() error {
	reader.Reset()
	reader.splitOnLines()
	if !reader.scanner.Scan() {
		return reader.scanner.Err()
	}
	b := reader.scanner.Bytes()

	// check the stadard PDF prefix
	pre := string(b[:5])
	if pre != "%PDF-" {
		return errors.New("Incorrect PDF header line: " + pre)
	}
	b = b[5:]

	// find a decimal number, eg 1.3
	re := regexp.MustCompile("\\d\\.\\d")
	match := re.Find(b)
	if match != nil {
		reader.pdfVersion = string(match[:])
	}
	return nil
}

// SplitLines separates data into lines using the calibrated line ending
func (reader *PDFTokenReader) SplitLines(data []byte) [][]byte {
	lines := make([][]byte, 0, len(data)/16)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(splitLines(reader.lineEnding))
	for scanner.Scan() {
		line := scanner.Bytes()
		lines = append(lines, line)
	}
	return lines
}

// SplitTokens separates data into tokens with PDF syntax
func (reader *PDFTokenReader) SplitTokens(data []byte) []Token {
	tokens := make([]Token, 0, len(data)/4)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(splitTokens)
	for scanner.Scan() {
		token := Token(scanner.Bytes())
		tokens = append(tokens, token)
	}
	return tokens
}

// Reset a reader to the start of the file
func (reader *PDFTokenReader) Reset() (int64, error) {
	newOffset, err := reader.file.Seek(0, 0)
	reader.scanner = bufio.NewScanner(reader.file)
	return newOffset, err
}

// Seek to a given point in the file
func (reader *PDFTokenReader) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := reader.file.Seek(offset, whence)
	reader.scanner = bufio.NewScanner(reader.file)
	return newOffset, err
}

// Peek looks ahead to get data but doesn't move the file pointer at all
func (reader *PDFTokenReader) Peek(n int) []byte {
	var peek []byte
	reader.splitPeek(n, &peek)
	reader.scanner.Scan()
	return peek
}

// PeekTokens looks ahead to get tokens but doesn't move the file pointer at all
func (reader *PDFTokenReader) PeekTokens(n int) []Token {
	var peek []Token
	reader.splitPeekTokens(n, &peek)
	reader.scanner.Scan()
	return peek
}

// ReadByte gets a single byte
func (reader *PDFTokenReader) ReadByte() (byte, bool) {
	reader.splitOnBytes()
	if reader.scanner.Scan() {
		return reader.scanner.Bytes()[0], true
	}
	return 0, false
}

// ReadToken gets the next PDF token
func (reader *PDFTokenReader) ReadToken() Token {
	reader.splitOnTokens()
	if reader.scanner.Scan() {
		return Token(reader.scanner.Bytes())
	}
	return Token([]byte{})
}

// ReadLine gets a line of tokens
func (reader *PDFTokenReader) ReadLine() []byte {
	reader.splitOnLines()

	if reader.scanner.Scan() {
		return reader.scanner.Bytes()
	}
	return []byte{}
}

func (reader *PDFTokenReader) SkipBytes(n int) bool {
	reader.splitNext(n)
	return reader.scanner.Scan()
}

// SkipToToken seeks ahead to the first instance of the given token
func (reader *PDFTokenReader) SkipToToken(token Token) bool {
	reader.splitUntil(token)
	if !reader.scanner.Scan() {
		return false
	}
	return true
}

// ReadTokens reads a fixed number of tokens
func (reader *PDFTokenReader) ReadTokens(n int) []Token {
	reader.splitOnTokens()
	result := make([]Token, 0, n)
	for len(result) < n && reader.scanner.Scan() {
		result = append(result, Token(reader.scanner.Bytes()))
	}
	return result
}

// ReadTokensToToken reads all tokens from the current position until the next instance of the given token
// If the token cannot be found, returns nil
func (reader *PDFTokenReader) ReadTokensToToken(token Token) ([]Token, bool) {
	data, ok := reader.ReadBytesToToken(token)
	return reader.SplitTokens(data), ok
}

// ReadLinesToToken reads all lines from the current position until the next instance of the given token
// If the token cannot be found, returns nil
func (reader *PDFTokenReader) ReadLinesToToken(token Token) ([][]byte, bool) {
	data, ok := reader.ReadBytesToToken(token)
	return reader.SplitLines(data), ok
}

// ReadBytesToToken reads all bytes from current position until the next instance of the given token
// If the token cannot be found, returns nil
func (reader *PDFTokenReader) ReadBytesToToken(token Token) ([]byte, bool) {
	reader.splitUntil(token)

	buf := bytes.NewBuffer([]byte{})
	for reader.scanner.Scan() {
		buf.Write(reader.scanner.Bytes())
	}
	return buf.Bytes(), reader.scanner.Err() == nil
}

// ReadBytes reads up to a fixed number of bytes
func (reader *PDFTokenReader) ReadBytes(n int) ([]byte, bool) {
	reader.splitNext(n)

	buf := bytes.NewBuffer([]byte{})
	for reader.scanner.Scan() {
		buf.Write(reader.scanner.Bytes())
	}
	return buf.Bytes(), reader.scanner.Err() != nil
}

// findXrefTable is a special function to read the offset of the xref table from somewhere near the end of the PDF file
func (reader *PDFTokenReader) findXrefTable() (int64, error) {
	// read the last chunk of file
	stat, err := reader.file.Stat()
	if err != nil {
		return 0, err
	}
	var readFrom = stat.Size() - searchForStartxrefLength
	var readLen = int64(searchForStartxrefLength)
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

func (reader *PDFTokenReader) checkXrefTable(offset int64) (int64, error) {
	if _, err := reader.Seek(offset, 0); err != nil {
		return offset, err
	}
	if peek := reader.Peek(4); string(peek) != "xref" {
		fmt.Println("NO! xref !=", string(peek), "- must try harder")
		// todo find the nearest "xref" token
	}

	return offset, nil
}
