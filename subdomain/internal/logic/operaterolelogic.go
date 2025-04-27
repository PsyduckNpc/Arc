package logic

import (
	"Arc/db/work/dbs"
	"Arc/subdomain/internal/comm/commconst"
	"Arc/subdomain/internal/comm/utils"
	"Arc/subdomain/internal/comm/utils/xerr"
	"context"
	"github.com/pkg/errors"

	"Arc/subdomain/internal/svc"
	"Arc/subdomain/work/subdomains"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOperateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperateRoleLogic {
	return &OperateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OperateRoleLogic) OperateRole(req *subdomains.OptRoleAO) (*subdomains.OptRoleVO, error) {

	logx.Info("进入前台logic层,参数:", utils.MustMarshal(req))
	//校验参数...

	//转化数据微服务rpc调用入参
	optRoleDTO, err := utils.SliceCopy[*subdomains.Role, subdomains.Role](req.Roles)
	if err != nil {
		return nil, err
	}

	//先调用rpc数据服务 查询是否存在
	dtosJson := utils.MustMarshal(optRoleDTO)
	logx.Info("调用rpc数据微服务参数:", dtosJson)
	exec, err := l.svcCtx.CenterDataRpc.RpcServiceExec(l.ctx, &dbs.DataContentDTO{
		CenterDataApi: &dbs.CenterDataApi{ApiId: 2025042401, ApiParam: dtosJson, OpType: commconst.ROOT_QUERY}}) //apiid用于标识调用哪个sql
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "调用rpc数据微服务错误: %+v", err)
	}
	logx.Info("调用rpc数据微服务返回参数:", utils.MustMarshal(exec))
	//判断是否存在，存在的话抛出异常
	if exec.Total != 0 {
		return nil, errors.Wrapf(xerr.PERSON_ERROR(xerr.DATA_ERROR_CODE, "数据异常,数据已存在"), "数据异常,数据已存在")
	}

	//调用rpc数据服务,执行插入或修改
	logx.Info("调用rpc数据微服务参数:", dtosJson)
	exec, err = l.svcCtx.CenterDataRpc.RpcServiceExec(l.ctx, &dbs.DataContentDTO{
		CenterDataApi: &dbs.CenterDataApi{ApiId: 2025042401, ApiParam: dtosJson, OpType: commconst.ROOT_INSERT}}) //apiid用于标识调用哪个sql
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "调用rpc数据微服务错误: %+v", err)
	}
	logx.Info("调用rpc数据微服务返回参数:", utils.MustMarshal(exec))

	logx.Info("前台logic执行完毕")
	return &subdomains.OptRoleVO{OptNum: exec.Total}, nil
}
