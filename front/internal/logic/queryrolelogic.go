package logic

import (
	"Arc/db/work/dbs"
	"Arc/front/internal/comm/utils"
	"Arc/front/internal/comm/utils/xerr"
	"context"
	"github.com/pkg/errors"

	"Arc/front/internal/svc"
	"Arc/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRoleLogic {
	return &QueryRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryRoleLogic) QueryRole(req *types.QryRoleAO) (resp *types.QryRoleVO, err error) {
	logx.Info("进入前台logic层,参数:", utils.MustMarshal(req))
	//校验参数

	//转化数据微服务rpc调用入参
	marshal := utils.MustMarshal(req)

	//调用rpc数据服务
	logx.Info("调用rpc数据微服务参数:", marshal)
	exec, err := l.svcCtx.CenterDataRpc.RpcServiceExec(l.ctx, &dbs.DataContentDTO{
		CenterDataApi: &dbs.CenterDataApi{ApiId: 2023051201, ApiParam: marshal}, Page: nil}) //apiid用于标识调用哪个sql
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "调用rpc数据微服务错误: %+v", err)
	}
	logx.Info("调用rpc数据微服务返回参数:", utils.MustMarshal(exec))

	//数据微服务反参转换
	slice, allNum, err := utils.ProtoToSlice[types.Role](exec)
	if err != nil {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "Proto结构体转Role错误: %+v", err)
	}
	logx.Info("前台logic执行完毕,返回参数:", utils.MustMarshal(slice))
	return &types.QryRoleVO{
		List: slice,
		Page: &types.Page{AllNum: allNum},
	}, nil
}
