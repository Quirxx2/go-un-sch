package golangunitedschoolcerts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/DzmitryYafremenka/golang-united-school-certs/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const host = "http://example.com/"

func initTestServerAndConn(t *testing.T, ctx context.Context) (rMock *MockRegistry, sMock *MockStorage, tMock *MockTemplater, client api.CertsServiceClient, closer func(), mux *runtime.ServeMux) {
	rMock = NewMockRegistry(t)
	sMock = NewMockStorage(t)
	tMock = NewMockTemplater(t)

	bufSize := 1024 * 1024
	lis := bufconn.Listen(bufSize)

	s := grpc.NewServer()
	api.RegisterCertsServiceServer(s, NewCertsServer(rMock, sMock, tMock, host))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("unexpected server exited with error: %v", err)
		}
	}()

	bufDialer := func(ctx context.Context, s string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		assert.FailNow(t, "unexpected error while dialing bufconn: %v", err)
	}

	client = api.NewCertsServiceClient(conn)

	mux = runtime.NewServeMux()
	err = api.RegisterCertsServiceHandlerClient(ctx, mux, client)
	if err != nil {
		assert.FailNow(t, "unexpected error while registering service handler: %v", err)
	}

	// automatically choosing open port
	sRest := httptest.NewServer(mux)

	closer = func() {
		s.Stop()
		sRest.Close()
	}

	return rMock, sMock, tMock, client, closer, mux
}

func Test_AddTemplate(t *testing.T) {
	name := "Name"
	content := "Test content"
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().AddTemplate(name, content).Return(nil)
		_, err := client.AddTemplate(ctx, &api.AddTemplateRequest{Name: name, Content: content})
		assert.NoError(t, err)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().AddTemplate(name, content).Return(fmt.Errorf("Registry error"))
		_, err := client.AddTemplate(ctx, &api.AddTemplateRequest{Name: name, Content: content})
		assert.ErrorContains(t, err, "Registry error")
	})

	t.Run("Send data through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().AddTemplate(name, content).Return(nil)
		body := `{"name": ` + `"` + name + `", "content": ` + `"` + content + `"}`

		req := httptest.NewRequest(http.MethodPost, "/template", strings.NewReader(body))
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
}

func Test_GetTemplate(t *testing.T) {
	name := "name"
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		exp := " "
		rMock.EXPECT().GetTemplatePK(name).Return(0, nil)
		rMock.EXPECT().GetTemplateContent(0).Return(&exp, nil)
		tpml, err := client.GetTemplate(ctx, &api.GetTemplateRequest{Name: name})
		assert.Equal(t, exp, tpml.Content)
		assert.NoError(t, err)
	})
	t.Run("Registry returns error (name not found)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(0, fmt.Errorf("Registry error (name not found)"))
		_, err := client.GetTemplate(ctx, &api.GetTemplateRequest{Name: name})
		assert.ErrorContains(t, err, "name not found")
	})
	t.Run("Registry returns error (content not found)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(0, nil)
		rMock.EXPECT().GetTemplateContent(0).Return(nil, fmt.Errorf("Registry error (content not found)"))
		_, err := client.GetTemplate(ctx, &api.GetTemplateRequest{Name: name})
		assert.ErrorContains(t, err, "content not found")
	})
	t.Run("Get data through REST proxy", func(t *testing.T) {
		exp := "something"
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(0, nil)
		rMock.EXPECT().GetTemplateContent(0).Return(&exp, nil)

		req := httptest.NewRequest(http.MethodGet, "/template/"+name, nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		tc, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(t, "failed to read response body: %v", err)
		}
		m := make(map[string]string)
		err = json.Unmarshal(tc, &m)
		if err != nil {
			assert.FailNow(t, "failed to unmarshal response: %v", err)
		}
		assert.Equal(t, exp, m["content"])
	})
}

func Test_DeleteTemplate(t *testing.T) {
	pk := 1
	name := "name"
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		rMock.EXPECT().DeleteTemplate(pk).Return(nil)
		_, err := client.DeleteTemplate(ctx, &api.DeleteTemplateRequest{Name: name})
		assert.NoError(t, err)
	})
	t.Run("Registry returns error (DeleteTemplate failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		rMock.EXPECT().DeleteTemplate(pk).Return(fmt.Errorf("DeleteTemplate error"))
		_, err := client.DeleteTemplate(ctx, &api.DeleteTemplateRequest{Name: name})
		assert.ErrorContains(t, err, "DeleteTemplate error")
	})
	t.Run("Registry returns error (GetTemplatePK failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(0, fmt.Errorf("GetTemplatePK error"))
		_, err := client.DeleteTemplate(ctx, &api.DeleteTemplateRequest{Name: name})
		assert.ErrorContains(t, err, "GetTemplatePK error")
	})
	t.Run("Delete data through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		rMock.EXPECT().DeleteTemplate(pk).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/template/"+name, nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
}

