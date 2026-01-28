package handler

import (
	"context"
	"errors"

	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/MikhailRaia/url-shortener/internal/proto"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerGRPCServer struct {
	proto.UnimplementedShortenerServiceServer
	urlService URLService
}

func NewShortenerGRPCServer(urlService URLService) *ShortenerGRPCServer {
	return &ShortenerGRPCServer{
		urlService: urlService,
	}
}

func (s *ShortenerGRPCServer) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	userID, _ := middleware.GetUserIDFromContext(ctx)

	shortURL, err := s.urlService.ShortenURLWithUser(ctx, req.Url, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			return &proto.URLShortenResponse{Result: shortURL}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to shorten URL: %v", err)
	}

	return &proto.URLShortenResponse{Result: shortURL}, nil
}

func (s *ShortenerGRPCServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	originalURL, err := s.urlService.GetOriginalURLWithDeletedStatus(ctx, req.Id)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			return nil, status.Error(codes.Unavailable, "url has been deleted")
		}
		return nil, status.Errorf(codes.Internal, "failed to expand URL: %v", err)
	}

	if originalURL == "" {
		return nil, status.Error(codes.NotFound, "url not found")
	}

	return &proto.URLExpandResponse{Result: originalURL}, nil
}

func (s *ShortenerGRPCServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	urls, err := s.urlService.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user URLs: %v", err)
	}

	resp := &proto.UserURLsResponse{
		Url: make([]*proto.URLData, 0, len(urls)),
	}

	for _, u := range urls {
		resp.Url = append(resp.Url, &proto.URLData{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		})
	}

	return resp, nil
}
