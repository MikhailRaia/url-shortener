package proto

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type URLShortenRequest struct {
	URL string
}

type URLShortenResponse struct {
	Result string
}

type URLExpandRequest struct {
	ID string
}

type URLExpandResponse struct {
	Result string
}

type UserURLsResponse struct {
	URL []*URLData
}

type URLData struct {
	ShortURL    string
	OriginalURL string
}

// ShortenerServiceServer is the server API for ShortenerService service.
type ShortenerServiceServer interface {
	ShortenURL(context.Context, *URLShortenRequest) (*URLShortenResponse, error)
	ExpandURL(context.Context, *URLExpandRequest) (*URLExpandResponse, error)
	ListUserURLs(context.Context, *emptypb.Empty) (*UserURLsResponse, error)
}

// UnimplementedShortenerServiceServer can be embedded to have forward compatible implementations.
type UnimplementedShortenerServiceServer struct{}

func (*UnimplementedShortenerServiceServer) ShortenURL(context.Context, *URLShortenRequest) (*URLShortenResponse, error) {
	return nil, nil
}
func (*UnimplementedShortenerServiceServer) ExpandURL(context.Context, *URLExpandRequest) (*URLExpandResponse, error) {
	return nil, nil
}
func (*UnimplementedShortenerServiceServer) ListUserURLs(context.Context, *emptypb.Empty) (*UserURLsResponse, error) {
	return nil, nil
}

func RegisterShortenerServiceServer(s *grpc.Server, srv ShortenerServiceServer) {
	s.RegisterService(&_ShortenerServiceServiceDesc, srv)
}

func _ShortenerServiceShortenURLHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLShortenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ShortenURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/shortener.ShortenerService/ShortenURL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ShortenURL(ctx, req.(*URLShortenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerServiceExpandURLHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLExpandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ExpandURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/shortener.ShortenerService/ExpandURL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ExpandURL(ctx, req.(*URLExpandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerServiceListUserURLsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ListUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/shortener.ShortenerService/ListUserURLs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ListUserURLs(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _ShortenerServiceServiceDesc = grpc.ServiceDesc{
	ServiceName: "shortener.ShortenerService",
	HandlerType: (*ShortenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ShortenURL",
			Handler:    _ShortenerServiceShortenURLHandler,
		},
		{
			MethodName: "ExpandURL",
			Handler:    _ShortenerServiceExpandURLHandler,
		},
		{
			MethodName: "ListUserURLs",
			Handler:    _ShortenerServiceListUserURLsHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "shortener.proto",
}
