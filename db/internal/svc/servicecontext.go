package svc

import (
	"Arc/db/internal/comm/utils/xerr"
	"Arc/db/internal/config"
	"Arc/db/internal/config/dbs"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	MySQL       sqlx.SqlConn
	CacheMySQL  sqlc.CachedConn
	RedisClient *redis.Redis
	LocalCache  *collection.Cache
}

func NewServiceContext(c config.Config) *ServiceContext {

	//初始化基础服务配置

	//Mysql服务连接必须提前启动
	conn := dbs.NewMySQLConnect(c.MySQLConfig)

	//redis服务必须提前启动
	redisClient := redis.MustNewRedis(c.Redis.RedisConf)

	//Cache中redis必须提前启动 新版初始化方式（自动加载多级缓存）
	//sqlc 只支持主键，唯一键的单条记录索引方式。其他批量查询的方式不支持。
	cacheConn := sqlc.NewConn(conn, c.DBSCache)

	//用例 先放在下面
	//主键缓存
	//func GetUserById(ctx context.Context, id int64) (*User, error) {
	//	var user User
	//	cacheKey := fmt.Sprintf("user:%d", id) // 推荐格式：表名:主键值
	//
	//	err := cachedConn.QueryRowCtx(ctx, &user, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
	//		return conn.QueryRowCtx(ctx, v, "SELECT * FROM user WHERE id = ?", id)
	//	})
	//
	//	if err != nil {
	//		return nil, fmt.Errorf("query user error: %v", err)
	//	}
	//	return &user, nil
	//}

	//唯一键缓存
	//func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	//	var user User
	//	cacheKey := fmt.Sprintf("user:email:%s", email) // 唯一键缓存键格式
	//
	//	err := cachedConn.QueryRowIndexCtx(ctx, &user, cacheKey,
	//		func(primary any) string {
	//			return fmt.Sprintf("user:%d", primary) // 主键缓存键生成
	//		},
	//		func(ctx context.Context, conn sqlx.SqlConn, v any) (any, error) {
	//			// 通过唯一键查询主键
	//			var id int64
	//			err := conn.QueryRowCtx(ctx, &id, "SELECT id FROM user WHERE email = ?", email)
	//			return id, err
	//		},
	//		func(ctx context.Context, conn sqlx.SqlConn, v any, primary any) error {
	//			// 通过主键查询完整数据
	//			return conn.QueryRowCtx(ctx, v, "SELECT * FROM user WHERE id = ?", primary)
	//		},
	//	)
	//
	//	if err != nil {
	//		return nil, fmt.Errorf("query user error: %v", err)
	//	}
	//	return &user, nil
	//}
	localCache, err := collection.NewCache(
		-1,                           // 默认过期时间
		collection.WithLimit(10240),  // 最大存储数量（超出时自动淘汰旧数据）
		collection.WithName("local"), // 缓存名称（日志中标识）
	)
	if err != nil {
		panic(errors.Wrapf(xerr.SERVER_COMMON_ERROR, "初始化本地缓存失败"))
	}

	//加载初始化数据
	initData(localCache)

	return &ServiceContext{
		Config:      c,
		MySQL:       conn,
		RedisClient: redisClient,
		CacheMySQL:  cacheConn,
		LocalCache:  localCache,
	}
}

func initData(localCache *collection.Cache) {
	//todo 初始化数据到rdis
}
