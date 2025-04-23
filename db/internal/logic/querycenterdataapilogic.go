package logic

import (
	"Arc/db/internal/comm/utils"
	"Arc/db/internal/model"
	"Arc/db/internal/svc"
	"Arc/db/work/dbs"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryCenterDataApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCenterDataApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCenterDataApiLogic {
	return &QueryCenterDataApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryCenterDataApiLogic) QueryCenterDataApi(in *dbs.CenterDataApi) (*dbs.CenterDataApiVO, error) {
	//查询
	slice, err := utils.QueryRowSlice[model.CenterDataApi](l.ctx, l.svcCtx, "select * from center_data_api where ApiId = ?", in.ApiId)
	if err != nil {
		return nil, err
	}

	proto, err := utils.SliceToProto[model.CenterDataApi, dbs.CenterDataApi](slice)
	if err != nil {
		return nil, err
	}

	return &dbs.CenterDataApiVO{ApiSlice: proto}, nil
}
