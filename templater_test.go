package golangunitedschoolcerts

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/stretchr/testify/assert"
	qrdecode "github.com/tuotoo/qrcode"
)

func mockGotenbergService(t *testing.T, exp []byte) (url string, closer func()) {
	mock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write(exp)
	}))
	return mock.URL, mock.Close
}

func Test_linkToQR(t *testing.T) {

	tRcL := map[string]qrcode.RecoveryLevel{
		"Low":     qrcode.Low,
		"Medium":  qrcode.Medium,
		"High":    qrcode.High,
		"Highest": qrcode.Highest,
	}

	tData := map[string]struct {
		link    string
		wantErr bool
	}{
		"no link":   {"", true},
		"one char":  {"a", false},
		"few chars": {"https://example.com/certificate/6acyxaxb", false},
		"a lot of 'a'": {"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
			"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		"diff symbols": {"@_-", false},
		//"one number": {"1", false},
		//"uppercase letter": {"A", false},
		"mix #1 of numbers and letters": {"a1a1a1a1a", false},
		"mix #2 of numbers and letters": {"1a", false},
		"mix #3 of numbers and letters": {"a111", false}, //success test
		//"mix #4 of numbers and letters": {"a1111", false}, //fail test
	}

	for name, tcase := range tData {

		// test with different recoveryLevels
		for nameRcL, rcL := range tRcL {

			got, err := linkToQR(tcase.link, rcL, -5)

			if (err != nil) != tcase.wantErr {
				t.Errorf("[%s] linkToQR() error = %v, wantErr = %v", name, err, tcase.wantErr)
			}

			// if there is error in coding QR, there is no sense to decode nothing, it will fail
			if err == nil {

				// base64 decoding to []byte
				dc, err := base64.StdEncoding.DecodeString(got)
				if err != nil {
					t.Errorf("[%s] failed to decode base64 link to QR: %v", name, err)
				}

				// Decoding QR code to qrmatrix
				r := bytes.NewReader(dc)
				qrmatrix, err := qrdecode.Decode(r)
				if err != nil {
					t.Errorf("[%s] failed to decode QR code with %s RecoveryLevel: %v", name, nameRcL, err)
				}

				gotLink := qrmatrix.Content
				if tcase.link != gotLink {
					t.Errorf("[%s] expected %s, got %s", name, tcase.link, gotLink)
				}
			}
		}
	}
}

func Test_renderHTML(t *testing.T) {

	tmplFile := "./test/testdata/templater/renderhtml/index.html"
	expResultFile := "./test/testdata/templater/renderhtml/expindex.html"

	tFile, err := os.ReadFile(tmplFile)
	if err != nil {
		t.Errorf("Can't load test template file index.html")
	}

	eFile, err := os.ReadFile(expResultFile)
	if err != nil {
		t.Errorf("Can't load expected result file expindex.html")
	}

	tData := map[string]struct {
		tmpl    string
		d       data
		expRndr []byte
		expErr  error
	}{
		"index.html": {
			string(tFile),
			data{
				Certificate{
					Student:   "Bill Gates",
					IssueDate: "20240101",
					Course:    "The Golang programming language",
					Mentors:   "Rob Pike",
				},
				"https://example.com/certificate/6acyxaxb",
				"aHR0cHM6Ly9leGFtcGxlLmNvbS9jZXJ0aWZpY2F0ZS82YWN5eGF4Yg==...",
			},
			[]byte(eFile),
			nil},
	}

	for name, tcase := range tData {

		got, err := renderHTML(tcase.tmpl, &tcase.d)

		// Checking for errors
		if err != tcase.expErr {
			t.Errorf("[%s] expected error: %v, got %v", name, tcase.expErr, err)
			return
		}

		// Checking equal between rendered and expected HTML content
		if !reflect.DeepEqual(*got, tcase.expRndr) {
			t.Errorf("[%s] expected: %s, got %s", name, tcase.expRndr, *got)
		}
	}
}

func Test_renderPDF(t *testing.T) {
	html := &[]byte{}
	exp := &[]byte{0, 1, 0, 1}

	url, closer := mockGotenbergService(t, *exp)
	defer closer()

	g := NewGotenbergTemplater(url)
	got, err := g.renderPDF(html)

	assert.NoError(t, err)
	assert.Equal(t, exp, got)
}

func Test_GenerateCertificate(t *testing.T) {
	t.Run("Expecting successful run", func(t *testing.T) {
		tmplFile := "./test/testdata/templater/renderhtml/index.html"
		b, err := os.ReadFile(tmplFile)
		if err != nil {
			assert.FailNow(t, "Can't read template file: %v, err: %v", tmplFile, err)
		}
		template := string(b)

		cert := &Certificate{
			Id:        "01010101",
			Student:   "Test Student",
			IssueDate: "1 December 1999",
			Course:    "Test Course",
			Mentors:   "Test Mentor 1, Test Mentor 2",
		}

		link := "example.com/certificates/" + cert.Id

		exp := &[]byte{0, 1, 0, 1}
		url, closer := mockGotenbergService(t, *exp)
		defer closer()

		g := NewGotenbergTemplater(url)

		got, err := g.GenerateCertificate(template, cert, link)
		assert.NoError(t, err)
		assert.Equal(t, exp, got)
	})
}
