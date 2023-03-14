//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	crt "gitlab.com/DzmitryYafremenka/golang-united-school-certs"
)

func Test_Gotenber(t *testing.T) {
	templatePath := "./test/testdata/templater/gotenberg/index.html"

	// tests use package directory as working directory, changing to project root
	err := os.Chdir("../..")
	if err != nil {
		assert.FailNow(t, "failed to change working directory: %v", err)
	}

	template, err := os.ReadFile(templatePath)
	if err != nil {
		assert.FailNow(t, "failed to read template file: %v", err)
	}

	cert := &crt.Certificate{
		Id:        "12345678",
		Student:   "Test Student",
		IssueDate: "1 December 1999",
		Course:    "Test Course",
		Mentors:   "Test Mentor 1, Test Mentor 2",
	}

	link := "http://example.com/certificates/" + cert.Id

	g := crt.NewGotenbergTemplater("http://localhost:3000")
	got, err := g.GenerateCertificate(string(template), cert, link)

	assert.NoError(t, err)
	assert.NotEmpty(t, got)

	outPath := "./tmp/integration/gotenberg/"
	outFile := "test.pdf"

	_, err = os.Stat(outPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(outPath, 0777)
			if err != nil {
				assert.FailNow(t, "unexpected error while creating temporary directory: %v", err)
			}
		} else {
			assert.FailNow(t, "unexpected error when checking if temporary directory exists: %v", err)
		}
	}
	err = os.WriteFile(outPath+outFile, *got, 0777)
	if err != nil {
		assert.FailNow(t, "unexpected error while writing test output file: %v, err")
	}
}
