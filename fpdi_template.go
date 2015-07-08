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
