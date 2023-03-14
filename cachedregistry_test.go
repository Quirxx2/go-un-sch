package golangunitedschoolcerts

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Check that struct implements interface
var _ Registry = &CachedRegistry{}

func createTestCachedRegistry(t *testing.T) (cr *CachedRegistry, rMock *MockRegistry) {
	rMock = NewMockRegistry(t)
	cr, err := NewCachedRegistry(rMock)
	if err != nil {
		assert.FailNow(t, "unexpected error creating CachedRegistry: %w", err)
	}
	return cr, rMock
}

func Test_CachedRegistry_GetTemplatePK(t *testing.T) {
	name := "Test Name"
	expPK := 1
	t.Run("Cache hits after repetitive calls with same name", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		// Allow only single call to underlying Registry
		rMock.EXPECT().GetTemplatePK(name).Return(expPK, nil).Once()
		for i := 0; i < 10; i++ {
			got, err := cr.GetTemplatePK(name)
			assert.NoError(t, err)
			assert.Equal(t, expPK, got)
		}
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().GetTemplatePK(name).Return(0, fmt.Errorf("GetTemplatePK error"))
		got, err := cr.GetTemplatePK(name)
		assert.ErrorContains(t, err, "GetTemplatePK error")
		assert.Zero(t, got)
	})
}

func Test_CachedRegistry_GetTemplateContent(t *testing.T) {
	pk := 1
	expContent := " "
	t.Run("Cache hits after repetitive calls with same name", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		// Allow only single call to underlying Registry
		rMock.EXPECT().GetTemplateContent(pk).Return(&expContent, nil).Once()
		for i := 0; i < 10; i++ {
			got, err := cr.GetTemplateContent(pk)
			assert.NoError(t, err)
			assert.Equal(t, expContent, *got)
		}
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().GetTemplateContent(pk).Return(nil, fmt.Errorf("GetTemplateContent error"))
		got, err := cr.GetTemplateContent(pk)
		assert.ErrorContains(t, err, "GetTemplateContent error")
		assert.Nil(t, got)
	})
}

func Test_CachedRegistry_GetCertificate(t *testing.T) {
	id := " "
	expCert := Certificate{Id: id}
	t.Run("Cache hits after repetitive calls with same name", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		// Allow only single call to underlying Registry
		rMock.EXPECT().GetCertificate(id).Return(&expCert, nil).Once()
		for i := 0; i < 10; i++ {
			got, err := cr.GetCertificate(id)
			assert.NoError(t, err)
			assert.Equal(t, expCert, *got)
		}
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().GetCertificate(id).Return(nil, fmt.Errorf("GetCertificate error"))
		got, err := cr.GetCertificate(id)
		assert.ErrorContains(t, err, "GetCertificate error")
		assert.Nil(t, got)
	})
}

func Test_CachedRegistry_ListTemplates(t *testing.T) {
	expNames := []string{" ", " ", " "}
	t.Run("Cache hits after repetitive calls with same name", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		// Allow only single call to underlying Registry
		rMock.EXPECT().ListTemplates().Return(expNames, nil).Once()
		for i := 0; i < 10; i++ {
			got, err := cr.ListTemplates()
			assert.NoError(t, err)
			assert.ElementsMatch(t, expNames, got)
		}
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().ListTemplates().Return(nil, fmt.Errorf("ListTemplates error"))
		got, err := cr.ListTemplates()
		assert.ErrorContains(t, err, "ListTemplates error")
		assert.Nil(t, got)
	})
}

func Test_CachedRegistry_AddTemplate(t *testing.T) {
	name := "name"
	content := "content"
	t.Run("Registry returns no error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().AddTemplate(name, content).Return(nil)
		cr.getListTmplCache = append(cr.getListTmplCache, " ")
		err := cr.AddTemplate(name, content)
		assert.NoError(t, err)
		assert.Nil(t, cr.getListTmplCache)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().AddTemplate(name, content).Return(fmt.Errorf("AddTemplate error"))
		cr.getListTmplCache = append(cr.getListTmplCache, " ")
		err := cr.AddTemplate(name, content)
		assert.ErrorContains(t, err, "AddTemplate error")
		assert.NotNil(t, cr.getListTmplCache)
	})
}

