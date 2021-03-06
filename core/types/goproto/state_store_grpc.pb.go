// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

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
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StateStoreClient is the client API for StateStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StateStoreClient interface {
	Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*ValueResponse, error)
	Set(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*Err, error)
	Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Err, error)
	Exist(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Bool, error)
	GetByBlockHash(ctx context.Context, in *KeyByHash, opts ...grpc.CallOption) (*ValueResponse, error)
	StartBlock(ctx context.Context, in *TxnHash, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SetCanRead(ctx context.Context, in *TxnHash, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Commit(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TxnHashResponse, error)
	Discard(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DiscardAll(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	NextTxn(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type stateStoreClient struct {
	cc grpc.ClientConnInterface
}

func NewStateStoreClient(cc grpc.ClientConnInterface) StateStoreClient {
	return &stateStoreClient{cc}
}

func (c *stateStoreClient) Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*ValueResponse, error) {
	out := new(ValueResponse)
	err := c.cc.Invoke(ctx, "/StateStore/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) Set(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*Err, error) {
	out := new(Err)
	err := c.cc.Invoke(ctx, "/StateStore/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Err, error) {
	out := new(Err)
	err := c.cc.Invoke(ctx, "/StateStore/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) Exist(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Bool, error) {
	out := new(Bool)
	err := c.cc.Invoke(ctx, "/StateStore/Exist", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) GetByBlockHash(ctx context.Context, in *KeyByHash, opts ...grpc.CallOption) (*ValueResponse, error) {
	out := new(ValueResponse)
	err := c.cc.Invoke(ctx, "/StateStore/GetByBlockHash", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) StartBlock(ctx context.Context, in *TxnHash, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/StateStore/StartBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) SetCanRead(ctx context.Context, in *TxnHash, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/StateStore/SetCanRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) Commit(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TxnHashResponse, error) {
	out := new(TxnHashResponse)
	err := c.cc.Invoke(ctx, "/StateStore/Commit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) Discard(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/StateStore/Discard", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) DiscardAll(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/StateStore/DiscardAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateStoreClient) NextTxn(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/StateStore/NextTxn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StateStoreServer is the server API for StateStore service.
// All implementations must embed UnimplementedStateStoreServer
// for forward compatibility
type StateStoreServer interface {
	Get(context.Context, *Key) (*ValueResponse, error)
	Set(context.Context, *KeyValue) (*Err, error)
	Delete(context.Context, *Key) (*Err, error)
	Exist(context.Context, *Key) (*Bool, error)
	GetByBlockHash(context.Context, *KeyByHash) (*ValueResponse, error)
	StartBlock(context.Context, *TxnHash) (*emptypb.Empty, error)
	SetCanRead(context.Context, *TxnHash) (*emptypb.Empty, error)
	Commit(context.Context, *emptypb.Empty) (*TxnHashResponse, error)
	Discard(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	DiscardAll(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	NextTxn(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedStateStoreServer()
}

// UnimplementedStateStoreServer must be embedded to have forward compatible implementations.
type UnimplementedStateStoreServer struct {
}

func (UnimplementedStateStoreServer) Get(context.Context, *Key) (*ValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedStateStoreServer) Set(context.Context, *KeyValue) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedStateStoreServer) Delete(context.Context, *Key) (*Err, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedStateStoreServer) Exist(context.Context, *Key) (*Bool, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exist not implemented")
}
func (UnimplementedStateStoreServer) GetByBlockHash(context.Context, *KeyByHash) (*ValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetByBlockHash not implemented")
}
func (UnimplementedStateStoreServer) StartBlock(context.Context, *TxnHash) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartBlock not implemented")
}
func (UnimplementedStateStoreServer) SetCanRead(context.Context, *TxnHash) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetCanRead not implemented")
}
func (UnimplementedStateStoreServer) Commit(context.Context, *emptypb.Empty) (*TxnHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Commit not implemented")
}
func (UnimplementedStateStoreServer) Discard(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Discard not implemented")
}
func (UnimplementedStateStoreServer) DiscardAll(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DiscardAll not implemented")
}
func (UnimplementedStateStoreServer) NextTxn(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NextTxn not implemented")
}
func (UnimplementedStateStoreServer) mustEmbedUnimplementedStateStoreServer() {}

// UnsafeStateStoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StateStoreServer will
// result in compilation errors.
type UnsafeStateStoreServer interface {
	mustEmbedUnimplementedStateStoreServer()
}

func RegisterStateStoreServer(s grpc.ServiceRegistrar, srv StateStoreServer) {
	s.RegisterService(&StateStore_ServiceDesc, srv)
}

func _StateStore_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).Get(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KeyValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).Set(ctx, req.(*KeyValue))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).Delete(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_Exist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).Exist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/Exist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).Exist(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_GetByBlockHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KeyByHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).GetByBlockHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/GetByBlockHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).GetByBlockHash(ctx, req.(*KeyByHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_StartBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxnHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).StartBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/StartBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).StartBlock(ctx, req.(*TxnHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_SetCanRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TxnHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).SetCanRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/SetCanRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).SetCanRead(ctx, req.(*TxnHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_Commit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).Commit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/Commit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).Commit(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_Discard_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).Discard(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/Discard",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).Discard(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_DiscardAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).DiscardAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/DiscardAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).DiscardAll(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _StateStore_NextTxn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateStoreServer).NextTxn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StateStore/NextTxn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateStoreServer).NextTxn(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// StateStore_ServiceDesc is the grpc.ServiceDesc for StateStore service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StateStore_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "StateStore",
	HandlerType: (*StateStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _StateStore_Get_Handler,
		},
		{
			MethodName: "Set",
			Handler:    _StateStore_Set_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _StateStore_Delete_Handler,
		},
		{
			MethodName: "Exist",
			Handler:    _StateStore_Exist_Handler,
		},
		{
			MethodName: "GetByBlockHash",
			Handler:    _StateStore_GetByBlockHash_Handler,
		},
		{
			MethodName: "StartBlock",
			Handler:    _StateStore_StartBlock_Handler,
		},
		{
			MethodName: "SetCanRead",
			Handler:    _StateStore_SetCanRead_Handler,
		},
		{
			MethodName: "Commit",
			Handler:    _StateStore_Commit_Handler,
		},
		{
			MethodName: "Discard",
			Handler:    _StateStore_Discard_Handler,
		},
		{
			MethodName: "DiscardAll",
			Handler:    _StateStore_DiscardAll_Handler,
		},
		{
			MethodName: "NextTxn",
			Handler:    _StateStore_NextTxn_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "state_store.proto",
}
