// Code generated by goctl. DO NOT EDIT.
// goctl 1.8.1
// Source: db.proto

package dbsclient

import (
	"context"

	"Arc/db/work/dbs"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	AnyMap          = dbs.AnyMap
	CenterDataApi   = dbs.CenterDataApi
	CenterDataApiVO = dbs.CenterDataApiVO
	DataContentDTO  = dbs.DataContentDTO
	DataMapVO       = dbs.DataMapVO
	Page            = dbs.Page

	Dbs interface {
		QueryCenterDataApi(ctx context.Context, in *CenterDataApi, opts ...grpc.CallOption) (*CenterDataApiVO, error)
		RpcServiceExec(ctx context.Context, in *DataContentDTO, opts ...grpc.CallOption) (*DataMapVO, error)
	}

	defaultDbs struct {
		cli zrpc.Client
	}
)

func NewDbs(cli zrpc.Client) Dbs {
	return &defaultDbs{
		cli: cli,
	}
}

func (m *defaultDbs) QueryCenterDataApi(ctx context.Context, in *CenterDataApi, opts ...grpc.CallOption) (*CenterDataApiVO, error) {
	client := dbs.NewDbsClient(m.cli.Conn())
	return client.QueryCenterDataApi(ctx, in, opts...)
}

func (m *defaultDbs) RpcServiceExec(ctx context.Context, in *DataContentDTO, opts ...grpc.CallOption) (*DataMapVO, error) {
	client := dbs.NewDbsClient(m.cli.Conn())
	return client.RpcServiceExec(ctx, in, opts...)
}