func Test_CachedRegistry_CertificatesByTemplatePK(t *testing.T) {
	pk := 1
	expIds := []string{" ", " ", " "}
	t.Run("Registry returns no error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().CertificatesByTemplatePK(pk).Return(expIds, nil)
		got, err := cr.CertificatesByTemplatePK(pk)
		assert.NoError(t, err)
		assert.Equal(t, expIds, got)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().CertificatesByTemplatePK(pk).Return(nil, fmt.Errorf("CertificatesByTemplatePK error"))
		got, err := cr.CertificatesByTemplatePK(pk)
		assert.ErrorContains(t, err, "CertificatesByTemplatePK error")
		assert.Nil(t, got)
	})
}

func Test_CachedRegistry_DeleteTemplate(t *testing.T) {
	pk := 1
	name := "1"
	content := "content"
	templates := []string{"1", "2", "3"}
	t.Run("Registry returns no error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getTmplPkCache.Add(name, pkCached{pk})
		cr.getTmplContentCache.Add(pk, contentCached{&content})
		cr.getListTmplCache = templates
		rMock.EXPECT().DeleteTemplate(pk).Return(nil)
		err := cr.DeleteTemplate(pk)
		assert.NoError(t, err)
		ok := cr.getTmplPkCache.Contains(name)
		assert.False(t, ok)
		ok = cr.getTmplContentCache.Contains(pk)
		assert.False(t, ok)
		assert.Nil(t, cr.getListTmplCache)
	})
	t.Run("Registry returns error (DeleteTemplate)", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getTmplPkCache.Add(name, pkCached{pk})
		cr.getTmplContentCache.Add(pk, contentCached{&content})
		cr.getListTmplCache = templates
		rMock.EXPECT().DeleteTemplate(pk).Return(fmt.Errorf("DeleteTemplate error"))
		err := cr.DeleteTemplate(pk)
		assert.ErrorContains(t, err, "DeleteTemplate error")
		ok := cr.getTmplPkCache.Contains(name)
		assert.True(t, ok)
		ok = cr.getTmplContentCache.Contains(pk)
		assert.True(t, ok)
		assert.NotNil(t, cr.getListTmplCache)
	})
}

func Test_CachedRegistry_UpdateTemplate(t *testing.T) {
	pk := 1
	name := "1"
	content := "content"
	templates := []string{"1", "2", "3"}
	t.Run("Registry returns no error (\"name\": \"new name\")", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getTmplPkCache.Add(name, pkCached{pk})
		cr.getListTmplCache = templates
		rMock.EXPECT().UpdateTemplate(pk, map[string]string{"name": "new name"}).Return(nil)
		err := cr.UpdateTemplate(pk, map[string]string{"name": "new name"})
		assert.NoError(t, err)
		ok := cr.getTmplPkCache.Contains(name)
		assert.False(t, ok)
		assert.Nil(t, cr.getListTmplCache)
	})
	t.Run("Registry returns error (\"name\": \"new name\")", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getTmplPkCache.Add(name, pkCached{pk})
		cr.getListTmplCache = templates
		rMock.EXPECT().UpdateTemplate(pk, map[string]string{"name": "new name"}).
			Return(fmt.Errorf("UpdateTemplate error"))
		err := cr.UpdateTemplate(pk, map[string]string{"name": "new name"})
		assert.ErrorContains(t, err, "UpdateTemplate error")
		ok := cr.getTmplPkCache.Contains(name)
		assert.True(t, ok)
		assert.NotNil(t, cr.getListTmplCache)
	})

	id := "1"
	expIds := []string{"1", "2", "3"}
	cert := Certificate{}

	t.Run("Registry returns no error (\"content\": \"new content\")", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getTmplContentCache.Add(pk, contentCached{&content})
		cr.getCertificateCache.Add(id, certCached{&cert})
		rMock.EXPECT().UpdateTemplate(pk, map[string]string{"content": "new content"}).Return(nil)
		rMock.EXPECT().CertificatesByTemplatePK(pk).Return(expIds, nil)
		err := cr.UpdateTemplate(pk, map[string]string{"content": "new content"})
		assert.NoError(t, err)
		ok := cr.getTmplContentCache.Contains(pk)
		assert.False(t, ok)
		for _, ids := range expIds {
			ok = cr.getCertificateCache.Contains(ids)
			assert.False(t, ok)
		}
	})
	t.Run("Registry returns error (\"content\": \"new content\")", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getTmplContentCache.Add(pk, contentCached{&content})
		cr.getCertificateCache.Add(id, certCached{&cert})
		rMock.EXPECT().UpdateTemplate(pk, map[string]string{"content": "new content"}).Return(nil)
		rMock.EXPECT().CertificatesByTemplatePK(pk).Return(nil, fmt.Errorf("CertificatesByTemplatePK error"))
		err := cr.UpdateTemplate(pk, map[string]string{"content": "new content"})
		assert.ErrorContains(t, err, "CertificatesByTemplatePK error")
		ok := cr.getTmplContentCache.Contains(pk)
		assert.False(t, ok)
		assert.Zero(t, cr.getCertificateCache.Size())
	})

	t.Run("Registry returns error (\"name\": \"new name\" + \"content\": \"new content\" + UpdateTemplate)", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().UpdateTemplate(pk, map[string]string{"name": "new name", "content": "new content"}).
			Return(fmt.Errorf("UpdateTemplate error"))
		err := cr.UpdateTemplate(pk, map[string]string{"name": "new name", "content": "new content"})
		assert.ErrorContains(t, err, "UpdateTemplate error")
	})
}

