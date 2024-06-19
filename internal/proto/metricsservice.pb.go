// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        v3.12.4
// source: internal/proto/metricsservice.proto

package proto

import (
	empty "github.com/golang/protobuf/ptypes/empty"
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

type Types int32

const (
	Types_GAUGE   Types = 0
	Types_COUNTER Types = 1
)

// Enum value maps for Types.
var (
	Types_name = map[int32]string{
		0: "GAUGE",
		1: "COUNTER",
	}
	Types_value = map[string]int32{
		"GAUGE":   0,
		"COUNTER": 1,
	}
)

func (x Types) Enum() *Types {
	p := new(Types)
	*p = x
	return p
}

func (x Types) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Types) Descriptor() protoreflect.EnumDescriptor {
	return file_internal_proto_metricsservice_proto_enumTypes[0].Descriptor()
}

func (Types) Type() protoreflect.EnumType {
	return &file_internal_proto_metricsservice_proto_enumTypes[0]
}

func (x Types) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Types.Descriptor instead.
func (Types) EnumDescriptor() ([]byte, []int) {
	return file_internal_proto_metricsservice_proto_rawDescGZIP(), []int{0}
}

type Metric struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//
	//	*Metric_Delta
	//	*Metric_Value
	Data isMetric_Data `protobuf_oneof:"data"`
	Id   string        `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	Type Types         `protobuf:"varint,4,opt,name=type,proto3,enum=metricssservice.Types" json:"type,omitempty"`
}

func (x *Metric) Reset() {
	*x = Metric{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metricsservice_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metric) ProtoMessage() {}

func (x *Metric) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metricsservice_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metric.ProtoReflect.Descriptor instead.
func (*Metric) Descriptor() ([]byte, []int) {
	return file_internal_proto_metricsservice_proto_rawDescGZIP(), []int{0}
}

func (m *Metric) GetData() isMetric_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *Metric) GetDelta() int64 {
	if x, ok := x.GetData().(*Metric_Delta); ok {
		return x.Delta
	}
	return 0
}

func (x *Metric) GetValue() float64 {
	if x, ok := x.GetData().(*Metric_Value); ok {
		return x.Value
	}
	return 0
}

func (x *Metric) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Metric) GetType() Types {
	if x != nil {
		return x.Type
	}
	return Types_GAUGE
}

type isMetric_Data interface {
	isMetric_Data()
}

type Metric_Delta struct {
	Delta int64 `protobuf:"varint,1,opt,name=delta,proto3,oneof"`
}

type Metric_Value struct {
	Value float64 `protobuf:"fixed64,2,opt,name=value,proto3,oneof"`
}

func (*Metric_Delta) isMetric_Data() {}

func (*Metric_Value) isMetric_Data() {}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metric *Metric `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
}

func (x *UpdateRequest) Reset() {
	*x = UpdateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metricsservice_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRequest) ProtoMessage() {}

func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metricsservice_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRequest.ProtoReflect.Descriptor instead.
func (*UpdateRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_metricsservice_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateRequest) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metric *Metric `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"`
}

func (x *UpdateResponse) Reset() {
	*x = UpdateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metricsservice_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateResponse) ProtoMessage() {}

func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metricsservice_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateResponse.ProtoReflect.Descriptor instead.
func (*UpdateResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_metricsservice_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateResponse) GetMetric() *Metric {
	if x != nil {
		return x.Metric
	}
	return nil
}

type UpdatesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metrics []*Metric `protobuf:"bytes,1,rep,name=metrics,proto3" json:"metrics,omitempty"`
}

func (x *UpdatesRequest) Reset() {
	*x = UpdatesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_metricsservice_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdatesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdatesRequest) ProtoMessage() {}

