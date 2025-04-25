package logic

import (
	"Arc/db/internal/comm/utils"
	"Arc/db/internal/comm/utils/xerr"
	"Arc/db/internal/model"
	"Arc/db/internal/svc"
	"Arc/db/work/dbs"
	"context"
	"github.com/bytedance/sonic"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type RpcServiceExecLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRpcServiceExecLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RpcServiceExecLogic {
	return &RpcServiceExecLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RpcServiceExecLogic) RpcServiceExec(in *dbs.DataContentDTO) (*dbs.DataMapVO, error) {

	//步骤1 获取数据库数据
	api, conds, err := getCentData(l, in)
	if err != nil {
		return nil, err
	}

	//步骤3 判断操作类型 OpType 1查询 2写入
	if api.OpType.String == "1" {
		//查询操作
		return query(l, in, api, conds)
	} else if api.OpType.String == "2" {
		//写入操作
		return write(l, in, api, conds)
	}
	return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "未匹配到任何操作, 请检查数据库中心数据配置")
}

// getCentData 获取数据库中心数据
// 入参1 上下文
// 入参2 上游传参
// 出参1 中心数据
// 出参2 中心关联数据
func getCentData(l *RpcServiceExecLogic, in *dbs.DataContentDTO) (*model.CenterDataApi, *[]model.CDataSrvRela, error) {
	// 根据 ApiId 从表中取出对应的 SQL 语句
	apis, err := utils.QueryRowSlice[model.CenterDataApi](l.ctx, l.svcCtx, "select * from center_data_api where ApiId = ?", in.CenterDataApi.ApiId)
	if err != nil {
		return nil, nil, errors.Wrapf(xerr.DB_ERROR, "查询数据库API异常: %+v", err)
	}
	conds, err := utils.QueryRowSlice[model.CDataSrvRela](l.ctx, l.svcCtx, "select * from c_data_srv_rela where ApiId = ? order by SqlSort", in.CenterDataApi.ApiId)
	if err != nil {
		return nil, nil, errors.Wrapf(xerr.DB_ERROR, "查询数据库RELA异常: %+v", err)
	}
	if apis == nil || len(apis) == 0 {
		return nil, nil, errors.Wrapf(xerr.DB_ERROR, "未找到对应的Api记录, apiId: %d", in.CenterDataApi.ApiId)
	}
	if len(apis) > 1 {
		return nil, nil, errors.Wrapf(xerr.DB_ERROR, "找到多个Api记录, apiId: %d", in.CenterDataApi.ApiId)
	}

	api := apis[0]
	logx.Info(api)
	logx.Info(conds)

	return &api, &conds, nil
}

// query 查询费方法
// 入参1 上下文
// 入参2 上游参数
// 入参3 中心数据
// 入参4 中心关联数据
// 出参1 查询结果
func query(l *RpcServiceExecLogic, in *dbs.DataContentDTO, api *model.CenterDataApi, conds *[]model.CDataSrvRela) (*dbs.DataMapVO, error) {

	//入参参数存放到 inMap
	var inMap map[string]any
	if err := sonic.Unmarshal([]byte(in.CenterDataApi.ApiParam), &inMap); err != nil {
		return nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "入参数有误,不符合json结构,检查ApiParam:%s, 错误:%v", in.CenterDataApi.ApiParam, err)
	}

	//组装SQL
	sql, params, err := AssemblSql(*api, *conds, inMap)
	if err != nil {
		return nil, err
	}

	//检查是否分页 如是, 则需要查询总数和分页查询
	if inMap["page"] != nil {
		var page model.Page
		pageJson, err := sonic.Marshal(inMap["page"])
		if err := sonic.Unmarshal(pageJson, &page); err != nil {
			return nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "入参数有误,不符合json结构,检查ApiParam:%s, 错误:%v", inMap["page"], err)
		}
		//当前页码和页大小不为0视为符合分页查询条件
		if page.CurPage != 0 && page.PageSize != 0 {

			//组装统计和分页查询sql
			numSql := AssemblNumSql(*sql)

			//查询数量
			var num int64
			logx.Info("执行SQL:[%s] 参数:[%+v]", numSql, params)
			err = l.svcCtx.MySQL.QueryRowCtx(l.ctx, &num, *numSql, params...)
			if err != nil {
				return nil, errors.Wrapf(xerr.DB_ERROR, "执行数量查询异常")
			}
			if num == 0 {
				return &dbs.DataMapVO{Total: 0}, nil
			}

			//分页查询
			pageSql := AssemblPageSql(*sql, page, &params)
			execRes, err2 := utils.QueryRowDataMapVO(l.ctx, l.svcCtx, *pageSql, params...)
			if err2 != nil {
				return nil, err2
			}
			execRes.Total = num
			return execRes, nil
		}
	}

	//非分页查询
	execRes, err2 := utils.QueryRowDataMapVO(l.ctx, l.svcCtx, *sql, params...)
	if err2 != nil {
		return nil, err2
	}
	return execRes, nil
}

// 写入操作
// 入参1 上下文
// 入参2 上游参数
// 入参3 中心数据
// 入参4 中心关联数据
// 出参1 查询结果, DataMapVO中的Total记录错做结果数
func write(l *RpcServiceExecLogic, in *dbs.DataContentDTO, api *model.CenterDataApi, conds *[]model.CDataSrvRela) (*dbs.DataMapVO, error) {

	//入参参数存放到 inMap
	var inMap map[string]any
	if err := sonic.Unmarshal([]byte(in.CenterDataApi.ApiParam), &inMap); err != nil {
		return nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "入参数有误,不符合json结构,检查ApiParam:%s, 错误:%v", in.CenterDataApi.ApiParam, err)
	}

	DataListMap := inMap["DataList"].([]map[string]any)

	//1 判断操作类型
	var err error
	var rowsAffected int64
	var dmVO *dbs.DataMapVO
	switch inMap[model.ROOT_OPERATE_TYPE] {
	case model.ROOT_INSERT:
		rowsAffected, err = ExecCreatSql(l, *api, *conds, DataListMap)
	case model.ROOT_UPDATE:
		rowsAffected, err = ExecUpdateSql(l, *api, *conds, DataListMap)
	case model.ROOT_DELETE:
		rowsAffected, err = ExecDeleteSql(l, *api, *conds, DataListMap)
	case model.ROOT_QUERY: //相当于通过id查询  查询过应该同时返回数据集映射和总数量(适配配置的id非主键)
		dmVO, err = ExecSelectSql(l, *api, *conds, DataListMap)
	default:
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "未匹配到Root操作类型, 当前类型:[%s]", inMap[model.ROOT_OPERATE_TYPE])
	}

	if err != nil {
		return nil, err
	}
	//Total记录影响行数
	return &dbs.DataMapVO{Maps: dmVO.Maps, Total: rowsAffected}, nil
}
