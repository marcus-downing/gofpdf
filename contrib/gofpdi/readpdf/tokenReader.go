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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	// "os"
	"regexp"
	"strconv"
	. "github.com/jung-kurt/gofpdf/contrib/gofpdi/types"
)

const (
	startxref                = "startxref"
	searchForStartxrefLength = 1024 // 5500 // distance from the end of the file to search for the startxref offset
)

// PDFTokenReader is a low-level reader for the tokens in a PDF file
// See pdf_parser.php and pdf_context.php
type PDFTokenReader struct {
	in         io.Reader          // the file or other source being read
	filesize   int64
	scanner    *SwitchingScanner  // the file scanner
	// scanner       MutantScanner
	// scanner    *bufio.Scanner // a syntax-specific tokenizer
	stack      []Token        // a stack of pre-read tokens
	lineEnding []byte         // the type of line ending this file uses
	PdfVersion string         // the version header
	// splitter   byte           // how was it last split
}

// NewTokenReader constructs a low level reader for a PDF file
func NewTokenReader(in io.Reader, filesize int64) (*PDFTokenReader, error) {
	// fmt.Println("Creating new token reader over", filesize, "bytes")
	reader := new(PDFTokenReader)
	reader.in = in
	reader.filesize = filesize
	reader.scanner = NewSwitchingScanner(in)
	// reader.scanner = NewMutantScanner(file)
	reader.calibrateLineEndings()
	// reader.scanner = bufio.NewScanner(reader.file)
	// reader.lineEnding = []byte("\n")

	// always read the PDF version first
	reader.PdfVersion = DefaultPdfVersion
	if err := reader.getPdfVersion(); err != nil {
		return nil, err
	}

	return reader, nil
}

// calibrateLineEndings uses some sample data to determine which line-ending is used by this file
func (reader *PDFTokenReader) calibrateLineEndings() {
	reader.lineEnding = []byte("\n")
	sample := reader.Peek(100)
	rx := regexp.MustCompile("\r\n|\n|\r")
	match := rx.Find(sample)
	if match != nil {
		reader.lineEnding = match
	}
}

// func (reader *PDFTokenReader) FileStat() (os.FileInfo, error) {
// 	fi, err := reader.file.Stat()
// 	return fi, err
// }

// Close releases references and closes the file handle of the parser
func (reader *PDFTokenReader) Close() {
	reader.scanner.Close()
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
		// fmt.Println("splitPeek: data:", string(data))
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
	// fmt.Println("splitNext: split on", n, "bytes")
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			// fmt.Println("splitNext: at EOF")
			return 0, nil, io.EOF
		}

		// make sure we have enough loaded
		ld := len(data)
		if ld < n && !atEOF {
			// request more data
			// fmt.Println("splitNext: need more, we only have", len(data), "bytes")
			return 0, nil, nil
		}

		// return as much as we can
		sb := n
		if sb > ld {
			// fmt.Println("splitNext: we can only get", ld, "bytes")
			sb = ld
		}
		// fmt.Println("splitNext: returning", sb, "bytes")
		return sb, data[0:sb], nil
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
			// fmt.Println("splitPeekTokens: eof")
			return 0, nil, io.EOF
		}

		// request more data
		// we don't actually know how much we'll need, but 4n is a safe bet
		// and it's cheaper to make this request before trying to extract
		// enough tokens and finding we've run out of buffer.
		if len(data) < (n * 4) && !atEOF {
			// fmt.Printf("splitPeekTokens: len(data) = %d, n = %d\n", len(data), n)
			// fmt.Println("splitPeekTokens: more")
			return 0, nil, nil
		}

		// split tokens until we have the right number or run out of data
		result := make([]Token, 0, n)
		for len(result) < n {
			padvance, ptoken, perr := splitTokens(data, atEOF)
			if padvance == 0 && ptoken == nil && perr == nil {
				// request more data
				// fmt.Println("splitPeekTokens: more")
				return 0, nil, nil
			}

			if ptoken != nil {
				result = append(result, ptoken)
			}

			data = data[padvance:]
			if atEOF && len(data) == 0 {
				// fmt.Println("splitPeekTokens: reached eof")
				*into = result
				return 0, nil, io.EOF
			}

			if perr == io.EOF {
				// stop reading
				break
			}
		}
		// fmt.Println("splitPeekTokens: reached end")
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
	var whitespace int
	for whitespace = 0; whitespace < len(data) && IsPdfWhitespace(data[whitespace]); whitespace++ {
	}

	// if we're at the end of the data, signal that with an EOF
	if whitespace >= len(data) {
		if atEOF {
			return whitespace, nil, io.EOF
		}
		return whitespace, nil, nil
	}
	// os.Stderr.WriteString(fmt.Sprintf("Data length: %d bytes, offset: %d, data: %v\n", len(data), whitespace, data[0:6]))
	

	// otherwise just advance over the whitespace without returning a token
	// if offset > 0 {
	// 	fmt.Printf("Space: advance = %d, token = %v, err = %v\n", offset, []byte{}, nil)
	// 	return offset, []byte{}, nil
	// }

	data2 := data[whitespace:]
	// fmt.Printf("Data length: %d bytes, offset: %d, data: %v\n", len(data), whitespace, data2[0:6])

	// Get the first character in the stream
	b := data2[0]
	switch b {
	case 0x28, // (
		0x29, // )
		0x5B, // [
		0x5D: // ]
		// os.Stderr.WriteString("Scanner: string or array\n")
		// This is either an array or literal string delimiter, return it
		// fmt.Printf("Array or string token: advance = 1, token = '%s', err = nil\n", string([]byte{b}))
		return whitespace+1, []byte{b}, nil

	case 0x3C, // <
		0x3E: // >
		// os.Stderr.WriteString("Scanner: hex string or dictionary\n")
		// This could be either a hex string of dictionary delimiter.
		// determine which it is and return the token
		b2 := data2[1]
		// fmt.Println("Checking dictionary or hex token:", string(b), string(b2))
		if b2 == b {
			// fmt.Printf("Dictionary token: advance = 2, token = '%s', err = nil\n", string([]byte{b,b2}))
			return whitespace+2, []byte{b,b2}, nil
		}
		// fmt.Printf("Hex token: advance = 1, token = '%s', err = nil\n", string([]byte{b}))
		return whitespace+1, []byte{b}, nil

	case 0x25: // %
		// This is a comment - jump over it!
		// os.Stderr.WriteString("Scanner: comment\n")
		a, token, err := bufio.ScanLines(data2, atEOF)
		// fmt.Println("Comment token: advance = %d, token = '%s', err = %v", a, token, err)
		return whitespace+a, token, err
	}

	// This is another type of token (probably a dictionary entry or a numeric value). Find the end and return it.
	// os.Stderr.WriteString("Scanner: other token\n")
	var nextToken int
	for nextToken = 0; nextToken < len(data2) && !IsPdfWhitespaceOrBreak(data2[nextToken]); nextToken++ {
		// os.Stderr.WriteString(fmt.Sprintf("Skipped char: %v\n", data[nextToken]))
	}
	token = data2[0:nextToken]
	// os.Stderr.WriteString(fmt.Sprintf("Token: advance = %d, token = %s, err = %v\n", nextToken, string(token), err))
	// fmt.Printf("Token: advance = %d, token = '%s', err = %v\n", nextToken, string(token), err)
	return whitespace+nextToken, token, nil
}

