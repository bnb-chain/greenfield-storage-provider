// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: base/types/gfspserver/receive.proto

package gfspserver

import (
	context "context"
	fmt "fmt"
	gfsperrors "github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	gfsptask "github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type GfSpReplicatePieceRequest struct {
	ReceivePieceTask *gfsptask.GfSpReceivePieceTask `protobuf:"bytes,1,opt,name=receive_piece_task,json=receivePieceTask,proto3" json:"receive_piece_task,omitempty"`
	PieceData        []byte                         `protobuf:"bytes,2,opt,name=piece_data,json=pieceData,proto3" json:"piece_data,omitempty"`
}

func (m *GfSpReplicatePieceRequest) Reset()         { *m = GfSpReplicatePieceRequest{} }
func (m *GfSpReplicatePieceRequest) String() string { return proto.CompactTextString(m) }
func (*GfSpReplicatePieceRequest) ProtoMessage()    {}
func (*GfSpReplicatePieceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_74aba35d72479cf7, []int{0}
}
func (m *GfSpReplicatePieceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GfSpReplicatePieceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GfSpReplicatePieceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GfSpReplicatePieceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GfSpReplicatePieceRequest.Merge(m, src)
}
func (m *GfSpReplicatePieceRequest) XXX_Size() int {
	return m.Size()
}
func (m *GfSpReplicatePieceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GfSpReplicatePieceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GfSpReplicatePieceRequest proto.InternalMessageInfo

func (m *GfSpReplicatePieceRequest) GetReceivePieceTask() *gfsptask.GfSpReceivePieceTask {
	if m != nil {
		return m.ReceivePieceTask
	}
	return nil
}

func (m *GfSpReplicatePieceRequest) GetPieceData() []byte {
	if m != nil {
		return m.PieceData
	}
	return nil
}

type GfSpReplicatePieceResponse struct {
	Err *gfsperrors.GfSpError `protobuf:"bytes,1,opt,name=err,proto3" json:"err,omitempty"`
}

func (m *GfSpReplicatePieceResponse) Reset()         { *m = GfSpReplicatePieceResponse{} }
func (m *GfSpReplicatePieceResponse) String() string { return proto.CompactTextString(m) }
func (*GfSpReplicatePieceResponse) ProtoMessage()    {}
func (*GfSpReplicatePieceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_74aba35d72479cf7, []int{1}
}
func (m *GfSpReplicatePieceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GfSpReplicatePieceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GfSpReplicatePieceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GfSpReplicatePieceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GfSpReplicatePieceResponse.Merge(m, src)
}
func (m *GfSpReplicatePieceResponse) XXX_Size() int {
	return m.Size()
}
func (m *GfSpReplicatePieceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GfSpReplicatePieceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GfSpReplicatePieceResponse proto.InternalMessageInfo

func (m *GfSpReplicatePieceResponse) GetErr() *gfsperrors.GfSpError {
	if m != nil {
		return m.Err
	}
	return nil
}

type GfSpDoneReplicatePieceRequest struct {
	ReceivePieceTask *gfsptask.GfSpReceivePieceTask `protobuf:"bytes,1,opt,name=receive_piece_task,json=receivePieceTask,proto3" json:"receive_piece_task,omitempty"`
}

func (m *GfSpDoneReplicatePieceRequest) Reset()         { *m = GfSpDoneReplicatePieceRequest{} }
func (m *GfSpDoneReplicatePieceRequest) String() string { return proto.CompactTextString(m) }
func (*GfSpDoneReplicatePieceRequest) ProtoMessage()    {}
func (*GfSpDoneReplicatePieceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_74aba35d72479cf7, []int{2}
}
func (m *GfSpDoneReplicatePieceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GfSpDoneReplicatePieceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GfSpDoneReplicatePieceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GfSpDoneReplicatePieceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GfSpDoneReplicatePieceRequest.Merge(m, src)
}
func (m *GfSpDoneReplicatePieceRequest) XXX_Size() int {
	return m.Size()
}
func (m *GfSpDoneReplicatePieceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GfSpDoneReplicatePieceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GfSpDoneReplicatePieceRequest proto.InternalMessageInfo

func (m *GfSpDoneReplicatePieceRequest) GetReceivePieceTask() *gfsptask.GfSpReceivePieceTask {
	if m != nil {
		return m.ReceivePieceTask
	}
	return nil
}

type GfSpDoneReplicatePieceResponse struct {
	Err           *gfsperrors.GfSpError `protobuf:"bytes,1,opt,name=err,proto3" json:"err,omitempty"`
	IntegrityHash []byte                `protobuf:"bytes,2,opt,name=integrity_hash,json=integrityHash,proto3" json:"integrity_hash,omitempty"`
	Signature     []byte                `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *GfSpDoneReplicatePieceResponse) Reset()         { *m = GfSpDoneReplicatePieceResponse{} }
func (m *GfSpDoneReplicatePieceResponse) String() string { return proto.CompactTextString(m) }
func (*GfSpDoneReplicatePieceResponse) ProtoMessage()    {}
func (*GfSpDoneReplicatePieceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_74aba35d72479cf7, []int{3}
}
func (m *GfSpDoneReplicatePieceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GfSpDoneReplicatePieceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GfSpDoneReplicatePieceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GfSpDoneReplicatePieceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GfSpDoneReplicatePieceResponse.Merge(m, src)
}
func (m *GfSpDoneReplicatePieceResponse) XXX_Size() int {
	return m.Size()
}
func (m *GfSpDoneReplicatePieceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GfSpDoneReplicatePieceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GfSpDoneReplicatePieceResponse proto.InternalMessageInfo

func (m *GfSpDoneReplicatePieceResponse) GetErr() *gfsperrors.GfSpError {
	if m != nil {
		return m.Err
	}
	return nil
}

func (m *GfSpDoneReplicatePieceResponse) GetIntegrityHash() []byte {
	if m != nil {
		return m.IntegrityHash
	}
	return nil
}

func (m *GfSpDoneReplicatePieceResponse) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*GfSpReplicatePieceRequest)(nil), "base.types.gfspserver.GfSpReplicatePieceRequest")
	proto.RegisterType((*GfSpReplicatePieceResponse)(nil), "base.types.gfspserver.GfSpReplicatePieceResponse")
	proto.RegisterType((*GfSpDoneReplicatePieceRequest)(nil), "base.types.gfspserver.GfSpDoneReplicatePieceRequest")
	proto.RegisterType((*GfSpDoneReplicatePieceResponse)(nil), "base.types.gfspserver.GfSpDoneReplicatePieceResponse")
}

func init() {
	proto.RegisterFile("base/types/gfspserver/receive.proto", fileDescriptor_74aba35d72479cf7)
}

var fileDescriptor_74aba35d72479cf7 = []byte{
	// 429 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x53, 0x4f, 0x8b, 0xd3, 0x40,
	0x1c, 0xcd, 0xec, 0x82, 0xb0, 0xe3, 0x1f, 0x64, 0x40, 0xa9, 0xc1, 0x1d, 0x6a, 0x44, 0x58, 0x0f,
	0xcd, 0x68, 0xd5, 0x2f, 0x20, 0xeb, 0x9f, 0xe3, 0x92, 0x15, 0x04, 0x2f, 0x75, 0x92, 0xfe, 0x9a,
	0x0c, 0xbb, 0x66, 0xc6, 0xdf, 0x4c, 0x83, 0x8b, 0x1f, 0xc0, 0xab, 0xe0, 0x49, 0xfc, 0x42, 0x1e,
	0x7b, 0xf4, 0x28, 0xed, 0x17, 0x91, 0x4c, 0xda, 0x5a, 0x42, 0x2a, 0xd4, 0xc3, 0x5e, 0x92, 0xf0,
	0xf2, 0xde, 0xcb, 0xfb, 0xbd, 0xcc, 0x8f, 0xde, 0x4f, 0xa5, 0x05, 0xe1, 0x2e, 0x0c, 0x58, 0x91,
	0x4f, 0xac, 0xb1, 0x80, 0x15, 0xa0, 0x40, 0xc8, 0x40, 0x55, 0x10, 0x1b, 0xd4, 0x4e, 0xb3, 0x5b,
	0x35, 0x29, 0xf6, 0xa4, 0xf8, 0x2f, 0x29, 0xbc, 0xd7, 0xd2, 0x02, 0xa2, 0x46, 0x2b, 0xfc, 0xad,
	0x51, 0x86, 0xbc, 0x45, 0x71, 0xd2, 0x9e, 0x89, 0xfa, 0xd2, 0xbc, 0x8f, 0xbe, 0x11, 0x7a, 0xe7,
	0xd5, 0xe4, 0xd4, 0x24, 0x60, 0xce, 0x55, 0x26, 0x1d, 0x9c, 0x28, 0xc8, 0x20, 0x81, 0x8f, 0x53,
	0xb0, 0x8e, 0xbd, 0xa5, 0x6c, 0x19, 0x64, 0x64, 0x6a, 0x7c, 0x54, 0x2b, 0x7b, 0xa4, 0x4f, 0x8e,
	0xae, 0x0e, 0x1f, 0xc6, 0xad, 0x50, 0xde, 0xb5, 0xf1, 0xf2, 0x12, 0xef, 0xf4, 0x46, 0xda, 0xb3,
	0xe4, 0x26, 0xb6, 0x10, 0x76, 0x48, 0x69, 0x63, 0x38, 0x96, 0x4e, 0xf6, 0xf6, 0xfa, 0xe4, 0xe8,
	0x5a, 0x72, 0xe0, 0x91, 0x63, 0xe9, 0x64, 0x74, 0x42, 0xc3, 0xae, 0x50, 0xd6, 0xe8, 0xd2, 0x02,
	0x1b, 0xd2, 0x7d, 0x40, 0x5c, 0xc6, 0xe8, 0xb7, 0x63, 0x34, 0x25, 0xf8, 0x20, 0x2f, 0xea, 0xc7,
	0xa4, 0x26, 0x47, 0x9f, 0xe8, 0x61, 0x8d, 0x1c, 0xeb, 0x12, 0x2e, 0x77, 0xd4, 0xe8, 0x3b, 0xa1,
	0x7c, 0xdb, 0xa7, 0xff, 0x7f, 0x20, 0xf6, 0x80, 0xde, 0x50, 0xa5, 0x83, 0x1c, 0x95, 0xbb, 0x18,
	0x15, 0xd2, 0x16, 0xcb, 0x16, 0xaf, 0xaf, 0xd1, 0xd7, 0xd2, 0x16, 0xec, 0x2e, 0x3d, 0xb0, 0x2a,
	0x2f, 0xa5, 0x9b, 0x22, 0xf4, 0xf6, 0x9b, 0x9e, 0xd7, 0xc0, 0xf0, 0xc7, 0x1e, 0x65, 0x1b, 0x63,
	0x9c, 0x02, 0x56, 0x2a, 0x03, 0xf6, 0x79, 0x85, 0x6e, 0xa6, 0x65, 0x8f, 0xe2, 0xce, 0x53, 0x18,
	0x6f, 0x3d, 0x3e, 0xe1, 0xe3, 0x1d, 0x14, 0x4d, 0x15, 0x51, 0xc0, 0xbe, 0x10, 0x7a, 0xbb, 0xbb,
	0x2f, 0xf6, 0xf4, 0x1f, 0x7e, 0x5b, 0xff, 0x6c, 0xf8, 0x6c, 0x47, 0xd5, 0x2a, 0xc9, 0xf3, 0xf7,
	0x3f, 0xe7, 0x9c, 0xcc, 0xe6, 0x9c, 0xfc, 0x9e, 0x73, 0xf2, 0x75, 0xc1, 0x83, 0xd9, 0x82, 0x07,
	0xbf, 0x16, 0x3c, 0x78, 0xf7, 0x32, 0x57, 0xae, 0x98, 0xa6, 0x71, 0xa6, 0x3f, 0x88, 0xb4, 0x4c,
	0x07, 0x59, 0x21, 0x55, 0x29, 0x72, 0x04, 0x28, 0x27, 0x0a, 0xce, 0xc7, 0x03, 0xeb, 0x34, 0xca,
	0x1c, 0x06, 0x06, 0x75, 0xa5, 0xc6, 0x80, 0xa2, 0x73, 0xcb, 0xd3, 0x2b, 0x7e, 0x09, 0x9f, 0xfc,
	0x09, 0x00, 0x00, 0xff, 0xff, 0xf4, 0x87, 0x1d, 0x18, 0x05, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GfSpReceiveServiceClient is the client API for GfSpReceiveService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GfSpReceiveServiceClient interface {
	GfSpReplicatePiece(ctx context.Context, in *GfSpReplicatePieceRequest, opts ...grpc.CallOption) (*GfSpReplicatePieceResponse, error)
	GfSpDoneReplicatePiece(ctx context.Context, in *GfSpDoneReplicatePieceRequest, opts ...grpc.CallOption) (*GfSpDoneReplicatePieceResponse, error)
}

type gfSpReceiveServiceClient struct {
	cc grpc1.ClientConn
}

func NewGfSpReceiveServiceClient(cc grpc1.ClientConn) GfSpReceiveServiceClient {
	return &gfSpReceiveServiceClient{cc}
}

func (c *gfSpReceiveServiceClient) GfSpReplicatePiece(ctx context.Context, in *GfSpReplicatePieceRequest, opts ...grpc.CallOption) (*GfSpReplicatePieceResponse, error) {
	out := new(GfSpReplicatePieceResponse)
	err := c.cc.Invoke(ctx, "/base.types.gfspserver.GfSpReceiveService/GfSpReplicatePiece", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gfSpReceiveServiceClient) GfSpDoneReplicatePiece(ctx context.Context, in *GfSpDoneReplicatePieceRequest, opts ...grpc.CallOption) (*GfSpDoneReplicatePieceResponse, error) {
	out := new(GfSpDoneReplicatePieceResponse)
	err := c.cc.Invoke(ctx, "/base.types.gfspserver.GfSpReceiveService/GfSpDoneReplicatePiece", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GfSpReceiveServiceServer is the server API for GfSpReceiveService service.
type GfSpReceiveServiceServer interface {
	GfSpReplicatePiece(context.Context, *GfSpReplicatePieceRequest) (*GfSpReplicatePieceResponse, error)
	GfSpDoneReplicatePiece(context.Context, *GfSpDoneReplicatePieceRequest) (*GfSpDoneReplicatePieceResponse, error)
}

// UnimplementedGfSpReceiveServiceServer can be embedded to have forward compatible implementations.
type UnimplementedGfSpReceiveServiceServer struct {
}

func (*UnimplementedGfSpReceiveServiceServer) GfSpReplicatePiece(ctx context.Context, req *GfSpReplicatePieceRequest) (*GfSpReplicatePieceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GfSpReplicatePiece not implemented")
}
func (*UnimplementedGfSpReceiveServiceServer) GfSpDoneReplicatePiece(ctx context.Context, req *GfSpDoneReplicatePieceRequest) (*GfSpDoneReplicatePieceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GfSpDoneReplicatePiece not implemented")
}

func RegisterGfSpReceiveServiceServer(s grpc1.Server, srv GfSpReceiveServiceServer) {
	s.RegisterService(&_GfSpReceiveService_serviceDesc, srv)
}

func _GfSpReceiveService_GfSpReplicatePiece_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GfSpReplicatePieceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GfSpReceiveServiceServer).GfSpReplicatePiece(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/base.types.gfspserver.GfSpReceiveService/GfSpReplicatePiece",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GfSpReceiveServiceServer).GfSpReplicatePiece(ctx, req.(*GfSpReplicatePieceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GfSpReceiveService_GfSpDoneReplicatePiece_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GfSpDoneReplicatePieceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GfSpReceiveServiceServer).GfSpDoneReplicatePiece(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/base.types.gfspserver.GfSpReceiveService/GfSpDoneReplicatePiece",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GfSpReceiveServiceServer).GfSpDoneReplicatePiece(ctx, req.(*GfSpDoneReplicatePieceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _GfSpReceiveService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "base.types.gfspserver.GfSpReceiveService",
	HandlerType: (*GfSpReceiveServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GfSpReplicatePiece",
			Handler:    _GfSpReceiveService_GfSpReplicatePiece_Handler,
		},
		{
			MethodName: "GfSpDoneReplicatePiece",
			Handler:    _GfSpReceiveService_GfSpDoneReplicatePiece_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "base/types/gfspserver/receive.proto",
}

func (m *GfSpReplicatePieceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GfSpReplicatePieceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GfSpReplicatePieceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PieceData) > 0 {
		i -= len(m.PieceData)
		copy(dAtA[i:], m.PieceData)
		i = encodeVarintReceive(dAtA, i, uint64(len(m.PieceData)))
		i--
		dAtA[i] = 0x12
	}
	if m.ReceivePieceTask != nil {
		{
			size, err := m.ReceivePieceTask.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintReceive(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GfSpReplicatePieceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GfSpReplicatePieceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GfSpReplicatePieceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Err != nil {
		{
			size, err := m.Err.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintReceive(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GfSpDoneReplicatePieceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GfSpDoneReplicatePieceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GfSpDoneReplicatePieceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ReceivePieceTask != nil {
		{
			size, err := m.ReceivePieceTask.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintReceive(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GfSpDoneReplicatePieceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GfSpDoneReplicatePieceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GfSpDoneReplicatePieceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintReceive(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.IntegrityHash) > 0 {
		i -= len(m.IntegrityHash)
		copy(dAtA[i:], m.IntegrityHash)
		i = encodeVarintReceive(dAtA, i, uint64(len(m.IntegrityHash)))
		i--
		dAtA[i] = 0x12
	}
	if m.Err != nil {
		{
			size, err := m.Err.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintReceive(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintReceive(dAtA []byte, offset int, v uint64) int {
	offset -= sovReceive(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GfSpReplicatePieceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ReceivePieceTask != nil {
		l = m.ReceivePieceTask.Size()
		n += 1 + l + sovReceive(uint64(l))
	}
	l = len(m.PieceData)
	if l > 0 {
		n += 1 + l + sovReceive(uint64(l))
	}
	return n
}

func (m *GfSpReplicatePieceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Err != nil {
		l = m.Err.Size()
		n += 1 + l + sovReceive(uint64(l))
	}
	return n
}

func (m *GfSpDoneReplicatePieceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ReceivePieceTask != nil {
		l = m.ReceivePieceTask.Size()
		n += 1 + l + sovReceive(uint64(l))
	}
	return n
}

func (m *GfSpDoneReplicatePieceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Err != nil {
		l = m.Err.Size()
		n += 1 + l + sovReceive(uint64(l))
	}
	l = len(m.IntegrityHash)
	if l > 0 {
		n += 1 + l + sovReceive(uint64(l))
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovReceive(uint64(l))
	}
	return n
}

func sovReceive(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozReceive(x uint64) (n int) {
	return sovReceive(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GfSpReplicatePieceRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReceive
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GfSpReplicatePieceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GfSpReplicatePieceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReceivePieceTask", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ReceivePieceTask == nil {
				m.ReceivePieceTask = &gfsptask.GfSpReceivePieceTask{}
			}
			if err := m.ReceivePieceTask.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PieceData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PieceData = append(m.PieceData[:0], dAtA[iNdEx:postIndex]...)
			if m.PieceData == nil {
				m.PieceData = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipReceive(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthReceive
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GfSpReplicatePieceResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReceive
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GfSpReplicatePieceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GfSpReplicatePieceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Err", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Err == nil {
				m.Err = &gfsperrors.GfSpError{}
			}
			if err := m.Err.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipReceive(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthReceive
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GfSpDoneReplicatePieceRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReceive
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GfSpDoneReplicatePieceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GfSpDoneReplicatePieceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReceivePieceTask", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ReceivePieceTask == nil {
				m.ReceivePieceTask = &gfsptask.GfSpReceivePieceTask{}
			}
			if err := m.ReceivePieceTask.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipReceive(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthReceive
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GfSpDoneReplicatePieceResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowReceive
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GfSpDoneReplicatePieceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GfSpDoneReplicatePieceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Err", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Err == nil {
				m.Err = &gfsperrors.GfSpError{}
			}
			if err := m.Err.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IntegrityHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IntegrityHash = append(m.IntegrityHash[:0], dAtA[iNdEx:postIndex]...)
			if m.IntegrityHash == nil {
				m.IntegrityHash = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthReceive
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthReceive
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipReceive(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthReceive
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipReceive(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowReceive
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowReceive
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthReceive
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupReceive
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthReceive
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthReceive        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowReceive          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupReceive = fmt.Errorf("proto: unexpected end of group")
)