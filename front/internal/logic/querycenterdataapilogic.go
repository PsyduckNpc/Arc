package logic

import (
	"Arc/db/work/dbs"
	"Arc/front/internal/comm/utils"
	"Arc/front/internal/svc"
	"Arc/front/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryCenterDataApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCenterDataApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCenterDataApiLogic {
	return &QueryCenterDataApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryCenterDataApiLogic) QueryCenterDataApi(req *types.CenterDataApi) (resp *[]types.CenterDataApi, err error) {
	api, err := l.svcCtx.CenterDataRpc.QueryCenterDataApi(l.ctx, &dbs.CenterDataApi{
		AfterSql: req.AfterSql,
		SqlParam: req.SqlParam,
		ApiId:    req.ApiId,
	})
	if err != nil {
		return nil, err
	}

	slice, err := utils.SliceCopy[*dbs.CenterDataApi, types.CenterDataApi](api.ApiSlice)
	return &slice, nil
}
