//go:build integration

package integration

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	crt "gitlab.com/DzmitryYafremenka/golang-united-school-certs"
)

func Test_DBCertificates(t *testing.T) {
	type Tmpl struct {
		name    string
		content string
	}

	var tmpl = []Tmpl{
		{"Name 1", "Content 1"},
		{"Name 2", "Content 2"},
	}

	var expCerts = []*crt.Certificate{
		{Id: "", TemplatePk: 1, Timestamp: time.Time{}, Student: "Student 1",
			IssueDate: "2024-01-01 00:00:00", Course: "Course 1", Mentors: "Mentors 1"},
		{Id: "", TemplatePk: 1, Timestamp: time.Time{}, Student: "Student 2",
			IssueDate: "2025-01-01 00:00:00", Course: "Course 2", Mentors: "Mentors 2"},
		{Id: "", TemplatePk: 1, Timestamp: time.Time{}, Student: "Student 3",
			IssueDate: "2026-01-01 00:00:00", Course: "Course 3", Mentors: "Mentors 3"},
		{Id: "", TemplatePk: 2, Timestamp: time.Time{}, Student: "Student 4",
			IssueDate: "2027-01-01 00:00:00", Course: "Course 4", Mentors: "Mentors 4"},
		{Id: "", TemplatePk: 2, Timestamp: time.Time{}, Student: "Student 5",
			IssueDate: "2028-01-01 00:00:00", Course: "Course 5", Mentors: "Mentors 5"},
		{Id: "", TemplatePk: 2, Timestamp: time.Time{}, Student: "Student 6",
			IssueDate: "2029-01-01 00:00:00", Course: "Course 6", Mentors: "Mentors 6"},
	}

	connString := "postgres://user:password@localhost:5432/registry"
	r, err := crt.NewDirectRegistry(connString)
	assert.NoError(t, err)

	// Add entries to template and template_content tables
	for _, i := range tmpl {
		err = r.AddTemplate(i.name, i.content)
		assert.NoError(t, err)
	}

	idsTmpl := make([][]string, len(tmpl))
	for i := 0; i < len(tmpl); i++ {
		idsTmpl[i] = make([]string, 0)
	}

	// Check adding certificates to DB
	for _, exp := range expCerts {
		got, err := r.AddCertificate(tmpl[exp.TemplatePk-1].name, exp.Student, exp.IssueDate, exp.Course, exp.Mentors)
		// Remember Id and timestamp to test data
		exp.Id = got.Id
		exp.Timestamp = got.Timestamp
		// Remember Ids for TemplatePk to test CertificatesByTemplatePK func
		idsTmpl[exp.TemplatePk-1] = append(idsTmpl[exp.TemplatePk-1], exp.Id)
		assert.NoError(t, err)
		assert.Regexp(t, regexp.MustCompile("[0-9a-f]{8}"), got.Id)
		assert.WithinDuration(t, time.Now().UTC(), got.Timestamp, time.Second)
	}

	// Check that registry returns correct certificates for given id
	for _, exp := range expCerts {
		got, err := r.GetCertificate(exp.Id)
		assert.NoError(t, err)
		assert.Equal(t, exp, got)
	}

	// Check that registry returns correct certificate ids for given template pk
	for pk := 0; pk < len(tmpl); pk++ {
		ids, err := r.CertificatesByTemplatePK(pk + 1)
		assert.NoError(t, err)
		assert.Equal(t, idsTmpl[pk], ids)
	}

	// Check updating data in certificate and timestamp field
	time.Sleep(time.Second * 1)
	for _, exp := range expCerts {
		m := make(map[string]string)
		m["course"] = "New " + exp.Course
		m["Mentors"] = "New " + exp.Mentors
		err := r.UpdateCertificate(exp.Id, m)
		assert.NoError(t, err)
		got, err := r.GetCertificate(exp.Id)
		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now().UTC(), got.Timestamp, time.Second)
		assert.Equal(t, exp.Id, got.Id)
		assert.Equal(t, exp.TemplatePk, got.TemplatePk)
		assert.Equal(t, exp.Student, got.Student)
		assert.Equal(t, exp.IssueDate, got.IssueDate)
		assert.Equal(t, m["course"], got.Course)
		assert.Equal(t, m["Mentors"], got.Mentors)
		assert.NotEqual(t, exp.Course, got.Course)
		assert.NotEqual(t, exp.Mentors, got.Mentors)
	}

	for i, e := range tmpl {
		time.Sleep(time.Second * 1)
		m := make(map[string]string)
		m["content"] = "New " + e.content
		err := r.UpdateTemplate(i+1, m)
		assert.NoError(t, err)
		for _, id := range idsTmpl[i] {
			got, err := r.GetCertificate(id)
			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now().UTC(), got.Timestamp, time.Second)
		}
	}
	// Check updating only needed certificate with proper template pk
	got1, err := r.GetCertificate(idsTmpl[0][0])
	assert.NoError(t, err)
	got2, err := r.GetCertificate(idsTmpl[1][0])
	assert.NoError(t, err)
	assert.NotEqual(t, got1.Timestamp.Round(time.Second), got2.Timestamp.Round(time.Second))

	// Check that after deletion we get proper errors for all operations with certificates
	for _, exp := range expCerts {
		err := r.DeleteCertificate(exp.Id)
		assert.NoError(t, err)
		got, err := r.GetCertificate(exp.Id)
		assert.Nil(t, got)
		assert.Error(t, err)
		m := make(map[string]string)
		m["course"] = "New " + exp.Course
		err = r.UpdateCertificate(exp.Id, m)
		assert.Error(t, err)
		err = r.DeleteCertificate(exp.Id)
		assert.Error(t, err)
	}

	for i := range tmpl {
		ids, err := r.CertificatesByTemplatePK(i + 1)
		//nil error when where is no certificates by templatePK
		assert.NoError(t, err)
		assert.Equal(t, []string{}, ids)
	}
}
