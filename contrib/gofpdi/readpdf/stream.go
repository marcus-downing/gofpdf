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
	. "github.com/jung-kurt/gofpdf/contrib/gofpdi/types"
)

// StreamFilter is any binary decode function
type StreamFilter func(Stream) Stream

func GetStreamFilter(name string) StreamFilter {
	switch name {
	case "/Fl", "/FlateDecode":
		return decodeFlate
	case "/LZWDecode":
		return decodeLZW
	case "/ASCII85Decode":
		return decodeAscii85
	case "/ASCIIHexDecode":
		return decodeAsciiHex
	case "":
		// do nothing
		return decodeNoop
	default:
		// error: cannot decode
		return decodeNoop
	}
}

// decodeNoop does nothing
func decodeNoop(stream Stream) Stream {
	return stream
}

// decodeFlate decodes a gzip/flate compressed stream
func decodeFlate(stream Stream) Stream {
	// ...
	return stream
}

// decodeAscii85 decodes an ASCII-85 encoded stream
func decodeAscii85(stream Stream) Stream {
	// ...
	return stream
}

// decodeAsciiHex decode an ASCII Hex encoded stream
func decodeAsciiHex(stream Stream) Stream {
	// ...
	return stream
}

// decodeLZW decodes an LZW-encoded stream
func decodeLZW(stream Stream) Stream {
	// ...
	return stream
}
