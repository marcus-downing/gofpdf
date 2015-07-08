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
)

const (
	defaultPdfVersion = 1.3
)

func openParser(filename string) (*parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	p := new(parser)
	p.f = f
	p.context = context{}
	return p, nil
}

type parser struct {
	f       *os.File
	context context
}

func (p *parser) getPdfVersion() float64 {
	p.f.Seek(0, 0)
	b := make([]byte, 16)
	if _, err := p.f.Read(b); err != nil {
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

func (p *parser) setPageNumber(pageNumber int) {

}

type context struct {
}