func Test_ListTemplates(t *testing.T) {
	expNames := []string{"1", "1", "1"}
	in := emptypb.Empty{}
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().ListTemplates().Return(expNames, nil)
		lt, err := client.ListTemplates(ctx, &in)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expNames, lt.Names)
	})
	t.Run("Registry returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().ListTemplates().Return(nil, fmt.Errorf("Registry error"))
		_, err := client.ListTemplates(ctx, &in)
		assert.ErrorContains(t, err, "Registry error")
	})
	t.Run("Get bunch of data through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().ListTemplates().Return(expNames, nil)

		req := httptest.NewRequest(http.MethodGet, "/templates", nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		lt, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(t, "failed to read response body: %v", err)
		}
		m := make(map[string][]string)
		err = json.Unmarshal(lt, &m)
		if err != nil {
			assert.FailNow(t, "failed to unmarshal response: %v", err)
		}
		assert.Equal(t, expNames, m["names"])
	})
}

func Test_DeleteCertificate(t *testing.T) {
	id := "1"
	cert := Certificate{}
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(&cert, nil)
		rMock.EXPECT().DeleteCertificate(id).Return(nil)
		sMock.EXPECT().Delete(id, cert.Timestamp)
		_, err := client.DeleteCertificate(ctx, &api.DeleteCertificateRequest{Id: id})
		assert.NoError(t, err)
	})
	t.Run("Registry returns error (DeleteCertificate failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(&cert, nil)
		rMock.EXPECT().DeleteCertificate(id).Return(fmt.Errorf("DeleteCertificate error"))
		_, err := client.DeleteCertificate(ctx, &api.DeleteCertificateRequest{Id: id})
		assert.ErrorContains(t, err, "DeleteCertificate error")
	})
	t.Run("Registry returns error (GetCertificate failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(nil, fmt.Errorf("GetCertificate error"))
		_, err := client.DeleteCertificate(ctx, &api.DeleteCertificateRequest{Id: id})
		assert.ErrorContains(t, err, "GetCertificate error")
	})
	t.Run("Delete data through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(&cert, nil)
		rMock.EXPECT().DeleteCertificate(id).Return(nil)
		sMock.EXPECT().Delete(id, cert.Timestamp)

		req := httptest.NewRequest(http.MethodDelete, "/certificate/"+id, nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
}

func Test_UpdateTemplate(t *testing.T) {
	pk := 1
	name := "name"
	m := make(map[string]string)
	nName := "new name"
	nContent := "new content"
	m["name"] = nName
	m["content"] = nContent
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		rMock.EXPECT().UpdateTemplate(pk, m).Return(nil)
		_, err := client.UpdateTemplate(ctx, &api.UpdateTemplateRequest{Name: name, NewName: &nName, NewContent: &nContent})
		assert.NoError(t, err)
	})
	t.Run("Registry returns error (UpdateTemplate failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		rMock.EXPECT().UpdateTemplate(pk, m).Return(fmt.Errorf("UpdateTemplate error"))
		_, err := client.UpdateTemplate(ctx, &api.UpdateTemplateRequest{Name: name, NewName: &nName, NewContent: &nContent})
		assert.ErrorContains(t, err, "UpdateTemplate error")
	})
	t.Run("Registry returns error (GetTemplatePK failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(0, fmt.Errorf("GetTemplatePK error"))
		_, err := client.UpdateTemplate(ctx, &api.UpdateTemplateRequest{Name: name, NewName: &nName, NewContent: &nContent})
		assert.ErrorContains(t, err, "GetTemplatePK error")
	})
	t.Run("Registry returns error (nothing to update)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		_, err := client.UpdateTemplate(ctx, &api.UpdateTemplateRequest{Name: name})
		assert.ErrorContains(t, err, "no fields to update was provided")
	})
	t.Run("Update data through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetTemplatePK(name).Return(pk, nil)
		rMock.EXPECT().UpdateTemplate(pk, m).Return(nil)
		body := `{"NewName": ` + `"` + nName + `", "NewContent": ` + `"` + nContent + `"}`

		req := httptest.NewRequest(http.MethodPatch, "/template/"+name, strings.NewReader(body))
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
}

