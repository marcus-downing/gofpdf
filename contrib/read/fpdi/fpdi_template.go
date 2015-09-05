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

// fpdiTemplate is a page template, read from an existing page, that can be used in other documents.
type fpdiTemplate struct {
	id       int64
	pageSize SizeType
}

// ID returns the global template identifier
func (t *fpdiTemplate) ID() int64 {
	return t.id
}

// Size gives the bounding dimensions of this template
func (t *fpdiTemplate) Size() (PointType, SizeType) {
	return PointType{0, 0}, t.pageSize
}

// Bytes returns the actual template data, not including resources
func (t *fpdiTemplate) Bytes() []byte {
	return nil
}

// Images returns a list of the images used by this template
func (t *fpdiTemplate) Images() map[string]*ImageInfoType {
	return nil
}

// Templates returns a list of templates used within this template
func (t *fpdiTemplate) Templates() []Template {
	return nil
}