func (x *UpdatesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_metricsservice_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdatesRequest.ProtoReflect.Descriptor instead.
func (*UpdatesRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_metricsservice_proto_rawDescGZIP(), []int{3}
}

func (x *UpdatesRequest) GetMetrics() []*Metric {
	if x != nil {
		return x.Metrics
	}
	return nil
}

var File_internal_proto_metricsservice_proto protoreflect.FileDescriptor

var file_internal_proto_metricsservice_proto_rawDesc = []byte{
	0x0a, 0x23, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x7c, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x16, 0x0a,
	0x05, 0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x05,
	0x64, 0x65, 0x6c, 0x74, 0x61, 0x12, 0x16, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x01, 0x48, 0x00, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x2a, 0x0a,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x54, 0x79,
	0x70, 0x65, 0x73, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x22, 0x40, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x2f, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x06, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x22, 0x41, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x06,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x22, 0x43, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x31, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x73, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2a, 0x1f, 0x0a, 0x05, 0x54,
	0x79, 0x70, 0x65, 0x73, 0x12, 0x09, 0x0a, 0x05, 0x47, 0x41, 0x55, 0x47, 0x45, 0x10, 0x00, 0x12,
	0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x55, 0x4e, 0x54, 0x45, 0x52, 0x10, 0x01, 0x32, 0x98, 0x01, 0x0a,
	0x07, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x49, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x12, 0x1e, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x42, 0x0a, 0x07, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x12, 0x1f,
	0x2e, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x73, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x17, 0x5a, 0x15, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x73, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_proto_metricsservice_proto_rawDescOnce sync.Once
	file_internal_proto_metricsservice_proto_rawDescData = file_internal_proto_metricsservice_proto_rawDesc
)

func file_internal_proto_metricsservice_proto_rawDescGZIP() []byte {
	file_internal_proto_metricsservice_proto_rawDescOnce.Do(func() {
		file_internal_proto_metricsservice_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_metricsservice_proto_rawDescData)
	})
	return file_internal_proto_metricsservice_proto_rawDescData
}

var file_internal_proto_metricsservice_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_internal_proto_metricsservice_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_internal_proto_metricsservice_proto_goTypes = []interface{}{
	(Types)(0),             // 0: metricssservice.Types
	(*Metric)(nil),         // 1: metricssservice.Metric
	(*UpdateRequest)(nil),  // 2: metricssservice.UpdateRequest
	(*UpdateResponse)(nil), // 3: metricssservice.UpdateResponse
	(*UpdatesRequest)(nil), // 4: metricssservice.UpdatesRequest
	(*empty.Empty)(nil),    // 5: google.protobuf.Empty
}
var file_internal_proto_metricsservice_proto_depIdxs = []int32{
	0, // 0: metricssservice.Metric.type:type_name -> metricssservice.Types
	1, // 1: metricssservice.UpdateRequest.metric:type_name -> metricssservice.Metric
	1, // 2: metricssservice.UpdateResponse.metric:type_name -> metricssservice.Metric
	1, // 3: metricssservice.UpdatesRequest.metrics:type_name -> metricssservice.Metric
	2, // 4: metricssservice.Metrics.Update:input_type -> metricssservice.UpdateRequest
	4, // 5: metricssservice.Metrics.Updates:input_type -> metricssservice.UpdatesRequest
	3, // 6: metricssservice.Metrics.Update:output_type -> metricssservice.UpdateResponse
	5, // 7: metricssservice.Metrics.Updates:output_type -> google.protobuf.Empty
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_internal_proto_metricsservice_proto_init() }
func file_internal_proto_metricsservice_proto_init() {
	if File_internal_proto_metricsservice_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_metricsservice_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metric); i {
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
		file_internal_proto_metricsservice_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRequest); i {
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
		file_internal_proto_metricsservice_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateResponse); i {
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
		file_internal_proto_metricsservice_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdatesRequest); i {
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
	file_internal_proto_metricsservice_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Metric_Delta)(nil),
		(*Metric_Value)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_proto_metricsservice_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_proto_metricsservice_proto_goTypes,
		DependencyIndexes: file_internal_proto_metricsservice_proto_depIdxs,
		EnumInfos:         file_internal_proto_metricsservice_proto_enumTypes,
		MessageInfos:      file_internal_proto_metricsservice_proto_msgTypes,
	}.Build()
	File_internal_proto_metricsservice_proto = out.File
	file_internal_proto_metricsservice_proto_rawDesc = nil
	file_internal_proto_metricsservice_proto_goTypes = nil
	file_internal_proto_metricsservice_proto_depIdxs = nil
}