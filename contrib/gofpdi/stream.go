package gofpdi

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
