package gofpdi_test

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi"
	"github.com/jung-kurt/gofpdf/internal/example"
	// "os"
	// "bufio"
	// "bytes"
	"time"
)

// ExampleRead tests the ability to read an existing PDF file
// and use a page of it as a template in another file
func ExampleRead() {
	filename := example.Filename("Fpdf_AddPage")

	// force the test to fail after 10 seconds
	go func() {
		time.Sleep(10000 * time.Millisecond)
		panic("Time out")
	}()

	reader, err := gofpdi.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	// page
	template := reader.Page(1)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.UseTemplate(template)
	fileStr := example.Filename("Fpdi_ExampleRead")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)

	// Output:
	// Successfully generated ../../../pdf/Fpdi_ExampleRead.pdf
}
