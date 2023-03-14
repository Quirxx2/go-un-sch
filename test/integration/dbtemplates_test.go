//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	crt "gitlab.com/DzmitryYafremenka/golang-united-school-certs"
)

func Test_RegistryBehavior(t *testing.T) {
	type Entry struct {
		name    string
		content string
	}
	var entry = []Entry{
		{"Name 1", "Content 1"},
		{"Name 2", "Content 2"},
		{"Name 3", "Content 3"},
		{"Name 4", "Content 4"},
		{"Name 5", "Content 5"},
	}

	connString := "postgres://user:password@localhost:5432/registry"
	r, err := crt.NewDirectRegistry(connString)
	assert.NoError(t, err)

	var names []string
	for _, i := range entry {
		names = append(names, i.name)
		err = r.AddTemplate(i.name, i.content)
		assert.NoError(t, err)
	}

	n, err := r.ListTemplates()
	assert.ElementsMatch(t, names, n)
	assert.Nil(t, err)

	err = r.AddTemplate(entry[0].name, entry[0].content)
	assert.Error(t, err)

	n, err = r.ListTemplates()
	assert.ElementsMatch(t, names, n)
	assert.Nil(t, err)

	for j, i := range entry {
		pk, err := r.GetTemplatePK(i.name)
		assert.Equal(t, j+1, pk)
		assert.NoError(t, err)
		content, err := r.GetTemplateContent(pk)
		assert.Equal(t, i.content, *content)
		assert.NoError(t, err)
	}

	m := make(map[string]string)

	for j, i := range entry {
		m["name"] = i.name
		m["content"] = "New " + i.content
		err = r.UpdateTemplate(j+1, m)
		assert.Nil(t, err)
	}

	for j := range entry {
		err = r.DeleteTemplate(j + 1)
		assert.NoError(t, err)
	}

	n, err = r.ListTemplates()
	assert.Zero(t, len(n))
	assert.Nil(t, err)

	for j, i := range entry {
		pk, err := r.GetTemplatePK(i.name)
		assert.Equal(t, 0, pk)
		assert.Error(t, err)
		content, err := r.GetTemplateContent(j + 1)
		assert.Nil(t, content)
		assert.Error(t, err)
	}

	for j, i := range entry {
		m["name"] = i.name
		m["content"] = "New " + i.content
		err = r.UpdateTemplate(j+1, m)
		assert.Error(t, err)
	}
}
