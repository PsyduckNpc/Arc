package logic

import (
	"Arc/db/internal/comm/utils"
	"Arc/db/internal/comm/utils/xerr"
	"Arc/db/internal/model"
	"Arc/db/work/dbs"
	"github.com/bytedance/sonic"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"reflect"
	"regexp"
	"strings"
)

// IN 列表数量 <= eq_range_index_dive_limit	✅ 走索引	    优化器通过索引潜水精确计算成本，认为索引更快。
// IN 列表数量 >  eq_range_index_dive_limit	❌ 可能不走索引	改用统计信息估算成本，若估算结果认为全表扫描更快，则放弃索引。
// IN 覆盖大部分主键值							❌ 可能不走索引	全表扫描成本低于多次索引随机访问。
// 主键索引碎片化严重							❌ 可能不走索引	索引访问效率下降，优化器选择全表扫描。
const (
	//用来控制批次大小，防止出现超过数据库限制或索引限制
	insertBitchSize int = 10
	deleteBitchSize int = 10
	selectBitchSize int = 10
)

type SqlParam struct {
	sql    string
	params []any
}

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

// ExecSingleSql 单次执行语句
func ExecSingleSql(l *RpcServiceExecLogic, sqlParam SqlParam) (rowsAffected int64, err error) {
	return ExecBatchSql(l, []SqlParam{sqlParam})
}

// ExecBatchSql 批量执行语句
func ExecBatchSql(l *RpcServiceExecLogic, sqlParams []SqlParam) (rowsAffected int64, err error) {
	for _, elem := range sqlParams {
		//预编译输出
		logx.Info("执行SQL:[%s] 参数:[%+v]", elem.sql, elem.params)
		res, err := l.svcCtx.MySQL.ExecCtx(l.ctx, elem.sql, elem.params...)
		if err != nil {
			return 0, errors.Wrapf(xerr.DB_ERROR, "执行sql出错, sql内容:[%s], 参数:[%+v]", elem.sql, elem.params)
		}
		// 获取影响行数
		rows, err := res.RowsAffected()
		if err != nil {
			return 0, errors.Wrapf(xerr.DB_ERROR, "执行sql出错, sql内容:[%s], 参数:[%+v]", elem.sql, elem.params)
		}
		rowsAffected += rows
	}
	return
}