func Test_CachedRegistry_DeleteCertificate(t *testing.T) {
	id := "1"
	expCert := Certificate{Id: id}
	t.Run("Registry returns no error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getCertificateCache.Add(id, certCached{&expCert})
		rMock.EXPECT().DeleteCertificate(id).Return(nil)
		err := cr.DeleteCertificate(id)
		assert.NoError(t, err)
		got, ok := cr.getCertificateCache.Peek(id)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getCertificateCache.Add(id, certCached{&expCert})
		rMock.EXPECT().DeleteCertificate(id).Return(fmt.Errorf("DeleteCertificate error"))
		err := cr.DeleteCertificate(id)
		assert.ErrorContains(t, err, "DeleteCertificate error")
		got, ok := cr.getCertificateCache.Peek(id)
		assert.True(t, ok)
		assert.Equal(t, got.cert, &expCert)
	})
}

func Test_CachedRegistry_AddCertificate(t *testing.T) {
	var (
		templateName = "test template"
		id           = "1"
		template_pk  = 1
		timestamp    = time.Now()
		student      = "test student"
		issueDate    = "test issue date"
		course       = "test course"
		mentors      = "test mentors"
	)
	expCert := Certificate{Id: id, TemplatePk: template_pk, Timestamp: timestamp, Student: student,
		IssueDate: issueDate, Course: course, Mentors: mentors}
	t.Run("Registry returns no error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().AddCertificate(templateName, student, issueDate, course, mentors).Return(&expCert, nil)
		got, err := cr.AddCertificate(templateName, student, issueDate, course, mentors)
		assert.NoError(t, err)
		assert.Equal(t, &expCert, got)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		rMock.EXPECT().AddCertificate(templateName, student, issueDate, course, mentors).Return(nil, fmt.Errorf("AddCertificate error"))
		got, err := cr.AddCertificate(templateName, student, issueDate, course, mentors)
		assert.ErrorContains(t, err, "AddCertificate error")
		assert.Nil(t, got)
	})
}

func Test_CachedRegistry_UpdateCertificate(t *testing.T) {
	var (
		id      = "1"
		expCert = Certificate{Id: id}
		m       = map[string]string{
			"template":   "test template",
			"student":    "test student",
			"issue_date": "test issue date",
			"course":     "test course",
			"mentors":    "test mentors",
		}
	)
	t.Run("Registry returns no error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getCertificateCache.Add(id, certCached{&expCert})
		rMock.EXPECT().UpdateCertificate(id, m).Return(nil)
		err := cr.UpdateCertificate(id, m)
		assert.NoError(t, err)
		got, ok := cr.getCertificateCache.Peek(id)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		cr, rMock := createTestCachedRegistry(t)
		cr.getCertificateCache.Add(id, certCached{&expCert})
		rMock.EXPECT().UpdateCertificate(id, m).Return(fmt.Errorf("AddCertificate error"))
		err := cr.UpdateCertificate(id, m)
		assert.ErrorContains(t, err, "AddCertificate error")
		got, ok := cr.getCertificateCache.Peek(id)
		assert.Equal(t, got.cert, &expCert)
		assert.True(t, ok)
	})
}
