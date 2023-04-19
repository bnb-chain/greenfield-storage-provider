package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorstypes "github.com/bnb-chain/greenfield-storage-provider/pkg/errors/types"
)

// StreamServerErrorInterceptor transfers an error to status error in stream server
func StreamServerErrorInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	err := handler(srv, ss)
	return toStatusError(err)
}

// StreamClientErrorInterceptor transfers an error to status error in stream client
func StreamClientErrorInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
	streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	stream, err := streamer(ctx, desc, cc, method, opts...)
	err = toServiceError(err)
	return stream, err
}

// UnaryServerErrorInterceptor transfers an error to status error in unary server
func UnaryServerErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	return resp, toStatusError(err)
}

// UnaryClientErrorInterceptor transfers an error to status error in unary client
func UnaryClientErrorInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	return toServiceError(err)
}

// toStatusError converts ServiceError to gRPC status error
func toStatusError(err error) error {
	if err == nil {
		return nil
	}
	myErr, ok := err.(*errorstypes.ServiceError)
	if ok {
		st := status.New(codes.Unknown, "service logic error")
		ds, e := st.WithDetails(&errorstypes.ServiceError{
			Code:   myErr.GetCode(),
			Reason: myErr.GetReason(),
		})
		if e != nil {
			return st.Err()
		}
		return ds.Err()
	}
	return err
}

// toServiceError converts gRPC status error to ServiceError
func toServiceError(err error) error {
	if err == nil {
		return nil
	}
	s := status.Convert(err)
	se := &errorstypes.ServiceError{}
	for _, d := range s.Details() {
		switch info := d.(type) {
		case *errorstypes.ServiceError:
			se.Code = info.GetCode()
			se.Reason = info.GetReason()
		default:
			return err
		}
	}
	return se
}
