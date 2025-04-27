package logic

import (
	"Arc/front/internal/comm/utils"
	"Arc/front/internal/comm/utils/xerr"
	"Arc/subdomain/work/subdomains"
	"context"
	"github.com/pkg/errors"

	"Arc/front/internal/svc"
	"Arc/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OptRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOptRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OptRoleLogic {
	return &OptRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OptRoleLogic) OptRole(req *types.OptRoleAO) error {
	logx.Info("进入前台logic层,参数:", utils.MustMarshal(req))
	//校验参数...

	//切片拷贝
	optRoleDTOs, err := utils.SliceCopy[types.Role, *subdomains.Role](req.Roles)
	if err != nil {
		return err
	}

	//调用中台服务
	logx.Info("调用rpc中台服务参数:", utils.MustMarshal(optRoleDTOs))
	optRoleVO, err := l.svcCtx.SubdomainRpc.OperateRole(l.ctx, &subdomains.OptRoleAO{Roles: optRoleDTOs})
	if err != nil {
		return errors.Wrapf(xerr.DB_ERROR, "调用rpc数据微服务错误: %+v", err)
	}
	logx.Info("调用rpc中台服务返回参数:", utils.MustMarshal(optRoleVO))

	//返回内容
	logx.Info("前台logic执行完毕")
	return nil
}
