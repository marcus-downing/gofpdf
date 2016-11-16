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
	// "os"
	"bufio"
	"io"
	// "fmt"
	"errors"
)

// SwitchingScanner extends bufio.Scanner to allow you to switch the split algorithm on the fly.
// It also implements Seek, if the underlying input reader supports it.
type SwitchingScanner struct {
	*bufio.Scanner
	in        io.Reader
	splitFunc bufio.SplitFunc
}

// NewSwitchingScanner creates a switching scanner to read from the input.
func NewSwitchingScanner(in io.Reader) *SwitchingScanner {
	ss := SwitchingScanner{bufio.NewScanner(in), in, nil}
	ss.Scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := ss.splitFunc(data, atEOF)
		return advance, token, err
	})
	return &ss
}

// Split sets the split function for the Scanner.
// Unlike the standard Scanner implementation, this does not panic if called after scanning has started.
func (ss *SwitchingScanner) Split(fn bufio.SplitFunc) {
	ss.splitFunc = fn
}

func (ss *SwitchingScanner) Scan() bool {
	return ss.Scanner.Scan()
}

/*func (ss *SwitchingScanner) Buffer(buf []byte, max int) {
	ss.scanner.Buffer(buf, max)
}

func (ss *SwitchingScanner) Bytes() []byte {
	return ss.scanner.Bytes()
}

func (ss *SwitchingScanner) Err() error {
	return ss.scanner.
}

func (ss *SwitchingScanner) Scan() bool
func (ss *SwitchingScanner) Text() string
*/

// Seek sets the offset for the next read, if the underlying input reader supports it.
// See https://golang.org/pkg/io/#Seeker
func (ss *SwitchingScanner) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := ss.in.(io.Seeker); ok {
		ss.Scanner = nil

		// fmt.Println("SwitchingScanner: Seeking to", offset, "from", whence)
		newOffset, err := seeker.Seek(offset, whence)
		// fmt.Println("SwitchingScanner: New offset =", newOffset)

		// the scanner will be holding onto some buffered data, so it has to be replaced with a fresh one
		ss.Scanner = bufio.NewScanner(ss.in)
		ss.Scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
			advance, token, err := ss.splitFunc(data, atEOF)
			return advance, token, err
		})

		return newOffset, err
	}
	return 0, errors.New("Unseekable")
}

// Close releases any resources held open
func (ss *SwitchingScanner) Close() error {
	ss.Scanner = nil
	if closer, ok := ss.in.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
