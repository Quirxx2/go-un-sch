package golangunitedschoolcerts

import "fmt"

type CachedRegistry struct {
	r                   Registry
	getTmplPkCache      Cache[string, pkCached]
	getTmplContentCache Cache[int, contentCached]
	getCertificateCache Cache[string, certCached]
	getListTmplCache    []string
}

type pkCached struct {
	pk int
}

func (p pkCached) Size() int {
	return 1
}

type contentCached struct {
	content *string
}

func (p contentCached) Size() int {
	return 1
}

type certCached struct {
	cert *Certificate
}

func (p certCached) Size() int {
	return 1
}

func NewCachedRegistry(r Registry) (cr *CachedRegistry, err error) {
	cr = &CachedRegistry{}
	cr.r = r
	if c, err := NewLRUCache[string, pkCached](0, nil); err != nil {
		return nil, fmt.Errorf("failed to create getTmplPkCache: %w", err)
	} else {
		cr.getTmplPkCache = NewSafeCache[string, pkCached](c)
	}
	if c, err := NewLRUCache[int, contentCached](0, nil); err != nil {
		return nil, fmt.Errorf("failed to create getTmplContentCache: %w", err)
	} else {
		cr.getTmplContentCache = NewSafeCache[int, contentCached](c)
	}
	if c, err := NewLRUCache[string, certCached](0, nil); err != nil {
		return nil, fmt.Errorf("failed to create getCertificateCache: %w", err)
	} else {
		cr.getCertificateCache = NewSafeCache[string, certCached](c)
	}
	return cr, nil
}

func (cr *CachedRegistry) GetTemplatePK(name string) (pk int, err error) {
	if pc, ok := cr.getTmplPkCache.Get(name); ok {
		return pc.pk, nil
	}
	if pk, err = cr.r.GetTemplatePK(name); err != nil {
		return 0, err
	}
	cr.getTmplPkCache.Add(name, pkCached{pk})
	return pk, nil
}

func (cr *CachedRegistry) GetTemplateContent(pk int) (content *string, err error) {
	cc, ok := cr.getTmplContentCache.Get(pk)
	if ok {
		return cc.content, nil
	}
	content, err = cr.r.GetTemplateContent(pk)
	if err != nil {
		return nil, err
	}
	cr.getTmplContentCache.Add(pk, contentCached{content})
	return content, nil
}

func (cr *CachedRegistry) GetCertificate(id string) (cert *Certificate, err error) {
	cc, ok := cr.getCertificateCache.Get(id)
	if ok {
		return cc.cert, nil
	}
	cert, err = cr.r.GetCertificate(id)
	if err != nil {
		return nil, err
	}
	cr.getCertificateCache.Add(id, certCached{cert})
	return cert, nil
}

func (cr *CachedRegistry) ListTemplates() (names []string, err error) {
	lc := cr.getListTmplCache
	if lc != nil {
		return lc, nil
	}
	lc, err = cr.r.ListTemplates()
	if err != nil {
		return nil, err
	}
	cr.getListTmplCache = lc
	return lc, nil
}

func (cr *CachedRegistry) AddTemplate(name string, content string) (err error) {
	err = cr.r.AddTemplate(name, content)
	if err != nil {
		return err
	}
	cr.getListTmplCache = nil
	return nil
}

func (cr *CachedRegistry) CertificatesByTemplatePK(pk int) (ids []string, err error) {
	return cr.r.CertificatesByTemplatePK(pk)
}

func (cr *CachedRegistry) DeleteTemplate(pk int) (err error) {
	err = cr.r.DeleteTemplate(pk)
	if err != nil {
		return
	}
	names := cr.getTmplPkCache.Keys()
	for _, name := range names {
		tmp, ok := cr.getTmplPkCache.Peek(name)
		if ok && pk == tmp.pk {
			cr.getTmplPkCache.Remove(name)
			break
		}
	}
	cr.getTmplContentCache.Remove(pk)
	cr.getListTmplCache = nil
	return nil
}

func (cr *CachedRegistry) UpdateTemplate(pk int, m map[string]string) (err error) {
	err = cr.r.UpdateTemplate(pk, m)
	if err != nil {
		return
	}
	for k := range m {
		switch k {
		case "name", "Name":
			names := cr.getTmplPkCache.Keys()
			for _, name := range names {
				tmp, ok := cr.getTmplPkCache.Peek(name)
				if ok && pk == tmp.pk {
					cr.getTmplPkCache.Remove(name)
					break
				}
			}
			cr.getListTmplCache = nil
		case "content", "Content":
			cr.getTmplContentCache.Remove(pk)
			ids, err := cr.r.CertificatesByTemplatePK(pk)
			if err != nil {
				cr.getCertificateCache.Purge()
				return err
			} else {
				for _, id := range ids {
					cr.getCertificateCache.Remove(id)
				}
			}
		}
	}
	return
}

func (cr *CachedRegistry) DeleteCertificate(id string) error {
	if err := cr.r.DeleteCertificate(id); err != nil {
		return err
	}
	cr.getCertificateCache.Remove(id)
	return nil
}

func (cr *CachedRegistry) AddCertificate(templateName, student, issueDate, course, mentors string) (*Certificate, error) {
	return cr.r.AddCertificate(templateName, student, issueDate, course, mentors)
}

func (cr *CachedRegistry) UpdateCertificate(id string, m map[string]string) error {
	if err := cr.r.UpdateCertificate(id, m); err != nil {
		return err
	}
	cr.getCertificateCache.Remove(id)
	return nil
}
