package rsc

import (
	"github.com/jung-kurt/gofpdf"
	rsc "rsc.io/pdf"
)

func PageToTemplate(page *rsc.Page) gofpdf.Template {
	id := gofpdf.GenerateTemplateID()
	return RscTemplate{id, page}
}

type RscTemplate struct {
	Id int64
	Page *rsc.Page
}

func (tpl RscTemplate) ID() int64 {
	return 0
}

func (tpl RscTemplate) Size() (gofpdf.PointType, gofpdf.SizeType) {
	
	return gofpdf.PointType{}, gofpdf.SizeType{}
}

func (tpl RscTemplate) Bytes() []byte {
	return nil
}

func (tpl RscTemplate) Images() map[string]*gofpdf.ImageInfoType {
	return nil
}

func (tpl RscTemplate) Templates() []gofpdf.Template {
	return nil
}