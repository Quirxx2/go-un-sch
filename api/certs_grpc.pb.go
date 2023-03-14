// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: certs.proto

package api

import (
	context "context"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CertsServiceClient is the client API for CertsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CertsServiceClient interface {
	AddTemplate(ctx context.Context, in *AddTemplateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetTemplate(ctx context.Context, in *GetTemplateRequest, opts ...grpc.CallOption) (*GetTemplateResponse, error)
	DeleteTemplate(ctx context.Context, in *DeleteTemplateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListTemplates(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ListTemplatesResponse, error)
	DeleteCertificate(ctx context.Context, in *DeleteCertificateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UpdateTemplate(ctx context.Context, in *UpdateTemplateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetCertificate(ctx context.Context, in *GetCertificateRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
	TestTemplate(ctx context.Context, in *TestTemplateRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
	UpdateCertificate(ctx context.Context, in *UpdateCertificateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	AddCertificate(ctx context.Context, in *AddCertificateRequest, opts ...grpc.CallOption) (*AddCertificateResponse, error)
	GetCertificateLink(ctx context.Context, in *GetCertificateLinkRequest, opts ...grpc.CallOption) (*GetCertificateLinkResponse, error)
}

type certsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCertsServiceClient(cc grpc.ClientConnInterface) CertsServiceClient {
	return &certsServiceClient{cc}
}

func (c *certsServiceClient) AddTemplate(ctx context.Context, in *AddTemplateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/certs.CertsService/AddTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) GetTemplate(ctx context.Context, in *GetTemplateRequest, opts ...grpc.CallOption) (*GetTemplateResponse, error) {
	out := new(GetTemplateResponse)
	err := c.cc.Invoke(ctx, "/certs.CertsService/GetTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) DeleteTemplate(ctx context.Context, in *DeleteTemplateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/certs.CertsService/DeleteTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) ListTemplates(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ListTemplatesResponse, error) {
	out := new(ListTemplatesResponse)
	err := c.cc.Invoke(ctx, "/certs.CertsService/ListTemplates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) DeleteCertificate(ctx context.Context, in *DeleteCertificateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/certs.CertsService/DeleteCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) UpdateTemplate(ctx context.Context, in *UpdateTemplateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/certs.CertsService/UpdateTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) GetCertificate(ctx context.Context, in *GetCertificateRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error) {
	out := new(httpbody.HttpBody)
	err := c.cc.Invoke(ctx, "/certs.CertsService/GetCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) TestTemplate(ctx context.Context, in *TestTemplateRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error) {
	out := new(httpbody.HttpBody)
	err := c.cc.Invoke(ctx, "/certs.CertsService/TestTemplate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) UpdateCertificate(ctx context.Context, in *UpdateCertificateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/certs.CertsService/UpdateCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) AddCertificate(ctx context.Context, in *AddCertificateRequest, opts ...grpc.CallOption) (*AddCertificateResponse, error) {
	out := new(AddCertificateResponse)
	err := c.cc.Invoke(ctx, "/certs.CertsService/AddCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certsServiceClient) GetCertificateLink(ctx context.Context, in *GetCertificateLinkRequest, opts ...grpc.CallOption) (*GetCertificateLinkResponse, error) {
	out := new(GetCertificateLinkResponse)
	err := c.cc.Invoke(ctx, "/certs.CertsService/GetCertificateLink", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CertsServiceServer is the server API for CertsService service.
// All implementations must embed UnimplementedCertsServiceServer
// for forward compatibility
type CertsServiceServer interface {
	AddTemplate(context.Context, *AddTemplateRequest) (*emptypb.Empty, error)
	GetTemplate(context.Context, *GetTemplateRequest) (*GetTemplateResponse, error)
	DeleteTemplate(context.Context, *DeleteTemplateRequest) (*emptypb.Empty, error)
	ListTemplates(context.Context, *emptypb.Empty) (*ListTemplatesResponse, error)
	DeleteCertificate(context.Context, *DeleteCertificateRequest) (*emptypb.Empty, error)
	UpdateTemplate(context.Context, *UpdateTemplateRequest) (*emptypb.Empty, error)
	GetCertificate(context.Context, *GetCertificateRequest) (*httpbody.HttpBody, error)
	TestTemplate(context.Context, *TestTemplateRequest) (*httpbody.HttpBody, error)
	UpdateCertificate(context.Context, *UpdateCertificateRequest) (*emptypb.Empty, error)
	AddCertificate(context.Context, *AddCertificateRequest) (*AddCertificateResponse, error)
	GetCertificateLink(context.Context, *GetCertificateLinkRequest) (*GetCertificateLinkResponse, error)
	mustEmbedUnimplementedCertsServiceServer()
}

// UnimplementedCertsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCertsServiceServer struct {
}

func (UnimplementedCertsServiceServer) AddTemplate(context.Context, *AddTemplateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddTemplate not implemented")
}
func (UnimplementedCertsServiceServer) GetTemplate(context.Context, *GetTemplateRequest) (*GetTemplateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTemplate not implemented")
}
func (UnimplementedCertsServiceServer) DeleteTemplate(context.Context, *DeleteTemplateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTemplate not implemented")
}
func (UnimplementedCertsServiceServer) ListTemplates(context.Context, *emptypb.Empty) (*ListTemplatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTemplates not implemented")
}
func (UnimplementedCertsServiceServer) DeleteCertificate(context.Context, *DeleteCertificateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCertificate not implemented")
}
func (UnimplementedCertsServiceServer) UpdateTemplate(context.Context, *UpdateTemplateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTemplate not implemented")
}
func (UnimplementedCertsServiceServer) GetCertificate(context.Context, *GetCertificateRequest) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertificate not implemented")
}
func (UnimplementedCertsServiceServer) TestTemplate(context.Context, *TestTemplateRequest) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestTemplate not implemented")
}
func (UnimplementedCertsServiceServer) UpdateCertificate(context.Context, *UpdateCertificateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCertificate not implemented")
}
func (UnimplementedCertsServiceServer) AddCertificate(context.Context, *AddCertificateRequest) (*AddCertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddCertificate not implemented")
}
func (UnimplementedCertsServiceServer) GetCertificateLink(context.Context, *GetCertificateLinkRequest) (*GetCertificateLinkResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertificateLink not implemented")
}
func (UnimplementedCertsServiceServer) mustEmbedUnimplementedCertsServiceServer() {}

// UnsafeCertsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CertsServiceServer will
// result in compilation errors.
type UnsafeCertsServiceServer interface {
	mustEmbedUnimplementedCertsServiceServer()
}

func RegisterCertsServiceServer(s grpc.ServiceRegistrar, srv CertsServiceServer) {
	s.RegisterService(&CertsService_ServiceDesc, srv)
}

func _CertsService_AddTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).AddTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/AddTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).AddTemplate(ctx, req.(*AddTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_GetTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).GetTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/GetTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).GetTemplate(ctx, req.(*GetTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_DeleteTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).DeleteTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/DeleteTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).DeleteTemplate(ctx, req.(*DeleteTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_ListTemplates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).ListTemplates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/ListTemplates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).ListTemplates(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_DeleteCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).DeleteCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/DeleteCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).DeleteCertificate(ctx, req.(*DeleteCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_UpdateTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).UpdateTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/UpdateTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).UpdateTemplate(ctx, req.(*UpdateTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_GetCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).GetCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/GetCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).GetCertificate(ctx, req.(*GetCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_TestTemplate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestTemplateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).TestTemplate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/TestTemplate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).TestTemplate(ctx, req.(*TestTemplateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_UpdateCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).UpdateCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/UpdateCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).UpdateCertificate(ctx, req.(*UpdateCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_AddCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).AddCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/AddCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).AddCertificate(ctx, req.(*AddCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertsService_GetCertificateLink_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCertificateLinkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertsServiceServer).GetCertificateLink(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/certs.CertsService/GetCertificateLink",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertsServiceServer).GetCertificateLink(ctx, req.(*GetCertificateLinkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CertsService_ServiceDesc is the grpc.ServiceDesc for CertsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CertsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "certs.CertsService",
	HandlerType: (*CertsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddTemplate",
			Handler:    _CertsService_AddTemplate_Handler,
		},
		{
			MethodName: "GetTemplate",
			Handler:    _CertsService_GetTemplate_Handler,
		},
		{
			MethodName: "DeleteTemplate",
			Handler:    _CertsService_DeleteTemplate_Handler,
		},
		{
			MethodName: "ListTemplates",
			Handler:    _CertsService_ListTemplates_Handler,
		},
		{
			MethodName: "DeleteCertificate",
			Handler:    _CertsService_DeleteCertificate_Handler,
		},
		{
			MethodName: "UpdateTemplate",
			Handler:    _CertsService_UpdateTemplate_Handler,
		},
		{
			MethodName: "GetCertificate",
			Handler:    _CertsService_GetCertificate_Handler,
		},
		{
			MethodName: "TestTemplate",
			Handler:    _CertsService_TestTemplate_Handler,
		},
		{
			MethodName: "UpdateCertificate",
			Handler:    _CertsService_UpdateCertificate_Handler,
		},
		{
			MethodName: "AddCertificate",
			Handler:    _CertsService_AddCertificate_Handler,
		},
		{
			MethodName: "GetCertificateLink",
			Handler:    _CertsService_GetCertificateLink_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "certs.proto",
}
