// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1-devel
// 	protoc        v3.17.3
// source: funcs.proto

package goproto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ReadContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ParamsStr string `protobuf:"bytes,1,opt,name=params_str,json=paramsStr,proto3" json:"params_str,omitempty"`
	Response  []byte `protobuf:"bytes,2,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *ReadContext) Reset() {
	*x = ReadContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_funcs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReadContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReadContext) ProtoMessage() {}

func (x *ReadContext) ProtoReflect() protoreflect.Message {
	mi := &file_funcs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReadContext.ProtoReflect.Descriptor instead.
func (*ReadContext) Descriptor() ([]byte, []int) {
	return file_funcs_proto_rawDescGZIP(), []int{0}
}

func (x *ReadContext) GetParamsStr() string {
	if x != nil {
		return x.ParamsStr
	}
	return ""
}

func (x *ReadContext) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

type WriteContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReadContext *ReadContext `protobuf:"bytes,1,opt,name=read_context,json=readContext,proto3" json:"read_context,omitempty"`
	Block       *Block       `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	Txn         *SignedTxn   `protobuf:"bytes,3,opt,name=txn,proto3" json:"txn,omitempty"`
	Events      []*Event     `protobuf:"bytes,4,rep,name=events,proto3" json:"events,omitempty"`
	Error       *Error       `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
	LeiCost     uint64       `protobuf:"varint,6,opt,name=lei_cost,json=leiCost,proto3" json:"lei_cost,omitempty"`
}

func (x *WriteContext) Reset() {
	*x = WriteContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_funcs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WriteContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WriteContext) ProtoMessage() {}

func (x *WriteContext) ProtoReflect() protoreflect.Message {
	mi := &file_funcs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WriteContext.ProtoReflect.Descriptor instead.
func (*WriteContext) Descriptor() ([]byte, []int) {
	return file_funcs_proto_rawDescGZIP(), []int{1}
}

func (x *WriteContext) GetReadContext() *ReadContext {
	if x != nil {
		return x.ReadContext
	}
	return nil
}

func (x *WriteContext) GetBlock() *Block {
	if x != nil {
		return x.Block
	}
	return nil
}

func (x *WriteContext) GetTxn() *SignedTxn {
	if x != nil {
		return x.Txn
	}
	return nil
}

func (x *WriteContext) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

func (x *WriteContext) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *WriteContext) GetLeiCost() uint64 {
	if x != nil {
		return x.LeiCost
	}
	return 0
}

var File_funcs_proto protoreflect.FileDescriptor

var file_funcs_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x66, 0x75, 0x6e, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x09, 0x74, 0x78, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x10, 0x62, 0x61, 0x73, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x48, 0x0a, 0x0b, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x5f, 0x73,
	0x74, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73,
	0x53, 0x74, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0xd4, 0x01, 0x0a, 0x0c, 0x57, 0x72, 0x69, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x12, 0x2f, 0x0a, 0x0c, 0x72, 0x65, 0x61, 0x64, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x52, 0x0b, 0x72, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78,
	0x74, 0x12, 0x1c, 0x0a, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x06, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x12,
	0x1c, 0x0a, 0x03, 0x74, 0x78, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x53,
	0x69, 0x67, 0x6e, 0x65, 0x64, 0x54, 0x78, 0x6e, 0x52, 0x03, 0x74, 0x78, 0x6e, 0x12, 0x1e, 0x0a,
	0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x1c, 0x0a,
	0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x19, 0x0a, 0x08, 0x6c,
	0x65, 0x69, 0x5f, 0x63, 0x6f, 0x73, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x6c,
	0x65, 0x69, 0x43, 0x6f, 0x73, 0x74, 0x32, 0x27, 0x0a, 0x07, 0x57, 0x72, 0x69, 0x74, 0x69, 0x6e,
	0x67, 0x12, 0x1c, 0x0a, 0x05, 0x57, 0x72, 0x69, 0x74, 0x65, 0x12, 0x0d, 0x2e, 0x57, 0x72, 0x69,
	0x74, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x1a, 0x04, 0x2e, 0x45, 0x72, 0x72, 0x32,
	0x25, 0x0a, 0x07, 0x52, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x1a, 0x0a, 0x04, 0x52, 0x65,
	0x61, 0x64, 0x12, 0x0c, 0x2e, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x1a, 0x04, 0x2e, 0x45, 0x72, 0x72, 0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x2f, 0x67, 0x6f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_funcs_proto_rawDescOnce sync.Once
	file_funcs_proto_rawDescData = file_funcs_proto_rawDesc
)

func file_funcs_proto_rawDescGZIP() []byte {
	file_funcs_proto_rawDescOnce.Do(func() {
		file_funcs_proto_rawDescData = protoimpl.X.CompressGZIP(file_funcs_proto_rawDescData)
	})
	return file_funcs_proto_rawDescData
}

var file_funcs_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_funcs_proto_goTypes = []interface{}{
	(*ReadContext)(nil),  // 0: ReadContext
	(*WriteContext)(nil), // 1: WriteContext
	(*Block)(nil),        // 2: Block
	(*SignedTxn)(nil),    // 3: SignedTxn
	(*Event)(nil),        // 4: Event
	(*Error)(nil),        // 5: Error
	(*Err)(nil),          // 6: Err
}
var file_funcs_proto_depIdxs = []int32{
	0, // 0: WriteContext.read_context:type_name -> ReadContext
	2, // 1: WriteContext.block:type_name -> Block
	3, // 2: WriteContext.txn:type_name -> SignedTxn
	4, // 3: WriteContext.events:type_name -> Event
	5, // 4: WriteContext.error:type_name -> Error
	1, // 5: Writing.Write:input_type -> WriteContext
	0, // 6: Reading.Read:input_type -> ReadContext
	6, // 7: Writing.Write:output_type -> Err
	6, // 8: Reading.Read:output_type -> Err
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_funcs_proto_init() }
func file_funcs_proto_init() {
	if File_funcs_proto != nil {
		return
	}
	file_result_proto_init()
	file_block_proto_init()
	file_txn_proto_init()
	file_base_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_funcs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReadContext); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_funcs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WriteContext); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_funcs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_funcs_proto_goTypes,
		DependencyIndexes: file_funcs_proto_depIdxs,
		MessageInfos:      file_funcs_proto_msgTypes,
	}.Build()
	File_funcs_proto = out.File
	file_funcs_proto_rawDesc = nil
	file_funcs_proto_goTypes = nil
	file_funcs_proto_depIdxs = nil
}