func Test_composeCertificateLink(t *testing.T) {
	host := "http://example.com/"
	id := "12345678"
	expLink := host + "certificate/" + id

	s := NewCertsServer(nil, nil, nil, "http://example.com/")
	got := s.composeCertificateLink(id)
	assert.Equal(t, expLink, got)
}

func Test_GetCertificate(t *testing.T) {
	expCert := Certificate{
		Id:         "12345678",
		TemplatePk: 1,
		Timestamp:  time.Now(),
		Student:    "Test Student",
		IssueDate:  "1 December 1999",
		Course:     "Test Course",
		Mentors:    "Test Mentor One, Test Mentor Two",
	}
	expPdf := []byte{0, 1, 0, 1}
	expTemplate := `
	<h1> Test Template </h1>
	<p> Student: {{.Cert.Student}} </p>
	<p> IssueDate: {{.Cert.IssueDate}} </p>
	<p> Course: {{.Cert.Course}} </p>
	<p> Mentors: {{.Cert.Mentors}} </p>
	<p> URL: {{.Link}} </p>
	<img src="data:image/png;base64,{{.Qr}}"/>`
	expLink := host + "certificate/" + expCert.Id

	t.Run("Return certificate from storage", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(true)
		sMock.EXPECT().Get(expCert.Id, expCert.Timestamp).Return(&expPdf, nil)

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.NoError(t, err)
		assert.Equal(t, expPdf, got.GetData())
	})

	t.Run("Generate new certificate and return it", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, tMock, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(false)
		rMock.EXPECT().GetTemplateContent(expCert.TemplatePk).Return(&expTemplate, nil)
		tMock.EXPECT().GenerateCertificate(expTemplate, &expCert, expLink).Return(&expPdf, nil)
		sMock.EXPECT().Add(expCert.Id, expCert.Timestamp, &expPdf).Return(nil)

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.NoError(t, err)
		assert.NotEmpty(t, got.GetData())
	})

	t.Run("Registry GetCertificate returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		expErr := "Registry GetCertificate error"
		rMock.EXPECT().GetCertificate(expCert.Id).Return(nil, fmt.Errorf(expErr))

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("Storage Get returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(true)
		expErr := "Storage Get error"
		sMock.EXPECT().Get(expCert.Id, expCert.Timestamp).Return(nil, fmt.Errorf(expErr))

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("Registry GetTemplateContent returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(false)
		expErr := "Registry GetTemplateContent error"
		rMock.EXPECT().GetTemplateContent(expCert.TemplatePk).Return(nil, fmt.Errorf(expErr))

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("Templater GenerateCertificate returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, tMock, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(false)
		rMock.EXPECT().GetTemplateContent(expCert.TemplatePk).Return(&expTemplate, nil)
		expErr := "Templater GenerateCertificate error"
		tMock.EXPECT().GenerateCertificate(expTemplate, &expCert, expLink).Return(nil, fmt.Errorf(expErr))

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)

	})

	t.Run("Storage Add returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, tMock, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(false)
		rMock.EXPECT().GetTemplateContent(expCert.TemplatePk).Return(&expTemplate, nil)
		tMock.EXPECT().GenerateCertificate(expTemplate, &expCert, expLink).Return(&expPdf, nil)
		expErr := "Storage Add error"
		sMock.EXPECT().Add(expCert.Id, expCert.Timestamp, mock.Anything).Return(fmt.Errorf(expErr))

		got, err := client.GetCertificate(ctx, &api.GetCertificateRequest{Id: expCert.Id})
		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("Generate new certificate and return it through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, sMock, tMock, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetCertificate(expCert.Id).Return(&expCert, nil)
		sMock.EXPECT().Contains(expCert.Id, expCert.Timestamp).Return(false)
		rMock.EXPECT().GetTemplateContent(expCert.TemplatePk).Return(&expTemplate, nil)
		tMock.EXPECT().GenerateCertificate(expTemplate, &expCert, expLink).Return(&expPdf, nil)
		sMock.EXPECT().Add(expCert.Id, expCert.Timestamp, mock.Anything).Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/certificate/"+expCert.Id, nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		pdf, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(t, "failed to read response body: %v", err)
		}
		assert.NotEmpty(t, pdf)
	})
}

