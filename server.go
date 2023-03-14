package golangunitedschoolcerts

import (
	"context"
	"fmt"

	"gitlab.com/DzmitryYafremenka/golang-united-school-certs/api"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/emptypb"
)

type certsServer struct {
	api.UnimplementedCertsServiceServer
	r    Registry
	s    Storage
	t    Templater
	host string
}

func NewCertsServer(r Registry, s Storage, t Templater, host string) *certsServer {
	return &certsServer{r: r, s: s, t: t, host: host}
}

func (s *certsServer) AddTemplate(ctx context.Context, request *api.AddTemplateRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.r.AddTemplate(request.GetName(), request.GetContent())
}

func (s *certsServer) GetTemplate(ctx context.Context, request *api.GetTemplateRequest) (*api.GetTemplateResponse, error) {
	i, err := s.r.GetTemplatePK(request.GetName())
	if err != nil {
		return nil, err
	}
	c, err := s.r.GetTemplateContent(i)
	if c != nil {
		return &api.GetTemplateResponse{Content: *c}, nil
	}
	return nil, err
}

func (s *certsServer) DeleteTemplate(ctx context.Context, request *api.DeleteTemplateRequest) (*emptypb.Empty, error) {
	pk, err := s.r.GetTemplatePK(request.GetName())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, s.r.DeleteTemplate(pk)
}

func (s *certsServer) ListTemplates(ctx context.Context, in *emptypb.Empty) (*api.ListTemplatesResponse, error) {
	t, err := s.r.ListTemplates()
	if err == nil {
		return &api.ListTemplatesResponse{Names: t}, nil
	}
	return nil, err
}

func (s *certsServer) DeleteCertificate(ctx context.Context, request *api.DeleteCertificateRequest) (*emptypb.Empty, error) {
	cert, err := s.r.GetCertificate(request.GetId())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	err = s.r.DeleteCertificate(request.GetId())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	s.s.Delete(request.GetId(), cert.Timestamp)
	return &emptypb.Empty{}, nil
}

func (s *certsServer) UpdateTemplate(ctx context.Context, request *api.UpdateTemplateRequest) (*emptypb.Empty, error) {
	pk, err := s.r.GetTemplatePK(request.GetName())
	if err != nil {
		return &emptypb.Empty{}, err
	}
	m := make(map[string]string)
	if request.NewContent != nil {
		m["content"] = request.GetNewContent()
	}
	if request.NewName != nil {
		m["name"] = request.GetNewName()
	}
	if len(m) != 0 {
		return &emptypb.Empty{}, s.r.UpdateTemplate(pk, m)
	}
	return &emptypb.Empty{}, fmt.Errorf("no fields to update was provided")
}

func (s *certsServer) composeCertificateLink(id string) string {
	return s.host + "certificate/" + id
}

func (s *certsServer) GetCertificate(ctx context.Context, request *api.GetCertificateRequest) (*httpbody.HttpBody, error) {
	cert, err := s.r.GetCertificate(request.GetId())
	if err != nil {
		return nil, err
	}
	if s.s.Contains(cert.Id, cert.Timestamp) {
		pdf, err := s.s.Get(cert.Id, cert.Timestamp)
		if err != nil {
			return nil, err
		}
		return &httpbody.HttpBody{ContentType: "application/pdf", Data: *pdf}, nil
	}
	template, err := s.r.GetTemplateContent(cert.TemplatePk)
	if err != nil {
		return nil, err
	}
	pdf, err := s.t.GenerateCertificate(*template, cert, s.composeCertificateLink(cert.Id))
	if err != nil {
		return nil, err
	}
	err = s.s.Add(cert.Id, cert.Timestamp, pdf)
	if err != nil {
		return nil, err
	}
	return &httpbody.HttpBody{ContentType: "application/pdf", Data: *pdf}, nil
}

func (s *certsServer) TestTemplate(ctx context.Context, request *api.TestTemplateRequest) (*httpbody.HttpBody, error) {
	pk, err := s.r.GetTemplatePK(request.GetName())
	if err != nil {
		return nil, err
	}
	template, err := s.r.GetTemplateContent(pk)
	if err != nil {
		return nil, err
	}
	cert := Certificate{
		Id:        request.GetCertificate().GetId(),
		Student:   request.GetCertificate().GetStudent(),
		IssueDate: request.GetCertificate().GetIssueDate(),
		Course:    request.GetCertificate().GetCourse(),
		Mentors:   request.GetCertificate().GetMentors(),
	}
	pdf, err := s.t.GenerateCertificate(*template, &cert, s.composeCertificateLink(cert.Id))
	if err != nil {
		return nil, err
	}
	return &httpbody.HttpBody{ContentType: "application/pdf", Data: *pdf}, nil
}

func (s *certsServer) UpdateCertificate(ctx context.Context, request *api.UpdateCertificateRequest) (*emptypb.Empty, error) {
	m := make(map[string]string)
	if request.NewTemplate != nil {
		m["template"] = request.GetNewTemplate()
	}
	if request.NewStudent != nil {
		m["student"] = request.GetNewStudent()
	}
	if request.NewIssueDate != nil {
		m["issue_date"] = request.GetNewIssueDate()
	}
	if request.NewCourse != nil {
		m["course"] = request.GetNewCourse()
	}
	if request.NewMentors != nil {
		m["mentors"] = request.GetNewMentors()
	}
	if len(m) != 0 {
		return &emptypb.Empty{}, s.r.UpdateCertificate(request.GetId(), m)
	}
	return &emptypb.Empty{}, fmt.Errorf("no fields to update was provided")
}

func (s *certsServer) AddCertificate(ctx context.Context, request *api.AddCertificateRequest) (*api.AddCertificateResponse, error) {
	cert, err := s.r.AddCertificate(request.GetTemplateName(), request.GetStudent(), request.GetIssueDate(), request.GetCourse(), request.GetMentors())
	if err != nil {
		return nil, err
	}
	return &api.AddCertificateResponse{Id: cert.Id}, nil
}

func (s *certsServer) GetCertificateLink(ctx context.Context, request *api.GetCertificateLinkRequest) (*api.GetCertificateLinkResponse, error) {
	cert, err := s.r.GetCertificate(request.GetId())
	if err != nil {
		return nil, err
	}
	return &api.GetCertificateLinkResponse{Link: s.composeCertificateLink(cert.Id)}, nil
}
