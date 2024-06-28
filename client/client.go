package grpccache

import (
	"context"
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/TryPerzh/grpc-cache/proto/grpcCache"
)

type CacheClient struct {
	cacheClient grpcCache.CacheServiceClient
	token       string
	ip          string
	port        string
	login       string
	password    string
	err         error
	conn        *grpc.ClientConn
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

	var err error = nil
	cc.conn, err = grpc.NewClient(cc.ip+":"+cc.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("client create : %v", err)
	}

	cc.cacheClient = grpcCache.NewCacheServiceClient(cc.conn)
	resp, err := cc.cacheClient.Login(context.Background(), &grpcCache.LoginRequest{Login: cc.login, Password: cc.password})
	if err != nil {
		cc.conn.Close()
		return fmt.Errorf("login error: %v", err)
	}
	if resp.Token == "" {
		cc.conn.Close()
		return fmt.Errorf("token receipt error")
	}
	cc.token = resp.Token
	return nil
}

func (cc *CacheClient) Close() error {

	return cc.conn.Close()
}

func (cc *CacheClient) Set(key string, value interface{}, duration time.Duration) {

	b, err := cbor.Marshal(value)
	if err != nil {
		cc.setError(fmt.Errorf("conversion to cbor: %v", err))
		return
	}

	request := &grpcCache.KeyValueDurationRequest{
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

func (cc *CacheClient) Add(key string, value interface{}, duration time.Duration) {

	b, err := cbor.Marshal(value)
	if err != nil {
		cc.setError(fmt.Errorf("conversion to cbor: %v", err))
		return
	}

	request := &grpcCache.KeyValueDurationRequest{
		Key:      key,
		Value:    b,
		Duration: durationpb.New(duration),
		Token:    cc.token,
	}

	_, err = cc.cacheClient.Add(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return
	}
}

func (cc *CacheClient) Replace(key string, value interface{}, duration time.Duration) {

	b, err := cbor.Marshal(value)
	if err != nil {
		cc.setError(fmt.Errorf("conversion to cbor: %v", err))
		return
	}

	request := &grpcCache.KeyValueDurationRequest{
		Key:      key,
		Value:    b,
		Duration: durationpb.New(duration),
		Token:    cc.token,
	}

	_, err = cc.cacheClient.Replace(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return
	}
}

func (cc *CacheClient) Get(key string) (interface{}, bool) {

	request := &grpcCache.KeyRequest{
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
		err = cbor.Unmarshal(resp.Value, &result)
		if err != nil {
			cc.setError(fmt.Errorf("conversion from cbor: %v", err))
			return nil, false
		}
		return result, resp.Found
	}

	return nil, resp.Found
}

func (cc *CacheClient) Delete(key string) {

	request := &grpcCache.KeyRequest{
		Key:   key,
		Token: cc.token,
	}

	_, err := cc.cacheClient.Delete(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return
	}
}

func (cc *CacheClient) Count(key string) int64 {

	request := &grpcCache.CountRequest{
		Token: cc.token,
	}

	resp, err := cc.cacheClient.Count(context.Background(), request)
	if err != nil {
		cc.setError(err)
		return -1
	}

	return resp.Count
}

func (cc *CacheClient) Error() error {
	err := cc.err
	cc.err = nil
	return err
}

func (cc *CacheClient) setError(err error) {
	cc.err = err
}
