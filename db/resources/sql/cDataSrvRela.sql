-- auto-generated definition
create table c_data_srv_rela
(
    rela_id              decimal(16) not null comment '关联标识',
    api_id               decimal(16) null comment '中心数据服务接口主键',
    sql_logic            varchar(32) null comment 'SQL逻辑符',
    sql_condition        text        null comment 'SQL条件',
    sql_sort             text        null comment 'SQL条件排序',
    fld_type_obhj_name   text        null comment '领域模型对象名称',
    data_model_obhj_name text        null comment '数据模型对象名称',
    data_obj_id          varchar(32) null comment '数据对象标识',
    rela_data_obj_id     text        null comment '关联数据对象标识',
    attr_mapping         text        null comment '属性映射',
    rela_mapping         varchar(32) null comment '关系映射',
    PRIMARY KEY (rela_id)
)
    comment '中心数据服务关联表';
