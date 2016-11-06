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
	// "bytes"
	// "errors"
	"fmt"
	// "io"
	"os"
	"regexp"
	// "strconv"
)

const (
	splitterNil        = iota
	splitterOnLines
	splitterPeek
	splitterNext
	splitterUntil
	splitterPeekTokens
	splitterOnTokens
	splitterOnBytes
)

// MutantScanner is a scanner that can switch its tokeniser mid-read
type MutantScanner struct {
	file       *os.File       // the file being read
	lineEnding []byte         // the type of line ending this file uses
	scanner    *bufio.Scanner // a syntax-specific tokenizer
	splitter   byte           // how was it last split
}

func NewMutantScanner (file *os.File) MutantScanner {
	scanner := MutantScanner{file, nil, nil, splitterNil}
	scanner.calibrateLineEndings()
	return scanner
}

// calibrateLineEndings uses some sample data to determine which line-ending is used by this file
func (reader *MutantScanner) calibrateLineEndings() {
	reader.lineEnding = []byte("\n")
	sample := make([]byte, 100)
	reader.file.Read(sample)
	rx := regexp.MustCompile("\r\n|\n|\r")
	match := rx.Find(sample)
	if match != nil {
		reader.lineEnding = match
	}
}

func (ms *MutantScanner) Reset() {
	ms.scanner = nil
	ms.splitter = splitterNil
}


func (ms *MutantScanner) Bytes() []byte {
	return ms.scanner.Bytes()
}

func (ms *MutantScanner) Err() error {
	return ms.scanner.Err()
}

func (ms *MutantScanner) Scan() bool {
	return ms.scanner.Scan()
}

func (ms *MutantScanner) Text() string {
	return ms.scanner.Text()
}


// Close the scanner
func (ms *MutantScanner) Close() {
	ms.scanner = nil
	ms.file.Close()
}

func (ms *MutantScanner) Split(split bufio.SplitFunc, splitter byte) {
	if ms.splitter != splitterNil {
		switch (splitter) {
		case splitterOnLines:
		case splitterOnBytes:
		case splitterOnTokens:
			if ms.splitter == splitter {
				return
			}
		}
	}
	fmt.Println("Switching to splitter:", splitter)
	ms.scanner = bufio.NewScanner(ms.file)
	ms.scanner.Split(split)
	ms.splitter = splitter
}

/*
// updateScanner switches to a new scanner, but only if the splitter is different
func (reader *MutantScanner) updateScanner(splitter byte) {
	if reader.splitter == splitterNil {
		fmt.Println("Initial splitter:", splitter)
		reader.splitter = splitter
		return
	}
	switch (splitter) {
	case splitterOnLines:
	case splitterOnBytes:
	case splitterOnTokens:
		if reader.splitter == splitter {
			return
		}
	}
	fmt.Println("Switching to splitter:", splitter)
	reader.scanner = bufio.NewScanner(reader.file)
	reader.splitter = splitter
}
*/