func ExecCreatSql(l *RpcServiceExecLogic, api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (rowNum int64, err error) {
	sqlParams, err := AssemblCreatSql(api, conds, inMaps)
	if err != nil {
		return 0, err
	}
	return ExecBatchSql(l, sqlParams)
}
func ExecUpdateSql(l *RpcServiceExecLogic, api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (rowNum int64, err error) {
	sqlParams, err := AssemblCreatSql(api, conds, inMaps)
	if err != nil {
		return 0, err
	}
	return ExecBatchSql(l, sqlParams)
}

func ExecDeleteSql(l *RpcServiceExecLogic, api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (rowNum int64, err error) {
	sqlParams, err := AssemblCreatSql(api, conds, inMaps)
	if err != nil {
		return 0, err
	}
	return ExecBatchSql(l, sqlParams)
}

// ExecSelectSql 使用分批次in查询，优化性能
func ExecSelectSql(l *RpcServiceExecLogic, api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (dmVO *dbs.DataMapVO, err error) {
	sqlParams, err := AssemblCreatSql(api, conds, inMaps)
	if err != nil {
		return nil, err
	}
	dmVO = &dbs.DataMapVO{}
	for _, elem := range sqlParams {
		vo, err := utils.QueryRowDataMapVO(l.ctx, l.svcCtx, elem.sql, elem.params...)
		if err != nil {
			return nil, err
		}
		dmVO.Maps = append(dmVO.Maps, vo.Maps...)
	}
	return
}

// AssemblCreatSql 组装 Insert SQL 该sql的list只会形成一条插入
// 配置参数限制insert语句每次插入的数量
// 出参1 形成的sql
// 出参2 预编译参数
func AssemblCreatSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sqlParams []SqlParam, err error) {
	//判断中心关联数据条数是否合规
	if len(conds) != 1 {
		return nil, errors.Wrapf(xerr.DB_ERROR, "%d的中心关联数据数量异常, 查询到的数量为:%d", api.ApiId, len(conds))
	}
	rela := conds[0]

	//检查配置
	if rela.DataModelObhjName.Valid == false || rela.DataModelObhjName.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[数据模型对象名称]字段为空")
	}

	//步骤1 转换数据库映射json配置到map
	var fieldMap map[string]string
	if err = sonic.Unmarshal([]byte(rela.AttrMapping.String), &fieldMap); err != nil {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[属性映射]字段配置有误,不符合json结构:%s, 错误:%v", rela.AttrMapping.String, err)
	}

	//步骤2 组成插入字段sql串
	//为保证有序，先根据数据库将字段映射形成切片
	insertSortList := make([]string, 0, len(fieldMap))
	for _, v := range fieldMap {
		insertSortList = append(insertSortList, v)
	}

	//步骤3 组成insert语句的表字段部分
	fieldSql := " ( " + strings.Join(insertSortList, ",") + " ) "

	//步骤4 组成insert语句的表内容部分, 根据inMaps入参的条数按批次生成对应的插入数量语句
	//新增拆批处理
	sqlParams = make([]SqlParam, 0) //初始化
	for i := 0; i < len(inMaps); i += insertBitchSize {
		//截取当前批次
		curInMaps := inMaps[i:min(i+insertBitchSize, len(inMaps)-1)]

		params := make([]any, 0) //当前批次的参数
		paramSql := ""           //当前批次的插入占位符
		var copies []string
		if copies, err = utils.NCopies(len(fieldMap), "?"); err != nil {
			return nil, err
		}
		for i, inMap := range curInMaps {
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
			if i != len(curInMaps)-1 {
				paramSql += ","
			}
		}

		//步骤5 组装完整SQL
		sql := "INSERT INTO " + rela.DataModelObhjName.String + fieldSql + " VALUES " + paramSql
		sqlParams = append(sqlParams, SqlParam{sql: sql, params: params})
	}
	return
}

// AssemblUpdateSql 组装 Update SQL
// update 每次只允许单条操作，因此不进行拆批处理
func AssemblUpdateSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sqlParams []SqlParam, err error) {
	//判断中心关联数据条数是否合规
	if len(conds) != 1 {
		return nil, errors.Wrapf(xerr.DB_ERROR, "%d的中心关联数据数量异常, 查询到的数量为:%d", api.ApiId, len(conds))
	}
	rela := conds[0]

	//检查配置
	if rela.DataModelObhjName.Valid == false || rela.DataModelObhjName.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[数据模型对象名称]字段为空")
	}

	//步骤1 转换数据库映射json配置到map
	var fieldMap map[string]string
	if err = sonic.Unmarshal([]byte(rela.AttrMapping.String), &fieldMap); err != nil {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[属性映射]字段配置有误,不符合json结构:%s, 错误:%v", rela.AttrMapping.String, err)
	}

	//步骤2 组成条件sql
	if rela.DataObjId.Valid == false || rela.DataObjId.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[数据对象标识]字段配置有误, 错误:%v", err)
	}
	whereSql := rela.DataObjId.String + " = ? "
	//查找主键key
	var primaryKey string
	for key, elem := range fieldMap {
		if strings.TrimSpace(elem) == rela.DataObjId.String {
			primaryKey = strings.TrimSpace(key)
		}
	}
	if primaryKey == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据配置错误, 未找到[数据对象标识]在[属性映射]中的映射关系")
	}

	//步骤3 组成修改Sql切片
	sqlParams = make([]SqlParam, 0) //初始化
	updateList := make([]string, 0) //用来记录 [数据库字段]=？ 这种结构
	params := make([]any, 0)        //记录参数
	for _, inMap := range inMaps {
		for inK, inV := range inMap {
			if fieldMap[inK] != "" {
				updateList = append(updateList, fieldMap[inK]+"=?")
				params = append(params, inV)
			}
		}
		params = append(params, inMap[primaryKey])
		//组装一次sql
		sql := "UPDATE " + rela.DataModelObhjName.String + " SET " + strings.Join(updateList, ", ") + " WHERE " + whereSql
		//添加到切片
		sqlParams = append(sqlParams, SqlParam{
			sql:    sql,
			params: params,
		})
		params = make([]any, 0) //重置
	}
	return
}

