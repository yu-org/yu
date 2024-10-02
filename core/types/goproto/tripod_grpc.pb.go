// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v4.25.2
// source: tripod.proto

package goproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Tripod_CheckTxn_FullMethodName      = "/Tripod/CheckTxn"
	Tripod_VerifyBlock_FullMethodName   = "/Tripod/VerifyBlock"
	Tripod_StartBlock_FullMethodName    = "/Tripod/StartBlock"
	Tripod_EndBlock_FullMethodName      = "/Tripod/EndBlock"
	Tripod_FinalizeBlock_FullMethodName = "/Tripod/FinalizeBlock"
)

// TripodClient is the client API for Tripod service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TripodClient interface {
	CheckTxn(ctx context.Context, in *TripodTxnRequest, opts ...grpc.CallOption) (*Err, error)
	VerifyBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error)
	StartBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error)
	EndBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error)
	FinalizeBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error)
}

type tripodClient struct {
	cc grpc.ClientConnInterface
}

func NewTripodClient(cc grpc.ClientConnInterface) TripodClient {
	return &tripodClient{cc}
}

func (c *tripodClient) CheckTxn(ctx context.Context, in *TripodTxnRequest, opts ...grpc.CallOption) (*Err, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Err)
	err := c.cc.Invoke(ctx, Tripod_CheckTxn_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tripodClient) VerifyBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Err)
	err := c.cc.Invoke(ctx, Tripod_VerifyBlock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tripodClient) StartBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Err)
	err := c.cc.Invoke(ctx, Tripod_StartBlock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tripodClient) EndBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Err)
	err := c.cc.Invoke(ctx, Tripod_EndBlock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tripodClient) FinalizeBlock(ctx context.Context, in *TripodBlockRequest, opts ...grpc.CallOption) (*Err, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Err)
	err := c.cc.Invoke(ctx, Tripod_FinalizeBlock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TripodServer is the server API for Tripod service.
// All implementations should embed UnimplementedTripodServer
// for forward compatibility.
type TripodServer interface {
	CheckTxn(context.Context, *TripodTxnRequest) (*Err, error)
	VerifyBlock(context.Context, *TripodBlockRequest) (*Err, error)
	StartBlock(context.Context, *TripodBlockRequest) (*Err, error)
	EndBlock(context.Context, *TripodBlockRequest) (*Err, error)
	FinalizeBlock(context.Context, *TripodBlockRequest) (*Err, error)
}

// UnimplementedTripodServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTripodServer struct{}

func (UnimplementedTripodServer) CheckTxn(context.Context, *TripodTxnRequest) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckTxn not implemented")
}
func (UnimplementedTripodServer) VerifyBlock(context.Context, *TripodBlockRequest) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyBlock not implemented")
}
func (UnimplementedTripodServer) StartBlock(context.Context, *TripodBlockRequest) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartBlock not implemented")
}
func (UnimplementedTripodServer) EndBlock(context.Context, *TripodBlockRequest) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EndBlock not implemented")
}
func (UnimplementedTripodServer) FinalizeBlock(context.Context, *TripodBlockRequest) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinalizeBlock not implemented")
}
func (UnimplementedTripodServer) testEmbeddedByValue() {}

// UnsafeTripodServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TripodServer will
// result in compilation errors.
type UnsafeTripodServer interface {
	mustEmbedUnimplementedTripodServer()
}

func RegisterTripodServer(s grpc.ServiceRegistrar, srv TripodServer) {
	// If the following call pancis, it indicates UnimplementedTripodServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Tripod_ServiceDesc, srv)
}

func _Tripod_CheckTxn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TripodTxnRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TripodServer).CheckTxn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tripod_CheckTxn_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TripodServer).CheckTxn(ctx, req.(*TripodTxnRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tripod_VerifyBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TripodBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TripodServer).VerifyBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tripod_VerifyBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TripodServer).VerifyBlock(ctx, req.(*TripodBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tripod_StartBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TripodBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TripodServer).StartBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tripod_StartBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TripodServer).StartBlock(ctx, req.(*TripodBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tripod_EndBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TripodBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TripodServer).EndBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tripod_EndBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TripodServer).EndBlock(ctx, req.(*TripodBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tripod_FinalizeBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TripodBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TripodServer).FinalizeBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tripod_FinalizeBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TripodServer).FinalizeBlock(ctx, req.(*TripodBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Tripod_ServiceDesc is the grpc.ServiceDesc for Tripod service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tripod_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Tripod",
	HandlerType: (*TripodServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckTxn",
			Handler:    _Tripod_CheckTxn_Handler,
		},
		{
			MethodName: "VerifyBlock",
			Handler:    _Tripod_VerifyBlock_Handler,
		},
		{
			MethodName: "StartBlock",
			Handler:    _Tripod_StartBlock_Handler,
		},
		{
			MethodName: "EndBlock",
			Handler:    _Tripod_EndBlock_Handler,
		},
		{
			MethodName: "FinalizeBlock",
			Handler:    _Tripod_FinalizeBlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tripod.proto",
}

const (
	Land_SetTripods_FullMethodName = "/Land/SetTripods"
)

// LandClient is the client API for Land service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LandClient interface {
	SetTripods(ctx context.Context, in *TripodsInfo, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type landClient struct {
	cc grpc.ClientConnInterface
}

func NewLandClient(cc grpc.ClientConnInterface) LandClient {
	return &landClient{cc}
}

func (c *landClient) SetTripods(ctx context.Context, in *TripodsInfo, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Land_SetTripods_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LandServer is the server API for Land service.
// All implementations should embed UnimplementedLandServer
// for forward compatibility.
type LandServer interface {
	SetTripods(context.Context, *TripodsInfo) (*emptypb.Empty, error)
}

// UnimplementedLandServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLandServer struct{}

func (UnimplementedLandServer) SetTripods(context.Context, *TripodsInfo) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetTripods not implemented")
}
func (UnimplementedLandServer) testEmbeddedByValue() {}

// UnsafeLandServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LandServer will
// result in compilation errors.
type UnsafeLandServer interface {
	mustEmbedUnimplementedLandServer()
}

func RegisterLandServer(s grpc.ServiceRegistrar, srv LandServer) {
	// If the following call pancis, it indicates UnimplementedLandServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Land_ServiceDesc, srv)
}

func _Land_SetTripods_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TripodsInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LandServer).SetTripods(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Land_SetTripods_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LandServer).SetTripods(ctx, req.(*TripodsInfo))
	}
	return interceptor(ctx, in, info, handler)
}

// Land_ServiceDesc is the grpc.ServiceDesc for Land service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Land_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Land",
	HandlerType: (*LandServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetTripods",
			Handler:    _Land_SetTripods_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tripod.proto",
}
