package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerServiceClient is the client API for ShortenerService service.
type ShortenerServiceClient interface {
	ShortenURL(ctx context.Context, in *URLShortenRequest, opts ...grpc.CallOption) (*URLShortenResponse, error)
	ExpandURL(ctx context.Context, in *URLExpandRequest, opts ...grpc.CallOption) (*URLExpandResponse, error)
	ListUserURLs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserURLsResponse, error)
}

type shortenerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerServiceClient(cc grpc.ClientConnInterface) ShortenerServiceClient {
	return &shortenerServiceClient{cc}
}

func (c *shortenerServiceClient) ShortenURL(ctx context.Context, in *URLShortenRequest, opts ...grpc.CallOption) (*URLShortenResponse, error) {
	out := new(URLShortenResponse)
	err := c.cc.Invoke(ctx, "/shortener.ShortenerService/ShortenURL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ExpandURL(ctx context.Context, in *URLExpandRequest, opts ...grpc.CallOption) (*URLExpandResponse, error) {
	out := new(URLExpandResponse)
	err := c.cc.Invoke(ctx, "/shortener.ShortenerService/ExpandURL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ListUserURLs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserURLsResponse, error) {
	out := new(UserURLsResponse)
	err := c.cc.Invoke(ctx, "/shortener.ShortenerService/ListUserURLs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServiceServer is the server API for ShortenerService service.
type ShortenerServiceServer interface {
	ShortenURL(context.Context, *URLShortenRequest) (*URLShortenResponse, error)
	ExpandURL(context.Context, *URLExpandRequest) (*URLExpandResponse, error)
	ListUserURLs(context.Context, *emptypb.Empty) (*UserURLsResponse, error)
	mustEmbedUnimplementedShortenerServiceServer()
}

// UnimplementedShortenerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedShortenerServiceServer struct {
}

func (UnimplementedShortenerServiceServer) ShortenURL(context.Context, *URLShortenRequest) (*URLShortenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShortenURL not implemented")
}
func (UnimplementedShortenerServiceServer) ExpandURL(context.Context, *URLExpandRequest) (*URLExpandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExpandURL not implemented")
}
func (UnimplementedShortenerServiceServer) ListUserURLs(context.Context, *emptypb.Empty) (*UserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUserURLs not implemented")
}
func (UnimplementedShortenerServiceServer) mustEmbedUnimplementedShortenerServiceServer() {}

func RegisterShortenerServiceServer(s grpc.ServiceRegistrar, srv ShortenerServiceServer) {
	s.RegisterService(&ShortenerServiceServiceDesc, srv)
}

var ShortenerServiceServiceDesc = grpc.ServiceDesc{
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
	Metadata: "proto/shortener.proto",
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
