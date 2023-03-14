package golangunitedschoolcerts

import (
	"bytes"
	"encoding/base64"
	"fmt"
	tmpl "html/template"
	"io"
	"mime/multipart"
	"net/http"

	qrcode "github.com/skip2/go-qrcode"
)

type Templater interface {
	GenerateCertificate(template string, certificate *Certificate, link string) (*[]byte, error)
}

type GotenbergTemplater struct {
	url string
}

// Data structure for html template
type data struct {
	Cert Certificate
	Link string
	Qr   string
}

func NewGotenbergTemplater(url string) *GotenbergTemplater {
	return &GotenbergTemplater{url}
}

// Generate QR code with needed recovery level, size in pxs and base64 encoding
func linkToQR(link string, rcL qrcode.RecoveryLevel, size int) (string, error) {

	// Encode link to QR code
	png, err := qrcode.Encode(link, rcL, size)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR code: %w", err)
	}

	return base64.StdEncoding.EncodeToString(png), nil
}

// Execute html template with given data
func renderHTML(template string, d *data) (*[]byte, error) {

	// Creating HTML template and checking for correct parsing
	t, err := tmpl.New("HTML").Parse(template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Paste structure fields to HTML template and write it to buffer
	b := bytes.Buffer{}
	err = t.Execute(&b, *d)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	sl := b.Bytes()
	return &sl, nil
}

func (g *GotenbergTemplater) renderPDF(html *[]byte) (*[]byte, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to crete form file: %w", err)
	}

	_, err = part.Write(*html)
	if err != nil {
		return nil, fmt.Errorf("failed to write template to multipart: %w", err)
	}

	// respect @page properties stated in css
	err = writer.WriteField("preferCssPageSize", "true")
	if err != nil {
		return nil, fmt.Errorf("failed to write preferCssPageSize to multipart: %w", err)
	}

	writer.Close()

	resp, err := http.Post(g.url+"/forms/chromium/convert/html", writer.FormDataContentType(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to perform POST request to gotenber: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gotenberg return error: %v", resp.Status)
	}

	pdf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gotenber response: %w", err)
	}
	return &pdf, nil
}

func (g *GotenbergTemplater) GenerateCertificate(template string, cert *Certificate, link string) (*[]byte, error) {
	// -4 makes each QR "pixel" to be 4px in size
	qr, err := linkToQR(link, qrcode.High, -4)
	if err != nil {
		return nil, err
	}
	d := &data{
		Cert: *cert,
		Link: link,
		Qr:   qr,
	}
	html, err := renderHTML(template, d)
	if err != nil {
		return nil, err
	}
	pdf, err := g.renderPDF(html)
	if err != nil {
		return nil, err
	}
	return pdf, nil
}
