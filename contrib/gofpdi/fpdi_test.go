package gofpdi_test

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi/readpdf"
	. "github.com/jung-kurt/gofpdf/contrib/gofpdi/types"
	"github.com/jung-kurt/gofpdf/internal/example"
	// "os"
	"bufio"
	// "bytes"
	"strings"
	"time"
	"testing"
)

func TestSwitchingScanner(t *testing.T) {
	inputString := "abc d e f g"
	scanner := readpdf.NewSwitchingScanner(strings.NewReader(inputString))
	tokens := make([]string, 0, 6)

	scanner.Split(bufio.ScanBytes)
	for i := 0; i < 3 && scanner.Scan(); i++ {
		tokens = append(tokens, scanner.Text())
	}
	scanner.Split(bufio.ScanWords)
	for i := 0; i < 3 && scanner.Scan(); i++ {
		tokens = append(tokens, scanner.Text())
	}

	result := fmt.Sprintf("%v", tokens)
	if result != "[a b c d e f]" {
		t.Logf("Not expected result: %s", result)
		t.Fail()
	}
}

func TestTokenReader(t *testing.T) {
	inputString := "%PDF-1.3\nabc d e f g << /Hi jk >>"
	reader, err := readpdf.NewTokenReader(strings.NewReader(inputString), int64(len(inputString)))
	if err != nil {
		t.Logf("Error: %v", err)
		t.Fail()
	}

	if reader.PdfVersion != "1.3" {
		t.Logf("Incorrect PDF version: %s", reader.PdfVersion)
		t.Fail()
	}

	reader.SkipToToken(Token("d"))
	if peek := string(reader.Peek(5)); peek != "d e f" {
		t.Logf("Unable to skip to token")
		t.Logf("Peek: %s", peek)
		t.Fail()
	}

	reader.SkipBytes(1)
	if peek := string(reader.Peek(5)); peek != " e f " {
		t.Logf("Unable to skip bytes")
		t.Logf("Peek: %s", peek)
		t.Fail()
	}

	reader.SkipToToken(Token("g"))
	if peek := string(reader.Peek(5)); peek != "g << " {
		t.Logf("Unable to skip to token")
		t.Logf("Peek: %s", peek)
		t.Fail()
	}
	if peek := reader.PeekTokens(5); fmt.Sprintf("%v", peek) != fmt.Sprintf("%v", []Token{Token("g"), Token("<<"), Token("/Hi"), Token("jk"), Token(">>")}) {
		t.Logf("Unable to tokenise")
		t.Logf("Peek tokens: %v", peek)
		t.Logf("Peek: %s", string(reader.Peek(100)))
		t.Fail()
	}

	// reader.SkipToken()
	// if peek := string(reader.Peek(5)); peek != "<< /h" {
	// 	t.Logf("Unable to skip token")
	// 	t.Logf("Peek: %s", peek)
	// 	t.Fail()
	// }
}

func TestTokenParser(t *testing.T) {

}

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
	// Successfully generated ../../pdf/Fpdi_ExampleRead.pdf
}