func Test_TestTemplate(t *testing.T) {
	expTemplateName := "Test Template"
	expTemplatePk := 1
	expTemplate := ""

	expCertRequest := api.TestTemplateRequest_TestCertificate{
		Id:        "12345678",
		Student:   "Test Student",
		IssueDate: "1 December 1999",
		Course:    "Test Course",
		Mentors:   "Test Mentor One, Test Mentor Two",
	}
	expCert := Certificate{
		Id:        expCertRequest.Id,
		Student:   expCertRequest.Student,
		IssueDate: expCertRequest.IssueDate,
		Course:    expCertRequest.Course,
		Mentors:   expCertRequest.Mentors,
	}
	expLink := host + "certificate/" + expCert.Id
	expPdf := []byte{0, 1, 0, 1}

	t.Run("Generate test certificate", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, tMock, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetTemplatePK(expTemplateName).Return(expTemplatePk, nil)
		rMock.EXPECT().GetTemplateContent(expTemplatePk).Return(&expTemplate, nil)
		tMock.EXPECT().GenerateCertificate(expTemplate, &expCert, expLink).Return(&expPdf, nil)

		got, err := client.TestTemplate(ctx, &api.TestTemplateRequest{Name: expTemplateName, Certificate: &expCertRequest})

		assert.NoError(t, err)
		assert.Equal(t, expPdf, got.GetData())
	})

	t.Run("Registry GetTemplatePK returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		expErr := "Registry GetTemplatePK error"
		rMock.EXPECT().GetTemplatePK(expTemplateName).Return(0, fmt.Errorf(expErr))

		got, err := client.TestTemplate(ctx, &api.TestTemplateRequest{Name: expTemplateName, Certificate: &expCertRequest})

		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("Registry GetTemplateContent returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetTemplatePK(expTemplateName).Return(expTemplatePk, nil)
		expErr := "Registry GetTemplateContent error"
		rMock.EXPECT().GetTemplateContent(expTemplatePk).Return(nil, fmt.Errorf(expErr))

		got, err := client.TestTemplate(ctx, &api.TestTemplateRequest{Name: expTemplateName, Certificate: &expCertRequest})

		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("GenerateCertificate returns error", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, tMock, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetTemplatePK(expTemplateName).Return(expTemplatePk, nil)
		rMock.EXPECT().GetTemplateContent(expTemplatePk).Return(&expTemplate, nil)
		expErr := "Templater GenerateCertificate error"
		tMock.EXPECT().GenerateCertificate(expTemplate, mock.Anything, expLink).Return(nil, fmt.Errorf(expErr))

		got, err := client.TestTemplate(ctx, &api.TestTemplateRequest{Name: expTemplateName, Certificate: &expCertRequest})

		assert.ErrorContains(t, err, expErr)
		assert.Nil(t, got)
	})

	t.Run("Send and get data through REST proxy", func(t *testing.T) {
		expTmplName := "testtmpl"
		ctx := context.Background()
		rMock, _, tMock, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()

		rMock.EXPECT().GetTemplatePK(expTmplName).Return(expTemplatePk, nil)
		rMock.EXPECT().GetTemplateContent(expTemplatePk).Return(&expTemplate, nil)
		tMock.EXPECT().GenerateCertificate(expTemplate, &expCert, expLink).Return(&expPdf, nil)

		body := `{"certificate": {"id": ` + `"` + expCertRequest.Id + `", "student": ` + `"` + expCertRequest.Student + `"` +
			`, "issueDate": ` + `"` + expCertRequest.IssueDate + `", "course": ` + `"` + expCertRequest.Course + `"` +
			`, "mentors": ` + `"` + expCertRequest.Mentors + `"}}`

		req := httptest.NewRequest(http.MethodPost, "/template/"+expTmplName+"/test", strings.NewReader(body))
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		gotPdf, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(t, "failed to read response body: %v", err)
		}

		assert.Equal(t, expPdf, gotPdf)
	})
}

