package golangunitedschoolcerts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/slices"
)

type Registry interface {
	AddTemplate(string, string) error
	ListTemplates() ([]string, error)
	DeleteTemplate(int) error
	GetTemplatePK(string) (int, error)
	GetTemplateContent(int) (*string, error)
	CertificatesByTemplatePK(int) ([]string, error)
	UpdateTemplate(int, map[string]string) error
	AddCertificate(string, string, string, string, string) (*Certificate, error)
	DeleteCertificate(string) error
	GetCertificate(string) (*Certificate, error)
	UpdateCertificate(string, map[string]string) error
}

type DirectRegistry struct {
	p pool
}

// Interface for ease of mocking, exposing only used methods of pgxpool.Pool
type pool interface {
	Ping(context.Context) error
	Close()
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type Certificate struct {
	Id         string
	TemplatePk int
	Timestamp  time.Time
	Student    string
	IssueDate  string
	Course     string
	Mentors    string
}

func NewDirectRegistry(connString string) (*DirectRegistry, error) {
	p, err := initDB(connString)
	if err != nil {
		return nil, err
	}
	// Ping after creating new pool, to check if we can acquire connection
	if err = p.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DirectRegistry{p}, nil
}

func initDB(connString string) (p pool, err error) {
	if p, err = pgxpool.New(context.Background(), connString); err != nil {
		return nil, fmt.Errorf("failed to create pool connection to database %s: %w", connString, err)
	}
	return p, nil
}

func (dr *DirectRegistry) AddTemplate(name string, content string) (err error) {
	tx, err := dr.p.Begin(context.Background())
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(context.Background())
		default:
			_ = tx.Rollback(context.Background())
		}
	}()

	var id int
	row := tx.QueryRow(context.Background(),
		"INSERT INTO template_content (content) VALUES ($1) RETURNING id", content)
	err = row.Scan(&id)
	if err != nil {
		return fmt.Errorf("unable to INSERT INTO template_content: %w", err)
	}
	_, err = tx.Exec(context.Background(),
		"INSERT INTO template (name, content) VALUES ($1, $2)", name, id)
	if err != nil {
		return fmt.Errorf("unable to INSERT INTO template: %w", err)
	}
	return nil
}

func (dr *DirectRegistry) ListTemplates() (names []string, err error) {
	rows, err := dr.p.Query(context.Background(), "SELECT name FROM template")
	if err != nil {
		return nil, fmt.Errorf("unable to SELECT name FROM template: %w", err)
	}
	names, err = pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("unable to convert request into names list: %w", err)
	}
	return names, nil
}

func (dr *DirectRegistry) DeleteTemplate(pk int) (err error) {
	tx, err := dr.p.Begin(context.Background())
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(context.Background())
		default:
			_ = tx.Rollback(context.Background())
		}
	}()

	var contentId int
	row := tx.QueryRow(context.Background(),
		"SELECT content FROM template WHERE id=$1", pk)
	err = row.Scan(&contentId)
	if err != nil {
		return fmt.Errorf("unable to SELECT content FROM template: %w", err)
	}

	commandTag, err := tx.Exec(context.Background(),
		"DELETE FROM template WHERE id=$1", pk)
	if err != nil {
		return fmt.Errorf("unable to DELETE FROM template: %w", err)
	} else if commandTag.RowsAffected() != 1 {
		return errors.New("no row found to DELETE FROM template")
	}

	commandTag, err = tx.Exec(context.Background(),
		"DELETE FROM template_content WHERE id=$1", contentId)
	if err != nil {
		return fmt.Errorf("unable to DELETE FROM template_content: %w", err)
	} else if commandTag.RowsAffected() != 1 {
		return errors.New("no row found to DELETE FROM template_content")
	}

	return nil
}

func (dr *DirectRegistry) GetTemplatePK(name string) (pk int, err error) {
	row := dr.p.QueryRow(context.Background(),
		"SELECT id FROM template WHERE name=$1", name)
	err = row.Scan(&pk)
	if err != nil {
		return 0, fmt.Errorf("unable to SELECT id FROM template: %w", err)
	}
	return
}

func (dr *DirectRegistry) GetTemplateContent(pk int) (content *string, err error) {
	var contentId int
	row := dr.p.QueryRow(context.Background(),
		"SELECT content FROM template WHERE id=$1", pk)
	err = row.Scan(&contentId)
	if err != nil {
		return nil, fmt.Errorf("unable to SELECT content FROM template: %w", err)
	}
	row = dr.p.QueryRow(context.Background(),
		"SELECT content FROM template_content WHERE id=$1", contentId)
	var ctx string
	err = row.Scan(&ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to SELECT content FROM template_content: %w", err)
	}
	return &ctx, nil
}

