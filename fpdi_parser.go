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
	// "strings"
	"regexp"
	"strconv"
	"bufio"
)

const (
	defaultPdfVersion = 1.3
)

func openFileParser(filename string) (*PDFParser, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := NewFileReader(file)
	p := &PDFParser{reader}
	
	return p, nil
}

type PDFParser struct {
	reader          *PDFReader
	pageNumber      int
	lastUsedPageBox string
}

func (parser *PDFParser) getPdfVersion() float64 {
	b, err := parser.reader.Peek(16)
	if err != nil {
		return defaultPdfVersion
	}

	re := regexp.MustCompile("\\d\\.\\d")
	match := re.Find(b)
	if match == nil {
		return defaultPdfVersion
	}
	version, err := strconv.ParseFloat(string(match[:]), 64)
	if err != nil {
		return defaultPdfVersion
	}
	return version
}

func (parser *PDFParser) setPageNumber(pageNumber int) {
	parser.pageNumber = pageNumber
}



type PageBoxes struct {
	pageBoxes       map[string]*PageBox
	lastUsedPageBox string
}

func (boxes PageBoxes) get(boxName string) *PageBox {
    /**
     * MediaBox
     * CropBox: Default -> MediaBox
     * BleedBox: Default -> CropBox
     * TrimBox: Default -> CropBox
     * ArtBox: Default -> CropBox
     */
    if pageBox, ok := boxes.pageBoxes[boxName]; ok {
    	boxes.lastUsedPageBox = boxName
    	return pageBox
    }
    switch pageBox {
    case BleedBox:
    case TrimBox:
    case ArtBox:
    	return boxes.get(CropBox)
    case CropBox:
    	return boxes.get(MediaBox)
    default:
    	return nil
    }
}

type PageBox struct {
	PointType
	SizeType
	Lower PointType // llx, lly
	Upper PointType // urx, ury
}

func (parser *PDFParser) getPageBoxes(pageNumber int, k float64) PageBoxes {
	page = parser.pages[pageNumber]
	boxes := make(map[string]*PageBox, 5)
	boxes[MediaBox] = parser.getPageBox(MediaBox, k)
	boxes[CropBox] = parser.getPageBox(CropBox, k)
	boxes[BleedBox] = parser.getPageBox(BleedBox, k)
	boxes[TrimBox] = parser.getPageBox(TrimBox, k)
	boxes[ArtBox] = parser.getPageBox(ArtBox, k)
	return PageBoxes{boxes,""}
}

func (parser *PDFParser) getPageBox(page dictionary, boxIndex string, k float64) {

}