func Test_UpdateCertificate(t *testing.T) {
	id := "123456789"
	template := "nTemplate"
	student := "nStudent"
	issue_date := "nIssueDate"
	course := "nCourse"
	mentors := "nMentors"
	m := make(map[string]string)
	m["template"] = template
	m["student"] = student
	m["issue_date"] = issue_date
	m["course"] = course
	m["mentors"] = mentors
	t.Run("Successful", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().UpdateCertificate(id, m).Return(nil)
		_, err := client.UpdateCertificate(ctx, &api.UpdateCertificateRequest{Id: id, NewTemplate: &template, NewStudent: &student, NewIssueDate: &issue_date, NewCourse: &course, NewMentors: &mentors})
		assert.NoError(t, err)
	})
	t.Run("Registry returns error (UpdateCertificate failed)", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().UpdateCertificate(id, m).Return(fmt.Errorf("UpdateCertificate error"))
		_, err := client.UpdateCertificate(ctx, &api.UpdateCertificateRequest{Id: id, NewTemplate: &template, NewStudent: &student, NewIssueDate: &issue_date, NewCourse: &course, NewMentors: &mentors})
		assert.ErrorContains(t, err, "UpdateCertificate error")
	})
	t.Run("Registry returns error (nothing to update)", func(t *testing.T) {
		ctx := context.Background()
		_, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		_, err := client.UpdateCertificate(ctx, &api.UpdateCertificateRequest{Id: id})
		assert.ErrorContains(t, err, "no fields to update was provided")
	})
	t.Run("Update data through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().UpdateCertificate(id, m).Return(nil)
		body := `{"NewTemplate": ` + `"` + template + `", "NewStudent": ` + `"` + student +
			`", "NewIssueDate": ` + `"` + issue_date + `", "NewCourse": ` + `"` + course +
			`", "NewMentors": ` + `"` + mentors + `"}`

		req := httptest.NewRequest(http.MethodPatch, "/certificate/"+id, strings.NewReader(body))
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	})
}

func Test_AddCertificate(t *testing.T) {
	expCert := &Certificate{Id: "1"}
	templateName := "test template"
	student := "test student"
	issueDate := "test issue date"
	course := "test course"
	mentors := "test mentors"
	t.Run("Successfull adding. No errors returns", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().AddCertificate(templateName, student, issueDate, course, mentors).Return(expCert, nil)
		got, err := client.AddCertificate(ctx, &api.AddCertificateRequest{TemplateName: templateName, Student: student, IssueDate: issueDate, Course: course, Mentors: mentors})
		assert.Equal(t, got.Id, expCert.Id)
		assert.NoError(t, err)
	})
	t.Run("Failed adding. Error returns", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().AddCertificate(templateName, student, issueDate, course, mentors).Return(nil, fmt.Errorf("AddCertificate error"))
		got, err := client.AddCertificate(ctx, &api.AddCertificateRequest{TemplateName: templateName, Student: student, IssueDate: issueDate, Course: course, Mentors: mentors})
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "AddCertificate error")
	})
	t.Run("Add certificate and receive certificate.Id through REST proxy", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().AddCertificate(templateName, student, issueDate, course, mentors).Return(expCert, nil)

		body := `{"templateName": ` + `"` + templateName + `", "student": ` + `"` + student + `"` +
			`, "issueDate": ` + `"` + issueDate + `", "course": ` + `"` + course + `"` +
			`, "mentors": ` + `"` + mentors + `"}`

		req := httptest.NewRequest(http.MethodPost, "/certificate", strings.NewReader(body))
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		got, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(t, "failed to read response body: %v", err)
		}

		assert.NotEmpty(t, got)

		m := make(map[string]string)
		err = json.Unmarshal(got, &m)
		if err != nil {
			assert.FailNow(t, "failed to unmarshal response: %v", err)
		}
		assert.Equal(t, expCert.Id, m["id"])

	})
}

func Test_GetCertificateLink(t *testing.T) {
	host := "http://example.com/"
	id := "12345678"
	expLink := host + "certificate/" + id
	expCert := &Certificate{Id: id}

	t.Run("Successfull getting certificate link. No errors returns", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(expCert, nil)
		got, err := client.GetCertificateLink(ctx, &api.GetCertificateLinkRequest{Id: id})
		assert.Equal(t, got.Link, expLink)
		assert.NoError(t, err)
	})
	t.Run("Failed getting certificate link. No certificate with Id Error returns", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, client, closer, _ := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(nil, fmt.Errorf("GetCertificate error"))
		got, err := client.GetCertificateLink(ctx, &api.GetCertificateLinkRequest{Id: id})
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "GetCertificate error")
	})
	t.Run("Get certificate link through REST proxe", func(t *testing.T) {
		ctx := context.Background()
		rMock, _, _, _, closer, mux := initTestServerAndConn(t, ctx)
		defer closer()
		rMock.EXPECT().GetCertificate(id).Return(expCert, nil)

		req := httptest.NewRequest(http.MethodGet, "/certificate/"+expCert.Id+"/link", nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)

		got, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(t, "failed to read response body: %v", err)
		}

		assert.NotEmpty(t, got)

		m := make(map[string]string)
		err = json.Unmarshal(got, &m)
		if err != nil {
			assert.FailNow(t, "failed to unmarshal response: %v", err)
		}
		assert.Equal(t, expLink, m["link"])
	})
}
