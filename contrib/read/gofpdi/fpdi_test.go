package gofpdi_test

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/read/gofpdi"
	"github.com/jung-kurt/gofpdf/internal/example"
)

func ExampleRead() {
	filename := example.Filename("Fpdf_AddPage")

	reader, err := gofpdi.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	// template := reader.Page(1)

	// page
	template := reader.Page(1)
	// template := rsc.PageToTemplate(&page)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.UseTemplate(template)
	fileStr := example.Filename("contrib_read_Read")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)

	// Output:
	// Successfully generated ../../../pdf/contrib_read_Read.pdf
}
