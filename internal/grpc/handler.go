package grpc

import (
	"context"
	"errors"

	"github.com/MikhailRaia/url-shortener/internal/grpc/proto"
	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/MikhailRaia/url-shortener/internal/service"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerServer implements the gRPC ShortenerService.
type ShortenerServer struct {
	proto.UnimplementedShortenerServiceServer
	urlService *service.URLService
}

// NewShortenerServer creates a new ShortenerServer.
func NewShortenerServer(urlService *service.URLService) *ShortenerServer {
	return &ShortenerServer{
		urlService: urlService,
	}
}

// ShortenURL shortens a given URL.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	if req.URL == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	userID, _ := middleware.GetUserIDFromContext(ctx)

	var shortURL string
	var err error

	if userID != "" {
		shortURL, err = s.urlService.ShortenURLWithUser(ctx, req.URL, userID)
	} else {
		shortURL, err = s.urlService.ShortenURL(ctx, req.URL)
	}

	if err != nil && !errors.Is(err, storage.ErrURLExists) {
		return nil, status.Errorf(codes.Internal, "failed to shorten URL: %v", err)
	}

	return &proto.URLShortenResponse{
		Result: shortURL,
	}, nil
}

// ExpandURL expands a short URL ID back to the original URL.
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	if req.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	originalURL, err := s.urlService.GetOriginalURLWithDeletedStatus(ctx, req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			return nil, status.Error(codes.NotFound, "URL is deleted")
		}
		return nil, status.Errorf(codes.NotFound, "URL not found: %v", err)
	}

	return &proto.URLExpandResponse{
		Result: originalURL,
	}, nil
}

// ListUserURLs returns all URLs created by the authenticated user.
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	urls, err := s.urlService.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user URLs: %v", err)
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no URLs found for user")
	}

	resp := &proto.UserURLsResponse{
		URL: make([]*proto.URLData, len(urls)),
	}

	for i, u := range urls {
		resp.URL[i] = &proto.URLData{
			ShortURL:    u.ShortURL,
			OriginalURL: u.OriginalURL,
		}
	}

	return resp, nil
}
