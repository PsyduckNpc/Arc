package logic

import (
	"Arc/db/internal/comm/utils"
	"Arc/db/internal/comm/utils/xerr"
	"Arc/db/internal/model"
	"github.com/bytedance/sonic"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"reflect"
	"regexp"
	"strings"
)

// AssemblSql
// 根据 CentDataApi 和 CDataSrvRela 组装SQL与参数
// 入参 api CentDataApi 数据
// 入参 conds CDataSrvRela数据
// 入参 root 上游服务的入参
// 出参 sql
// 出参 params
// 出参 e
func AssemblSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMap map[string]any) (sql *string, params []any, e error) {

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

// AssemblNumSql 组装统计数量sql
func AssemblNumSql(sql string) *string {
	numSql := "select count(1) from (" + sql + ") adult"
	logx.Info("组装完成查询数量sql: %s", numSql)
	return &numSql
}

// AssemblPageSql 组装分页sql
func AssemblPageSql(sql string, page model.Page, param *[]any) *string {
	numSql := "select * from (" + sql + ") adult LIMIT ?, ?"
	*param = append(*param, (page.CurPage-1)*page.PageSize)
	*param = append(*param, page.CurPage*page.PageSize)
	logx.Info("组装完成查询数量sql: %s", numSql)
	return &numSql
}

// AssemblCreatSql 组装 Insert SQL
func AssemblCreatSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sql string, params []any, err error) {
	//判断中心关联数据条数是否合规
	if len(conds) != 1 {
		return "", nil, errors.Wrapf(xerr.DB_ERROR, "%d的中心关联数据数量异常, 查询到的数量为:%d", api.ApiId, len(conds))
	}
	rela := conds[0]

	//步骤1 转换数据库映射json配置到map
	var fieldMap map[string]string
	if err = sonic.Unmarshal([]byte(rela.AttrMapping.String), &fieldMap); err != nil {
		return "", nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据属性映射字段配置有误,不符合json结构:%s, 错误:%v", rela.AttrMapping.String, err)
	}

	//步骤2 组成插入字段sql串
	//为保证有序，先根据数据库将字段映射形成切片
	insertSortList := make([]string, 0, len(fieldMap))
	for _, v := range fieldMap {
		insertSortList = append(insertSortList, v)
	}

	//步骤3 组成insert语句的表字段部分
	fieldSql := " ( " + strings.Join(insertSortList, ",") + " ) "

	//步骤4 组成insert语句的表内容部分, 根据inMaps入参的条数生成对应的插入数量语句
	params = make([]any, 0)
	paramSql := ""
	var copies []string
	if copies, err = utils.NCopies(len(fieldMap), "?"); err != nil {
		return "", nil, err
	}
	for i, inMap := range inMaps {
		paramSql += "("
		for _, v := range insertSortList {
			if inMap[v] != nil {
				params = append(params, inMap[v])
			} else {
				params = append(params, nil)
			}
		}
		paramSql += strings.Join(copies, ",")
		paramSql += ")"
		if i != len(inMaps)-1 {
			paramSql += ","
		}
	}

	//步骤5 组装完整SQL
	sql = "INSERT INTO " + rela.DataModelObhjName.String + fieldSql + " VALUES " + paramSql
	return
}

// AssemblUpdateSql 组装 Update SQL
func AssemblUpdateSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sql string, params []any, err error) {
	//判断中心关联数据条数是否合规
	if len(conds) != 1 {
		return "", nil, errors.Wrapf(xerr.DB_ERROR, "%d的中心关联数据数量异常, 查询到的数量为:%d", api.ApiId, len(conds))
	}
	rela := conds[0]

	//步骤1 转换数据库映射json配置到map
	var fieldMap map[string]string
	if err = sonic.Unmarshal([]byte(rela.AttrMapping.String), &fieldMap); err != nil {
		return "", nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据属性映射字段配置有误,不符合json结构:%s, 错误:%v", rela.AttrMapping.String, err)
	}

	//步骤2 组成修改Sql
	updateSql := ""
	updateList := make([]string, 0)
	params = make([]any, 0)
	for _, inMap := range inMaps {
		for inK, inV := range inMap {
			if fieldMap[inK] != "" {
				updateList = append(updateList, fieldMap[inK]+"=?")
				params = append(params, inV)
			}
		}

	}

	//步骤2 组成条件sql
	if rela.DataObjId.Valid == false || rela.DataObjId.String == "" {
		return "", nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据数据对象标识字段配置有误, 错误:%v", err)
	}

	whereSql := rela.DataObjId.String + "= ? "

	//步骤
	sql = "UPDATE " + rela.DataModelObhjName.String + " SET " + updateSql + " WHERE " + whereSql
	return

}
func AssemblDeleteSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sql string, params []any, e error) {
	return
}
func AssemblQuerySql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sql string, params []any, e error) {
	return
}

// 记录 入参 - 数据库字段
type ParamField struct {
	Param any
	Field string
}

// FindCommonKeysOptimized 返回公共键集合（自动选择遍历较小的 map 以优化性能）
func FindCommonKeysOptimized(inMap map[string]any, fieldMap map[string]string) map[string]ParamField {
	// 确保遍历较小的 map 以减少循环次数
	common := make(map[string]ParamField)
	if len(inMap) > len(fieldMap) { //FieldMap较小
		for k := range fieldMap {
			if _, exists := inMap[k]; exists {
				common[k] = ParamField{Field: fieldMap[k], Param: inMap[k]}
			}
		}
	} else {
		for k := range inMap {
			if _, exists := fieldMap[k]; exists {
				common[k] = ParamField{Field: fieldMap[k], Param: inMap[k]}
			}
		}
	}
	return common
}
