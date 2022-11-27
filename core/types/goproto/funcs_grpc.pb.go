// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package goproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WritingClient is the client API for Writing service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WritingClient interface {
	Write(ctx context.Context, in *WriteContext, opts ...grpc.CallOption) (*Err, error)
}

type writingClient struct {
	cc grpc.ClientConnInterface
}

func NewWritingClient(cc grpc.ClientConnInterface) WritingClient {
	return &writingClient{cc}
}

func (c *writingClient) Write(ctx context.Context, in *WriteContext, opts ...grpc.CallOption) (*Err, error) {
	out := new(Err)
	err := c.cc.Invoke(ctx, "/Writing/Write", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WritingServer is the server API for Writing service.
// All implementations must embed UnimplementedWritingServer
// for forward compatibility
type WritingServer interface {
	Write(context.Context, *WriteContext) (*Err, error)
	mustEmbedUnimplementedWritingServer()
}

// UnimplementedWritingServer must be embedded to have forward compatible implementations.
type UnimplementedWritingServer struct {
}

func (UnimplementedWritingServer) Write(context.Context, *WriteContext) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Write not implemented")
}
func (UnimplementedWritingServer) mustEmbedUnimplementedWritingServer() {}

// UnsafeWritingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WritingServer will
// result in compilation errors.
type UnsafeWritingServer interface {
	mustEmbedUnimplementedWritingServer()
}

func RegisterWritingServer(s grpc.ServiceRegistrar, srv WritingServer) {
	s.RegisterService(&Writing_ServiceDesc, srv)
}

func _Writing_Write_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteContext)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WritingServer).Write(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Writing/Write",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WritingServer).Write(ctx, req.(*WriteContext))
	}
	return interceptor(ctx, in, info, handler)
}

// Writing_ServiceDesc is the grpc.ServiceDesc for Writing service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Writing_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Writing",
	HandlerType: (*WritingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Write",
			Handler:    _Writing_Write_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "funcs.proto",
}

// ReadingClient is the client API for Reading service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReadingClient interface {
	Read(ctx context.Context, in *ReadContext, opts ...grpc.CallOption) (*Err, error)
}

type readingClient struct {
	cc grpc.ClientConnInterface
}

func NewReadingClient(cc grpc.ClientConnInterface) ReadingClient {
	return &readingClient{cc}
}

func (c *readingClient) Read(ctx context.Context, in *ReadContext, opts ...grpc.CallOption) (*Err, error) {
	out := new(Err)
	err := c.cc.Invoke(ctx, "/Reading/Read", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReadingServer is the server API for Reading service.
// All implementations must embed UnimplementedReadingServer
// for forward compatibility
type ReadingServer interface {
	Read(context.Context, *ReadContext) (*Err, error)
	mustEmbedUnimplementedReadingServer()
}

// UnimplementedReadingServer must be embedded to have forward compatible implementations.
type UnimplementedReadingServer struct {
}

func (UnimplementedReadingServer) Read(context.Context, *ReadContext) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedReadingServer) mustEmbedUnimplementedReadingServer() {}

// UnsafeReadingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReadingServer will
// result in compilation errors.
type UnsafeReadingServer interface {
	mustEmbedUnimplementedReadingServer()
}

func RegisterReadingServer(s grpc.ServiceRegistrar, srv ReadingServer) {
	s.RegisterService(&Reading_ServiceDesc, srv)
}

func _Reading_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadContext)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReadingServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Reading/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReadingServer).Read(ctx, req.(*ReadContext))
	}
	return interceptor(ctx, in, info, handler)
}

// Reading_ServiceDesc is the grpc.ServiceDesc for Reading service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Reading_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Reading",
	HandlerType: (*ReadingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Read",
			Handler:    _Reading_Read_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "funcs.proto",
}
