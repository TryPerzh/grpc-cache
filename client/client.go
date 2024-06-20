package grpccache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"

	grpc_cache "github.com/TryPerzh/grpc-cache/proto/grpcCache"
)

type CacheClient struct {
	cacheClient grpc_cache.CacheServiceClient
	token       string
	ip          string
	port        string
	login       string
	password    string
	err         error
}

func New(ip string, port string, login string, password string) *CacheClient {
	return &CacheClient{
		ip:       ip,
		port:     port,
		login:    login,
		password: password,
	}
}

func (cc *CacheClient) Connect() error {

	conn, err := grpc.NewClient(cc.ip+":"+cc.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("client create : %v", err)
	}

	cc.cacheClient = grpc_cache.NewCacheServiceClient(conn)
	resp, err := cc.cacheClient.Login(context.Background(), &grpc_cache.LoginRequest{Login: cc.login, Password: cc.password})
	if err != nil {
		conn.Close()
		return fmt.Errorf("login error: %v", err)
	}
	if resp.Token == "" {
		conn.Close()
		return fmt.Errorf("token receipt error")
	}
	cc.token = resp.Token
	return nil
}

func (cc *CacheClient) Set(key string, value interface{}, duration time.Duration) {

	b, err := json.Marshal(value)
	if err != nil {
		cc.setError(fmt.Errorf("conversion to json: %v", err))
		return
	}

	request := &grpc_cache.SetRequest{
		Key:      key,
		Value:    b,
		Duration: durationpb.New(duration),
		Token:    cc.token,
	}

	_, err = cc.cacheClient.Set(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return
	}
}

func (cc *CacheClient) Get(key string) (interface{}, bool) {

	request := &grpc_cache.GetRequest{
		Key:   key,
		Token: cc.token,
	}

	resp, err := cc.cacheClient.Get(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return nil, false
	}

	if resp.Found {
		var result interface{}
		err = json.Unmarshal(resp.Value, &result)
		if err != nil {
			cc.setError(fmt.Errorf("conversion from json: %v", err))
			return nil, false
		}
		return result, resp.Found
	}

	return nil, resp.Found
}

func (cc *CacheClient) Delete(key string) {

	request := &grpc_cache.DeleteRequest{
		Key:   key,
		Token: cc.token,
	}

	_, err := cc.cacheClient.Delete(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return
	}
}

func (cc *CacheClient) Count(key string) int {

	request := &grpc_cache.CountRequest{
		Token: cc.token,
	}

	resp, err := cc.cacheClient.Count(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return -1
	}

	return int(resp.Count)
}

func (cc *CacheClient) Error() error {
	err := cc.err
	cc.err = nil
	return err
}

func (cc *CacheClient) setError(err error) {
	cc.err = err
}
