package golangunitedschoolcerts

import (
	"fmt"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

// Check that struct implements interface
var _ Registry = &DirectRegistry{}

func Test_DirectRegistry_AddTemplate(t *testing.T) {
	t.Run("Check inserting into template and template_content tables (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		id := 1
		name := "Test name"
		content := "Test content"
		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"id"}).AddRow(id)
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO template_content").WithArgs(content).
			WillReturnRows(rows)
		mock.ExpectExec("INSERT INTO template").WithArgs(name, id).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectCommit()

		err = dr.AddTemplate(name, content)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check inserting into template and template_content tables (wrong id)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		id := 1
		name := "Test name"
		content := "Test content"
		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"id"}).AddRow(id)
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO template_content").WithArgs(content).
			WillReturnRows(rows)
		mock.ExpectExec("INSERT INTO template").WithArgs(name, id).
			WillReturnError(fmt.Errorf("id error"))
		mock.ExpectRollback()

		err = dr.AddTemplate(name, content)
		assert.ErrorContains(t, err, "id error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check inserting into template and template_content tables (wrong content)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		name := "Test name"
		content := "Test content"
		dr := &DirectRegistry{mock}
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO template_content").WithArgs(content).
			WillReturnError(fmt.Errorf("content error"))
		mock.ExpectRollback()

		err = dr.AddTemplate(name, content)
		assert.ErrorContains(t, err, "content error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_ListTemplates(t *testing.T) {
	t.Run("Check retrieving template names from template table (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		rows := mock.NewRows([]string{"name"})
		expNames := []string{"Test name 1", "Test name 1", "Test name 3"}
		for _, name := range expNames {
			rows.AddRow(name)
		}
		mock.ExpectQuery("SELECT name FROM").WillReturnRows(rows)

		templates, err := dr.ListTemplates()
		assert.ElementsMatch(t, templates, expNames)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check retrieving template names from template table (wrong template names list)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		rows := mock.NewRows([]string{"name"})
		expNames := []string{"Test name 1", "Test name 1", "Test name 3"}
		for _, name := range expNames {
			rows.AddRow(name)
		}
		mock.ExpectQuery("SELECT name FROM").
			WillReturnError(fmt.Errorf("names list error"))

		templates, err := dr.ListTemplates()
		assert.Nil(t, templates)
		assert.ErrorContains(t, err, "names list error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_DeleteTemplate(t *testing.T) {
	t.Run("Check deleting from template and template_content tables (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		id := 1
		pk := 1
		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"content"}).AddRow(id)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT content FROM template").WithArgs(pk).
			WillReturnRows(rows)
		mock.ExpectExec("DELETE FROM template").WithArgs(pk).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectExec("DELETE FROM template_content").WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()

		err = dr.DeleteTemplate(pk)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check deleting from template and template_content tables (wrong id)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		id := 1
		pk := 1
		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"content"}).AddRow(id)
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT content FROM template").WithArgs(pk).
			WillReturnRows(rows)
		mock.ExpectExec("DELETE FROM template").WithArgs(pk).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectExec("DELETE FROM template_content").WithArgs(pgxmock.AnyArg()).
			WillReturnError(fmt.Errorf("id error"))
		mock.ExpectRollback()

		err = dr.DeleteTemplate(pk)
		assert.ErrorContains(t, err, "id error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check deleting from template and template_content tables (wrong template name)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT content FROM template").WithArgs(pgxmock.AnyArg()).
			WillReturnError(fmt.Errorf("template name error"))
		mock.ExpectRollback()

		err = dr.DeleteTemplate(pk)
		assert.ErrorContains(t, err, "template name error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_GetTemplatePK(t *testing.T) {
	t.Run("Check retrieving template id from template table (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		name := "Test name"
		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT id FROM").WithArgs(name).
			WillReturnRows(rows)

		pk, err := dr.GetTemplatePK(name)
		assert.NotZero(t, pk)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check retrieving template id from template table (wrong name)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		name := "Test name"
		dr := &DirectRegistry{mock}
		mock.ExpectQuery("SELECT id FROM").WithArgs(name).
			WillReturnError(fmt.Errorf("name error"))

		pk, err := dr.GetTemplatePK(name)
		assert.Zero(t, pk)
		assert.ErrorContains(t, err, "name error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_GetTemplateContent(t *testing.T) {
	t.Run("Check retrieving template content from template_content table (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		rows1 := pgxmock.NewRows([]string{"content"}).AddRow(1)
		mock.ExpectQuery("SELECT content FROM").WithArgs(pk).
			WillReturnRows(rows1)
		rows2 := pgxmock.NewRows([]string{"content"}).AddRow("Test content")
		mock.ExpectQuery("SELECT content FROM").WithArgs(pk).
			WillReturnRows(rows2)

		content, err := dr.GetTemplateContent(pk)
		assert.NotEqual(t, content, nil)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check retrieving template content from template_content table (wrong template id)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		mock.ExpectQuery("SELECT content FROM").WithArgs(pgxmock.AnyArg()).
			WillReturnError(fmt.Errorf("template id error"))

		content, err := dr.GetTemplateContent(pk)
		assert.Nil(t, content)
		assert.ErrorContains(t, err, "template id error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check retrieving template content from template_content table (wrong template_content id)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		rows1 := pgxmock.NewRows([]string{"content"}).AddRow(1)
		mock.ExpectQuery("SELECT content FROM").WithArgs(pk).
			WillReturnRows(rows1)
		mock.ExpectQuery("SELECT content FROM").WithArgs(pgxmock.AnyArg()).
			WillReturnError(fmt.Errorf("template_content id error"))

		content, err := dr.GetTemplateContent(pk)
		assert.Nil(t, content)
		assert.ErrorContains(t, err, "template_content id error")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_UpdateTemplate(t *testing.T) {
	t.Run("Check updating template and template_content tables (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		mock.MatchExpectationsInOrder(false)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE template").WithArgs("new name", pk).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectExec("UPDATE template_content").WithArgs("new content", pk).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()

		err = dr.UpdateTemplate(pk, map[string]string{"name": "new name", "content": "new content"})
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check updating template and template_content tables (wrong id - testing name field)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		mock.MatchExpectationsInOrder(false)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE template").WithArgs("new name", pk).
			WillReturnError(fmt.Errorf("no row found to UPDATE template"))
		mock.ExpectRollback()

		err = dr.UpdateTemplate(pk, map[string]string{"name": "new name"})
		assert.ErrorContains(t, err, "no row found to UPDATE template")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check updating template and template_content tables (wrong id - testing content field)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		mock.MatchExpectationsInOrder(false)
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE template_content").WithArgs("new content", pk).
			WillReturnError(fmt.Errorf("no row found to UPDATE template_content"))
		mock.ExpectRollback()

		err = dr.UpdateTemplate(pk, map[string]string{"content": "new content"})
		assert.ErrorContains(t, err, "no row found to UPDATE template_content")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check updating template and template_content tables (attempting to update an unknown field)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		mock.ExpectBegin()
		mock.ExpectRollback()
		err = dr.UpdateTemplate(pk, map[string]string{"mistake key": "mistake value"})
		assert.ErrorContains(t, err, "illegal key in a map")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_AddCertificate(t *testing.T) {
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

	t.Run("Check inserting into certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"id"}).AddRow(template_pk)
		mock.ExpectQuery("SELECT id FROM template").WithArgs(templateName).WillReturnRows(rows)
		rows = pgxmock.NewRows([]string{"id", "timestamp"}).AddRow(id, timestamp)
		mock.ExpectQuery("INSERT INTO certificate").WithArgs(template_pk, student, issueDate, course, mentors).WillReturnRows(rows)

		cert, err := dr.AddCertificate(templateName, student, issueDate, course, mentors)
		assert.Equal(t, &Certificate{id, template_pk, timestamp, student, issueDate, course, mentors}, cert)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error when it is unable to select id from certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		rows := pgxmock.NewRows([]string{"id"})
		dr := &DirectRegistry{mock}
		mock.ExpectQuery("SELECT id FROM template").WithArgs(templateName).WillReturnRows(rows)

		cert, err := dr.AddCertificate(templateName, student, issueDate, course, mentors)
		assert.Nil(t, cert)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error whet it is unable to insert into certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"id"}).AddRow(template_pk)
		mock.ExpectQuery("SELECT id FROM template").WithArgs(templateName).WillReturnRows(rows)
		rows = pgxmock.NewRows([]string{"id", "timestamp"})
		mock.ExpectQuery("INSERT INTO certificate").WithArgs(template_pk, student, issueDate, course, mentors).WillReturnRows(rows)

		cert, err := dr.AddCertificate(templateName, student, issueDate, course, mentors)
		assert.Nil(t, cert)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_DeleteCertificate(t *testing.T) {
	var id = "1"

	t.Run("Check deleting from certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("DELETE FROM certificate").WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = dr.DeleteCertificate(id)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error when it is unable to delete from certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("DELETE FROM certificate").WithArgs(id).WillReturnError(fmt.Errorf("some error"))

		err = dr.DeleteCertificate(id)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error when there is no affected rows to delete from certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("DELETE FROM certificate").WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = dr.DeleteCertificate(id)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
	})

}

func Test_DirectRegistry_GetCertificate(t *testing.T) {
	var id = "1"
	t.Run("Check getting certificate by Id", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		template := 1
		timestamp := time.Now()
		student := "test student"
		issueDate := "test issue date"
		course := "test course"
		mentors := "test mentors"
		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"template", "timestamp", "student", "issue_date", "course", "mentors"}).
			AddRow(template, timestamp, student, issueDate, course, mentors)
		mock.ExpectQuery("SELECT template, timestamp, student, issue_date, course, mentors").WithArgs(id).WillReturnRows(rows)

		cert, err := dr.GetCertificate(id)
		assert.Equal(t, &Certificate{id, template, timestamp, student, issueDate, course, mentors}, cert)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error when getting certificate by Id", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		rows := pgxmock.NewRows([]string{"template", "timestamp", "student", "issue_date", "course", "mentors"})
		mock.ExpectQuery("SELECT template, timestamp").WithArgs(id).WillReturnRows(rows)

		cert, err := dr.GetCertificate(id)
		assert.Nil(t, cert)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_CertificatesByTemplatePK(t *testing.T) {
	t.Run("Check retrieving template pks from certificate table (no errors)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		rows := mock.NewRows([]string{"id"})
		expIds := []string{"1", "2", "3"}
		for _, id := range expIds {
			rows.AddRow(id)
		}
		mock.ExpectQuery("SELECT id FROM").WithArgs(pk).WillReturnRows(rows)

		ids, err := dr.CertificatesByTemplatePK(pk)
		assert.ElementsMatch(t, ids, expIds)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check retrieving template pks from certificate table (wrong template id)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		rows := mock.NewRows([]string{"id"})
		expIds := []string{"1", "2", "3"}
		for _, id := range expIds {
			rows.AddRow(id)
		}
		mock.ExpectQuery("SELECT id FROM").WithArgs(pk).WillReturnRows(&pgxmock.Rows{})

		ids, err := dr.CertificatesByTemplatePK(pk)
		assert.NotEqualValues(t, ids, expIds)
		assert.Nil(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check retrieving template pks from certificate table (wrong certificate id's list)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		pk := 1
		dr := &DirectRegistry{mock}
		rows := mock.NewRows([]string{"id"})
		expIds := []string{"1", "2", "3"}
		for _, id := range expIds {
			rows.AddRow(id)
		}
		mock.ExpectQuery("SELECT id FROM").WithArgs(pk).
			WillReturnError(fmt.Errorf("wrong certificate id's list"))

		ids, err := dr.CertificatesByTemplatePK(pk)
		assert.Nil(t, ids)
		assert.ErrorContains(t, err, "wrong certificate id's list")
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})
}

func Test_DirectRegistry_UpdateCertificate(t *testing.T) {
	var (
		id = "1"
		m  = map[string]string{
			"template":   "test template",
			"student":    "test student",
			"issue_date": "test issue date",
			"course":     "test course",
			"mentors":    "test mentors",
		}
	)

	t.Run("Check updating certificate table with one field", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("UPDATE certificate SET").WithArgs(id).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = dr.UpdateCertificate(id, map[string]string{"course": "test course"})
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Check updating certificate table with multiple fields", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("UPDATE certificate SET").WithArgs(id).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = dr.UpdateCertificate(id, m)
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoErrorf(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error when it is unable to update certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("UPDATE certificate SET").WithArgs(id).WillReturnError(fmt.Errorf("some error"))

		err = dr.UpdateCertificate(id, m)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")

	})

	t.Run("Expecting error when there is no affected rows to update certificate table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}
		mock.ExpectExec("UPDATE certificate SET").WithArgs(id).WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err = dr.UpdateCertificate(id, m)
		assert.Error(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err, "there were unfulfilled expectations")
	})

	t.Run("Expecting error when there is illegal key in map", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening mock", err)
		}
		defer mock.Close()

		dr := &DirectRegistry{mock}

		// There is no need to add mock.ExpectExec, error returns earlier
		err = dr.UpdateCertificate(id, map[string]string{"illegal_key": "some value"})
		assert.Error(t, err)
	})
}