// AssemblDeleteSql 组装 delete SQL
// 限制in的条数 进行拆批处理
func AssemblDeleteSql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sqlParams []SqlParam, err error) {
	//判断中心关联数据条数是否合规
	if len(conds) != 1 {
		return nil, errors.Wrapf(xerr.DB_ERROR, "%d的中心关联数据数量异常, 查询到的数量为:%d", api.ApiId, len(conds))
	}
	rela := conds[0]

	//检查配置
	if rela.DataModelObhjName.Valid == false || rela.DataModelObhjName.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据数据模型对象名称字段为空")
	}

	//步骤1 转换数据库映射json配置到map
	var fieldMap map[string]string
	if err = sonic.Unmarshal([]byte(rela.AttrMapping.String), &fieldMap); err != nil {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[属性映射]字段配置有误,不符合json结构:%s, 错误:%v", rela.AttrMapping.String, err)
	}

	//步骤2组成条件sql
	if rela.DataObjId.Valid == false || rela.DataObjId.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[数据对象标识]字段配置有误, 错误:%v", err)
	}
	//查找主键key
	var primaryKey string
	for key, elem := range fieldMap {
		if strings.TrimSpace(elem) == rela.DataObjId.String {
			primaryKey = strings.TrimSpace(key)
		}
	}
	if primaryKey == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据配置错误, 未找到[数据对象标识]在[属性映射]中的映射关系")
	}

	//拆批处理
	sqlParams = make([]SqlParam, 0) //初始化
	for i := 0; i < len(inMaps); i += deleteBitchSize {
		//截取当前批次
		curInMaps := inMaps[i:min(i+deleteBitchSize, len(inMaps)-1)]
		copies, err := utils.NCopies(len(curInMaps), "?")
		if err != nil {
			return nil, err
		}
		params := make([]any, 0)
		for _, inMap := range curInMaps {
			if inMap[primaryKey] == nil {
				return nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "参数中未找到[数据对象标识]值")
			}
			params = append(params, inMap[primaryKey])
		}
		sql := "DELETE FROM " + rela.DataModelObhjName.String + " WHERE " + rela.DataObjId.String + " IN (" + strings.Join(copies, ", ") + ")"
		sqlParams = append(sqlParams, SqlParam{sql: sql, params: params})
	}
	return
}

// AssemblQuerySql 组装 select SQL
// 限制in的条数 进行拆批处理
func AssemblQuerySql(api model.CenterDataApi, conds []model.CDataSrvRela, inMaps []map[string]any) (sqlParams []SqlParam, err error) {
	//判断中心关联数据条数是否合规
	if len(conds) != 1 {
		return nil, errors.Wrapf(xerr.DB_ERROR, "%d的中心关联数据数量异常, 查询到的数量为:%d", api.ApiId, len(conds))
	}
	rela := conds[0]

	//检查配置
	if rela.DataModelObhjName.Valid == false || rela.DataModelObhjName.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据数据模型对象名称字段为空")
	}

	//步骤1 转换数据库映射json配置到map
	var fieldMap map[string]string
	if err = sonic.Unmarshal([]byte(rela.AttrMapping.String), &fieldMap); err != nil {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[属性映射]字段配置有误,不符合json结构:%s, 错误:%v", rela.AttrMapping.String, err)
	}

	//步骤2组成条件sql
	if rela.DataObjId.Valid == false || rela.DataObjId.String == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据[数据对象标识]字段配置有误, 错误:%v", err)
	}
	//查找主键key
	var primaryKey string
	for key, elem := range fieldMap {
		if strings.TrimSpace(elem) == rela.DataObjId.String {
			primaryKey = strings.TrimSpace(key)
		}
	}
	if primaryKey == "" {
		return nil, errors.Wrapf(xerr.SERVER_COMMON_ERROR, "中心关联数据配置错误, 未找到[数据对象标识]在[属性映射]中的映射关系")
	}

	//组装查询输出字段
	selectCondList := make([]string, 0)
	for _, s := range fieldMap {
		selectCondList = append(selectCondList, s)
	}
	selectCond := strings.Join(selectCondList, ", ")

	//组装 拆批处理
	sqlParams = make([]SqlParam, 0) //初始化
	for i := 0; i < len(inMaps); i += deleteBitchSize {
		//截取当前批次
		curInMaps := inMaps[i:min(i+deleteBitchSize, len(inMaps)-1)]
		wherePlaceholder, err := utils.NCopies(len(curInMaps), "?")
		if err != nil {
			return nil, err
		}
		params := make([]any, 0)
		for _, inMap := range curInMaps {
			if inMap[primaryKey] == nil {
				return nil, errors.Wrapf(xerr.REUQEST_PARAM_ERROR, "参数中未找到[数据对象标识]值")
			}
			params = append(params, inMap[primaryKey])
		}

		sql := "SELECT " + selectCond + " FROM " + rela.DataModelObhjName.String + " WHERE " + rela.DataObjId.String + " IN (" + strings.Join(wherePlaceholder, ", ") + ")"
		sqlParams = append(sqlParams, SqlParam{sql: sql, params: params})
	}
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
