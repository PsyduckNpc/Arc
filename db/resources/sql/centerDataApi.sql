CREATE TABLE center_data_api
(
    after_sql      TEXT NULL COMMENT '后置SQL',
    sql_param      TEXT NULL COMMENT '自定义SQL入参',
    api_id         DECIMAL(16) NOT NULL COMMENT '接口标识',
    center_name    TEXT NULL COMMENT '中心名称',
    api_name       TEXT NULL COMMENT '接口名称',
    api_path       TEXT NULL COMMENT '接口路径',
    op_type        VARCHAR(8) NULL COMMENT '操作类型',
    call_source    VARCHAR(32) NULL COMMENT '调用来源',
    api_param      TEXT NULL COMMENT '接口入参',
    before_sql     TEXT NULL COMMENT '前置SQL',
    decrypt_flag   VARCHAR(8) NULL COMMENT '属性描述：解密标识',
    decrypt_fld    TEXT NULL COMMENT '属性描述：解密字段',
    before_extend  LONGTEXT NULL COMMENT '属性描述：before_sql的扩展字段',
    before_extend2 LONGTEXT NULL COMMENT '属性描述：before_sql的扩展字段',
    PRIMARY KEY (api_id) -- 将主键声明到表结构外部
) COMMENT '中心数据服务接口';