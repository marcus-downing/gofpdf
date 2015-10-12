package gofpdi

// StreamFilter is any binary decode function
type StreamFilter func(Stream) Stream

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
