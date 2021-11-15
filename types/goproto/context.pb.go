// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: context.proto

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

type Context struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Caller []byte   `protobuf:"bytes,1,opt,name=caller,proto3" json:"caller,omitempty"`
	Params string   `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
	Events []*Event `protobuf:"bytes,3,rep,name=events,proto3" json:"events,omitempty"`
	Error  *Error   `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *Context) Reset() {
	*x = Context{}
	if protoimpl.UnsafeEnabled {
		mi := &file_context_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Context) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Context) ProtoMessage() {}

func (x *Context) ProtoReflect() protoreflect.Message {
	mi := &file_context_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Context.ProtoReflect.Descriptor instead.
func (*Context) Descriptor() ([]byte, []int) {
	return file_context_proto_rawDescGZIP(), []int{0}
}

func (x *Context) GetCaller() []byte {
	if x != nil {
		return x.Caller
	}
	return nil
}

func (x *Context) GetParams() string {
	if x != nil {
		return x.Params
	}
	return ""
}

func (x *Context) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

func (x *Context) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

var File_context_proto protoreflect.FileDescriptor

var file_context_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x0c, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x77, 0x0a,
	0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x61, 0x6c, 0x6c,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72,
	0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x1e, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x1c, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52,
	0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x2f, 0x67, 0x6f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_context_proto_rawDescOnce sync.Once
	file_context_proto_rawDescData = file_context_proto_rawDesc
)

func file_context_proto_rawDescGZIP() []byte {
	file_context_proto_rawDescOnce.Do(func() {
		file_context_proto_rawDescData = protoimpl.X.CompressGZIP(file_context_proto_rawDescData)
	})
	return file_context_proto_rawDescData
}

var file_context_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_context_proto_goTypes = []interface{}{
	(*Context)(nil), // 0: Context
	(*Event)(nil),   // 1: Event
	(*Error)(nil),   // 2: Error
}
var file_context_proto_depIdxs = []int32{
	1, // 0: Context.events:type_name -> Event
	2, // 1: Context.error:type_name -> Error
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_context_proto_init() }
func file_context_proto_init() {
	if File_context_proto != nil {
		return
	}
	file_result_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_context_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Context); i {
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
			RawDescriptor: file_context_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_context_proto_goTypes,
		DependencyIndexes: file_context_proto_depIdxs,
		MessageInfos:      file_context_proto_msgTypes,
	}.Build()
	File_context_proto = out.File
	file_context_proto_rawDesc = nil
	file_context_proto_goTypes = nil
	file_context_proto_depIdxs = nil
}
