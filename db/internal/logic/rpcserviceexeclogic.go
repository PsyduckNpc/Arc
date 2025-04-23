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
	"reflect"
	"regexp"
	"strings"

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
	// 1. 根据 ApiId 从表中取出对应的 SQL 语句
	apis, err := utils.QueryRowSlice[model.CenterDataApi](l.ctx, l.svcCtx, "select * from center_data_api where ApiId = ?", in.CenterDataApi.ApiId)
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "查询数据库API异常: %+v", err)
	}
	conds, err := utils.QueryRowSlice[model.CDataSrvRela](l.ctx, l.svcCtx, "select * from c_data_srv_rela where ApiId = ? order by SqlSort", in.CenterDataApi.ApiId)
	if err != nil {
		return nil, errors.Wrapf(xerr.DB_ERROR, "查询数据库RELA异常: %+v", err)
	}
	if apis == nil || len(apis) == 0 {
		return nil, errors.Wrapf(xerr.DB_ERROR, "未找到对应的Api记录, apiId: %d", in.CenterDataApi.ApiId)
	}
	if len(apis) > 1 {
		return nil, errors.Wrapf(xerr.DB_ERROR, "找到多个Api记录, apiId: %d", in.CenterDataApi.ApiId)
	}

	api := apis[0]
	logx.Info(api)
	logx.Info(conds)

	//入参参数存放到 inMap
	var inMap map[string]any
	if err := sonic.Unmarshal([]byte(in.CenterDataApi.ApiParam), &inMap); err != nil {
		return nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "入参数有误,不符合json结构,检查ApiParam:%s, 错误:%v", in.CenterDataApi.ApiParam, err)
	}

	//组装SQL
	sql, params, err := AssemblingSql(api, conds, inMap)
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
			numSql := AssemblingNumSql(*sql)

			//查询数量
			var num int64
			err = l.svcCtx.MySQL.QueryRowCtx(l.ctx, &num, *numSql, params...)
			if err != nil {
				return nil, errors.Wrapf(xerr.DB_ERROR, "执行数量查询异常")
			}
			if num == 0 {
				return &dbs.DataMapVO{Total: 0}, nil
			}

			//分页查询
			pageSql := AssemblingPageSql(*sql, page, &params)
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

// AssemblingSql
// 根据 CentDataApi 和 CDataSrvRela 组装SQL与参数
// 入参 api CentDataApi 数据
// 入参 conds CDataSrvRela数据
// 入参 root 上游服务的入参
// 出参 sql
// 出参 params
// 出参 e
func AssemblingSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMap map[string]any) (sql *string, params []any, e error) {

	if api.BeforeSql.Valid == false {
		return nil, nil, errors.Wrapf(xerr.DB_ERROR, "API配置错误: 未配置API前置SQL")
	}

	if conds == nil || len(conds) == 0 {
		return sql, nil, nil
	}

	var apiMap map[string]any
	if err := sonic.Unmarshal([]byte(api.ApiParam.String), &apiMap); err != nil {
		return nil, nil, errors.Wrapf(xerr.DB_ERROR, "中心数据参数有误,不符合json结构:%s, 错误:%v", api.ApiParam.String, err)
	}

	tarMap := make(map[string]any)
	var sqlWhere string
	for apiK := range apiMap {
		if val, exists := inMap[apiK]; exists && !utils.IsDefaultValue(val) { //todo 这里会导致默认值不能被处理,但是现在没有好办法
			tarMap[apiK] = val
		}
	}
	// 定义正则表达式：匹配 #{字母或数字}
	re := regexp.MustCompile(`#{([a-zA-Z0-9]+)}`)
	//检查条件是否需要添加
	for _, cond := range conds {
		addFlag := true
		// 查找所有匹配项
		matches := re.FindAllStringSubmatch(cond.SqlCondition.String, -1)
		for _, match := range matches {
			if tarMap[match[1]] != nil {
				addFlag = true
				break
			}
			addFlag = false
		}
		if addFlag == true {
			if strings.Trim(sqlWhere, " ") == "" {
				sqlWhere += " WHERE " + strings.Trim(cond.SqlCondition.String, " ") + " "
			} else {
				sqlWhere += " " + strings.Trim(cond.SqlLogic.String, " ") + " " + strings.Trim(cond.SqlCondition.String, " ") + " "
			}
		}
	}

	//完整SQL
	allSql := strings.TrimSpace(api.BeforeSql.String) + " " + sqlWhere + " " + strings.TrimSpace(api.AfterSql.String)
	submatchs := re.FindAllStringSubmatch(allSql, -1)

	//预编译参追加
	for _, sub := range submatchs {
		if tarMap[sub[1]] == nil {
			return nil, nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "SQL中存在无法匹配的候选项")
		}

		//如果是切片类型,预编译SQL处理
		if reflect.TypeOf(apiMap[sub[1]]).Kind() == reflect.Slice {
			split := strings.Split(tarMap[sub[1]].(string), ",")
			var quest string
			for _, item := range split {
				quest += "?,"
				params = append(params, item)
			}
			if len(quest) > 1 {
				quest = quest[:len(quest)-1]
			}
			allSql = strings.Replace(allSql, sub[0], quest, 1)

		} else { //非切片类型
			params = append(params, tarMap[sub[1]])
		}
	}
	allSql = re.ReplaceAllString(allSql, "?")
	sql = &allSql
	logx.Info("组装完成sql: %s", sql)
	return
}

// 组装统计数量sql
func AssemblingNumSql(sql string) *string {
	numSql := "select count(1) from (" + sql + ") adult"
	logx.Info("组装完成查询数量sql: %s", numSql)
	return &numSql
}

// 组装分页sql
func AssemblingPageSql(sql string, page model.Page, param *[]any) *string {
	numSql := "select * from (" + sql + ") adult LIMIT ?, ?"
	*param = append(*param, (page.CurPage-1)*page.PageSize)
	*param = append(*param, page.CurPage*page.PageSize)
	logx.Info("组装完成查询数量sql: %s", numSql)
	return &numSql
}