func IsPdfWhitespace(b byte) bool {
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

func IsPdfWhitespaceOrBreak(b byte) bool {
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
		reader.PdfVersion = string(match[:])
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
	newOffset, err := reader.scanner.Seek(0, 0)
	return newOffset, err
}

// Seek to a given point in the file
func (reader *PDFTokenReader) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := reader.scanner.Seek(offset, whence)
	return newOffset, err
}

// Peek looks ahead to get data but doesn't move the file pointer at all
func (reader *PDFTokenReader) Peek(n int) []byte {
	var peek []byte
	// fmt.Println("PDFTokenReader: Peeking", n, "bytes")
	reader.splitPeek(n, &peek)
	if ok := reader.scanner.Scan(); !ok {
		// fmt.Println("PDFTokenReader: Unable to peek")
		// err := reader.scanner.Err()
		// if err != nil {
		// 	fmt.Println("PDFTokenReader: Error", err)
		// }
	}
	// fmt.Println("PDFTokenReader: Peeked", len(peek), "bytes")
	// fmt.Println("PDFTokenReader:", string(peek))
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

func (reader *PDFTokenReader) SkipToken() {
	reader.splitOnTokens()
	reader.scanner.Scan()
}

func (reader *PDFTokenReader) SkipTokens(n int) {
	reader.splitOnTokens()
	for i := 0; i < n; i++ {
		reader.scanner.Scan()
	}
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

// FindXrefTable is a special function to read the offset of the xref table from somewhere near the end of the PDF file
func (reader *PDFTokenReader) FindXrefTable() (int64, error) {
	// read the last chunk of file
	// stat, err := reader.file.Stat()
	// if err != nil {
	// 	return 0, err
	// }
	// fmt.Println("FindXrefTable: file size =", reader.filesize)
	// var readFrom = stat.Size() - searchForStartxrefLength
	readFrom := reader.filesize - searchForStartxrefLength
	var readLen = searchForStartxrefLength
	if readFrom < 0 {
		readFrom = 0
		readLen = int(reader.filesize)
	}

	// b := make([]byte, readLen)
	if _, err := reader.Seek(readFrom, io.SeekStart); err != nil {
		return 0, err
	}
	b := reader.Peek(readLen)
	// if err == io.EOF {
	// 	fmt.Println("EOF while reading")
	// } else if err != nil {
	// 	return 0, err
	// }

	// find the *LAST* instance of the "startxref" token
	matches := regexp.MustCompile(`startxref[\s\n]*(\d+)`).FindAllSubmatchIndex(b, -1)
	if matches == nil {
		fmt.Println("FindXrefTable: Reading from", readFrom, "for", readLen, "bytes")
		fmt.Println("FindXrefTable: Peek data:", b)
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
