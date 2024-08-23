// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v4.24.4
// source: protobuf/service.proto

package protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	KafkaStreamer_Subscribe_FullMethodName = "/protobuf.KafkaStreamer/Subscribe"
)

// KafkaStreamerClient is the client API for KafkaStreamer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KafkaStreamerClient interface {
	Subscribe(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[SubscribeRequest, Message], error)
}

type kafkaStreamerClient struct {
	cc grpc.ClientConnInterface
}

func NewKafkaStreamerClient(cc grpc.ClientConnInterface) KafkaStreamerClient {
	return &kafkaStreamerClient{cc}
}

func (c *kafkaStreamerClient) Subscribe(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[SubscribeRequest, Message], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &KafkaStreamer_ServiceDesc.Streams[0], KafkaStreamer_Subscribe_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[SubscribeRequest, Message]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type KafkaStreamer_SubscribeClient = grpc.BidiStreamingClient[SubscribeRequest, Message]

// KafkaStreamerServer is the server API for KafkaStreamer service.
// All implementations must embed UnimplementedKafkaStreamerServer
// for forward compatibility.
type KafkaStreamerServer interface {
	Subscribe(grpc.BidiStreamingServer[SubscribeRequest, Message]) error
	mustEmbedUnimplementedKafkaStreamerServer()
}

// UnimplementedKafkaStreamerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedKafkaStreamerServer struct{}

func (UnimplementedKafkaStreamerServer) Subscribe(grpc.BidiStreamingServer[SubscribeRequest, Message]) error {
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedKafkaStreamerServer) mustEmbedUnimplementedKafkaStreamerServer() {}
func (UnimplementedKafkaStreamerServer) testEmbeddedByValue()                       {}

// UnsafeKafkaStreamerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KafkaStreamerServer will
// result in compilation errors.
type UnsafeKafkaStreamerServer interface {
	mustEmbedUnimplementedKafkaStreamerServer()
}

func RegisterKafkaStreamerServer(s grpc.ServiceRegistrar, srv KafkaStreamerServer) {
	// If the following call pancis, it indicates UnimplementedKafkaStreamerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&KafkaStreamer_ServiceDesc, srv)
}

func _KafkaStreamer_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(KafkaStreamerServer).Subscribe(&grpc.GenericServerStream[SubscribeRequest, Message]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type KafkaStreamer_SubscribeServer = grpc.BidiStreamingServer[SubscribeRequest, Message]

// KafkaStreamer_ServiceDesc is the grpc.ServiceDesc for KafkaStreamer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KafkaStreamer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protobuf.KafkaStreamer",
	HandlerType: (*KafkaStreamerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _KafkaStreamer_Subscribe_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protobuf/service.proto",
}