func (dr *DirectRegistry) UpdateTemplate(pk int, m map[string]string) (err error) {
	tx, err := dr.p.Begin(context.Background())
	if err != nil {
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(context.Background())
		default:
			_ = tx.Rollback(context.Background())
		}
	}()

	for k, v := range m {
		switch k {
		case "name", "Name":
			commandTag, err := tx.Exec(context.Background(),
				"UPDATE template SET name=$1 WHERE id=$2",
				v, pk)
			if err != nil {
				return fmt.Errorf("unable to UPDATE template: %w", err)
			} else if commandTag.RowsAffected() != 1 {
				return errors.New("no row found to UPDATE template")
			}
		case "content", "Content":
			commandTag, err := tx.Exec(context.Background(),
				"UPDATE template_content SET content=$1 WHERE id=(SELECT content FROM template WHERE id=$2)",
				v, pk)
			if err != nil {
				return fmt.Errorf("unable to UPDATE template_content: %w", err)
			} else if commandTag.RowsAffected() != 1 {
				return errors.New("no row found to UPDATE template_content")
			}
		default:
			return fmt.Errorf("illegal key in a map")
		}
	}
	return nil
}

func (dr *DirectRegistry) AddCertificate(templateName, student, issueDate, course, mentors string) (*Certificate, error) {
	cert := &Certificate{}
	row := dr.p.QueryRow(context.Background(),
		"SELECT id FROM template WHERE name=$1", templateName)
	err := row.Scan(&cert.TemplatePk)
	if err != nil {
		return nil, fmt.Errorf("unable to scan Id for template %s: %w", templateName, err)
	}

	row = dr.p.QueryRow(context.Background(),
		`INSERT INTO certificate (template, student, issue_date, course, mentors)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, timestamp`,
		cert.TemplatePk, student, issueDate, course, mentors)
	err = row.Scan(&cert.Id, &cert.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("unable to scan Id and/or timestamp fields after INSERT INTO certificate: %w", err)
	}

	cert.Student = student
	cert.IssueDate = issueDate
	cert.Course = course
	cert.Mentors = mentors
	return cert, nil
}

func (dr *DirectRegistry) DeleteCertificate(id string) error {
	ct, err := dr.p.Exec(context.Background(),
		"DELETE FROM certificate WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("unable to DELETE FROM certificate: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected when attempt to DELETE FROM certificate with id: %v", id)
	}
	return nil
}

func (dr *DirectRegistry) GetCertificate(id string) (*Certificate, error) {
	cert := &Certificate{}
	row := dr.p.QueryRow(context.Background(),
		"SELECT template, timestamp, student, issue_date, course, mentors FROM certificate WHERE id=$1", id)
	err := row.Scan(&cert.TemplatePk, &cert.Timestamp, &cert.Student, &cert.IssueDate, &cert.Course, &cert.Mentors)
	if err != nil {
		return nil, fmt.Errorf("unable to get certificate with Id %s: %w", id, err)
	}

	cert.Id = id
	return cert, nil
}

func (dr *DirectRegistry) UpdateCertificate(id string, m map[string]string) error {
	fields := []string{
		"template", "Template",
		"student", "Student",
		"issue_date", "Issue_date",
		"course", "Course",
		"mentors", "Mentors"}
	s := []string{}
	for k, v := range m {
		if !slices.Contains(fields, k) {
			return fmt.Errorf("illegal key in a map")
		}
		s = append(s, fmt.Sprintf("%s='%s'", k, v))
	}
	q := fmt.Sprintf("UPDATE certificate SET %s WHERE id=$1", strings.Join(s, ","))
	ct, err := dr.p.Exec(context.Background(), q, id)
	if err != nil {
		return fmt.Errorf("unable to UPDATE certificate: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected when attempt to UPDATE certificate with id: %v", id)
	}
	return nil
}

func (dr *DirectRegistry) CertificatesByTemplatePK(pk int) (ids []string, err error) {
	rows, err := dr.p.Query(context.Background(), "SELECT id FROM certificate WHERE template=$1", pk)
	if err != nil {
		return nil, fmt.Errorf("unable to SELECT id FROM certificate: %w", err)
	}
	ids, err = pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("unable to convert request into ids list: %w", err)
	}
	return ids, nil
}
