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
	"fmt"
	"github.com/jung-kurt/gofpdf"
	. "github.com/jung-kurt/gofpdf/contrib/gofpdi/types"
	"math"
)

// PDFPage is a page extracted from an existing PDF document
type PDFPageParser struct {
	object *ObjectDeclaration
	parser *PDFParser
	number int
}

func (page *PDFPageParser) Get(key string) (Value, bool) {
	v, b := page.object.Get(key)
	return v, b
}

func (page *PDFPageParser) Bytes() []byte {
	if contentsRef, ok := page.object.Get(ContentsRef); ok && contentsRef != nil {
		if contents := page.parser.resolveObject(contentsRef); contents != nil {
			if stream := contents.GetStream(); stream != nil {
				return stream.Bytes
			} else {
				fmt.Println("Unable to load stream from page contents object:", contents)
			}
		} else {
			fmt.Println("Unable to resolve page contents")
		}
	} else {
		fmt.Println("Page has no contents")
	}
	return []byte{}
}

func (page *PDFPageParser) Fonts() map[string]gofpdf.FontDefType {
	// fmt.Println("Loading fonts")
	if resourcesRef, ok := page.object.Get(ResourcesRef); ok && resourcesRef != nil {
		if resources := page.parser.resolveObject(resourcesRef); resources != nil {
			if fontListObj, ok := resources.Get(FontRef); ok && fontListObj != nil {
				if fontListDic, ok := fontListObj.(Dictionary); ok {
					// fmt.Println("Loading fonts:", fontListDic)
					fonts := make(map[string]gofpdf.FontDefType, len(fontListDic))
					for fontId, fontRef := range fontListDic {
						fontObj := page.parser.resolveObject(fontRef)
						// fmt.Println("Loading font", fontId, "=", fontObj)

						if dictType, ok := fontObj.Get(TypeRef); ok && dictType.ToString().String() == FontRef {
							fontDef := gofpdf.FontDefType{}
							if baseFont, ok := fontObj.Get(BaseFontRef); ok {
								fontDef.Name = baseFont.ToString().String()
							}
							// if encoding, ok := fontObj.Get(EncodingRef); ok {
							// 	fontDef.Enc = encoding.ToString().String()
							// }
							if subtype, ok := fontObj.Get(SubtypeRef); ok {
								fontDef.Tp = subtype.ToString().String()
							}
							fonts[fontId] = fontDef
						}
					}
					return fonts
				}
			}
		}
	}
	return nil
}

func (page *PDFPageParser) Images() map[string]*gofpdf.ImageInfoType {
	return nil
}

func (page *PDFPageParser) Templates() []gofpdf.Template {
	return nil
}

// GetPageBoxes gets the all the bounding boxes for a given page
//
// pageNumber is 1-indexed
// k is a scaling factor from user space units to points
func (page *PDFPageParser) GetPageBoxes(k float64) PageBoxes {
	boxes := make(map[string]*PageBox, 5)

	// fmt.Println("Reading page boxes from page:", page)
	if box := page.getPageBox(MediaBox, k); box != nil {
		boxes[MediaBox] = box
	}
	if box := page.getPageBox(CropBox, k); box != nil {
		boxes[CropBox] = box
	}
	if box := page.getPageBox(BleedBox, k); box != nil {
		boxes[BleedBox] = box
	}
	if box := page.getPageBox(TrimBox, k); box != nil {
		boxes[TrimBox] = box
	}
	if box := page.getPageBox(ArtBox, k); box != nil {
		boxes[ArtBox] = box
	}
	return PageBoxes{boxes, DefaultBox}
}

// getPageBox reads a bounding box from a page.
//
// page is a /Page dictionary.
//
// k is a scaling factor from user space units to points.
func (page *PDFPageParser) getPageBox(boxIndex string, k float64) *PageBox {
	/*
	   $page = $this->resolveObject($page);
	   $box = null;
	   if (isset($page[1][1][$boxIndex])) {
	       $box = $page[1][1][$boxIndex];
	   }

	   if (!is_null($box) && $box[0] == pdf_parser::TYPE_OBJREF) {
	       $tmp_box = $this->resolveObject($box);
	       $box = $tmp_box[1];
	   }

	   if (!is_null($box) && $box[0] == pdf_parser::TYPE_ARRAY) {
	       $b = $box[1];
	       return array(
	           'x' => $b[0][1] / $k,
	           'y' => $b[1][1] / $k,
	           'w' => abs($b[0][1] - $b[2][1]) / $k,
	           'h' => abs($b[1][1] - $b[3][1]) / $k,
	           'llx' => min($b[0][1], $b[2][1]) / $k,
	           'lly' => min($b[1][1], $b[3][1]) / $k,
	           'urx' => max($b[0][1], $b[2][1]) / $k,
	           'ury' => max($b[1][1], $b[3][1]) / $k,
	       );
	   } else if (!isset($page[1][1]['/Parent'])) {
	       return false;
	   } else {
	       return $this->_getPageBox($this->resolveObject($page[1][1]['/Parent']), $boxIndex, $k);
	   }
	*/

	// page = parser.resolveObject(page)

	if box, ok := page.Get(boxIndex); ok {
		if boxRef, ok := box.(ObjectRef); ok {
			box = page.parser.resolveObject(boxRef)
		}

		if arr, ok := box.(Array); ok {
			// fmt.Println("getPageBox: Parsing", arr)
			x := arr[0].ToReal().ToFloat64()
			y := arr[1].ToReal().ToFloat64()
			x2 := arr[2].ToReal().ToFloat64()
			y2 := arr[3].ToReal().ToFloat64()
			// fmt.Printf("getPageBox: x = %f, y = %f, x2 = %f, y2 = %f\n", x, y, x2, y2)
			w := math.Abs(x2 - x)
			h := math.Abs(y2 - y)
			llx := math.Min(x, x2)
			lly := math.Min(y, y2)
			urx := math.Max(x, x2)
			ury := math.Max(y, y2)

			pb := PageBox{
				gofpdf.PointType{x / k, y / k},
				gofpdf.SizeType{w / k, h / k},
				gofpdf.PointType{llx / k, lly / k},
				gofpdf.PointType{urx / k, ury / k},
			}
			// fmt.Println("getPageBox:",boxIndex,"=",pb)
			return &pb
		}

	}
	if parent, ok := page.Get(ParentRef); ok {
		// the page doesn't define its own bounding box, so fall back on the parent
		// which could be another page or the Pages container
		if parent := page.parser.resolveObject(parent); parent != nil {
			// fmt.Println("getPageBox:",boxIndex,"(parent)")
			parentPage := &PDFPageParser{parent, page.parser, 0}
			return parentPage.getPageBox(boxIndex, k)
		}
	}
	// fmt.Println("getPageBox:",boxIndex,"= nil")
	return nil
}
