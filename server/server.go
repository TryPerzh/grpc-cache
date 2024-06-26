package grpccacheserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/fxamacker/cbor/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/TryPerzh/grpc-cache/cache"
	"github.com/TryPerzh/grpc-cache/proto/grpcCache"
	"github.com/TryPerzh/grpc-cache/server/tokens"
)

type Config struct {
	Port                   string
	DefaultCacheExpiration time.Duration
	CleanupCacheInterval   time.Duration
	TokensFile             string
}

type Server struct {
	grpcCache.CacheServiceServer
	Tokens *tokens.Tokens
	Cahce  *cache.Cache
	Port   string
}

func New() *Server {
	return &Server{}
}

func NewWithConfig(conf Config) *Server {
	var s Server

	if conf.Port == "" {
		s.Port = "8080"
	} else {
		s.Port = conf.Port
	}

	if conf.TokensFile == "" {
		s.Tokens = tokens.New()
	} else {
		s.Tokens = tokens.NewFromFile(conf.TokensFile)
	}

	var defaultCacheExpiration time.Duration
	var cleanupCacheInterval time.Duration

	if conf.DefaultCacheExpiration == time.Duration(0) {
		defaultCacheExpiration = 10 * time.Minute
	} else {
		defaultCacheExpiration = conf.DefaultCacheExpiration
	}

	if conf.CleanupCacheInterval == time.Duration(0) {
		cleanupCacheInterval = 5 * time.Minute
	} else {
		cleanupCacheInterval = conf.CleanupCacheInterval
	}

	s.Cahce = cache.New(defaultCacheExpiration, cleanupCacheInterval)

	return &s
}

func (s *Server) Login(ctx context.Context, req *grpcCache.LoginRequest) (*grpcCache.LoginResponse, error) {

	pass, f := s.Tokens.GetPassword(req.Login)
	if !f {
		return nil, fmt.Errorf("user with login %s not found", req.Login)
	}

	if pass != req.Password {
		return nil, fmt.Errorf("password is incorrect")
	}

	tok, err := s.Tokens.GetToken(req.Login)
	if err != nil {
		return nil, err
	}

	return &grpcCache.LoginResponse{
		Token: tok,
	}, nil
}

func (s *Server) Set(ctx context.Context, req *grpcCache.KeyValueDurationRequest) (*emptypb.Empty, error) {

	f, _ := s.Tokens.ValidToken(req.Token)

	if !f {
		return nil, fmt.Errorf("wrong token")
	}

	var value interface{}
	err := cbor.Unmarshal(req.Value, &value)
	if err != nil {
		return nil, err
	}

	s.Cahce.Set(req.Key, value, req.Duration.AsDuration())
	return &emptypb.Empty{}, err
}

func (s *Server) Add(ctx context.Context, req *grpcCache.KeyValueDurationRequest) (*emptypb.Empty, error) {

	f, _ := s.Tokens.ValidToken(req.Token)

	if !f {
		return nil, fmt.Errorf("wrong token")
	}

	var value interface{}
	err := cbor.Unmarshal(req.Value, &value)
	if err != nil {
		return nil, err
	}

	err = s.Cahce.Add(req.Key, value, req.Duration.AsDuration())
	return &emptypb.Empty{}, err
}

func (s *Server) Replace(ctx context.Context, req *grpcCache.KeyValueDurationRequest) (*emptypb.Empty, error) {

	f, _ := s.Tokens.ValidToken(req.Token)

	if !f {
		return nil, fmt.Errorf("wrong token")
	}

	var value interface{}
	err := cbor.Unmarshal(req.Value, &value)
	if err != nil {
		return nil, err
	}

	err = s.Cahce.Replace(req.Key, value, req.Duration.AsDuration())
	return &emptypb.Empty{}, err
}

func (s *Server) Get(ctx context.Context, req *grpcCache.KeyRequest) (*grpcCache.GetResponse, error) {

	f, _ := s.Tokens.ValidToken(req.Token)

	if !f {
		return nil, fmt.Errorf("wrong token")
	}

	item, f := s.Cahce.Get(req.Key)
	b, err := cbor.Marshal(item)
	if err != nil {
		fmt.Println("cbor error - ", err)
	}
	return &grpcCache.GetResponse{Value: b, Found: f}, nil
}

func (s *Server) Delete(ctx context.Context, req *grpcCache.KeyRequest) (*emptypb.Empty, error) {

	f, _ := s.Tokens.ValidToken(req.Token)

	if !f {
		return &emptypb.Empty{}, fmt.Errorf("wrong token")
	}

	s.Cahce.Delete(req.Key)
	return &emptypb.Empty{}, nil
}

func (s *Server) Count(ctx context.Context, req *grpcCache.CountRequest) (*grpcCache.CountResponse, error) {

	f, _ := s.Tokens.ValidToken(req.Token)

	if !f {
		return nil, fmt.Errorf("wrong token")
	}

	count := s.Cahce.Count()

	return &grpcCache.CountResponse{Count: int64(count)}, nil
}

func (s *Server) RunServer() {
	listener, err := net.Listen("tcp", ":"+s.Port)
	// listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	serv := grpc.NewServer()
	grpcCache.RegisterCacheServiceServer(serv, s)
	go func() {
		if err := serv.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}
