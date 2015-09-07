package rsc_test

import (
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/read/rsc"
	"github.com/jung-kurt/gofpdf/internal/example"
	"rsc.io/pdf"
	"fmt"
)


func ExampleRead() {
	filename := example.Filename("Fpdf_AddPage")
	reader, err := pdf.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	// page
	page := reader.Page(1)
	template := rsc.PageToTemplate(&page)


	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.UseTemplate(template)
	fileStr := example.Filename("contrib_read_Read")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated ../../../pdf/contrib_read_Read.pdf
}