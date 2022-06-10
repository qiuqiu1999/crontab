package master

import (
	"context"
	"github.com/qiuqiu1999/crontab/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// mongodb日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() error {
	var (
		client *mongo.Client
		err    error
	)

	if client, err = common.InitMongo(common.MongoConfig{
		ConnectTimeOut: time.Duration(G_config.MongodbConnectTimeout),
		Uri:            G_config.MongodbUri,
	}); err != nil {
		return err
	}

	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return nil
}

// 查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) ([]*common.JobLog, error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
		err     error
	)

	// len(logArr)
	logArr := make([]*common.JobLog, 0)

	// 过滤条件
	filter = &common.JobLogFilter{JobName: name}

	// 按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	findOpt := options.Find()
	findOpt.SetSort(logSort)
	findOpt.SetSkip(int64(skip))
	findOpt.SetLimit(int64(limit))
	findOpt.SetSort(logSort)

	// 查询
	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, findOpt); err != nil {
		return nil, err
	}
	// 延迟释放游标
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue // 有日志不合法
		}

		logArr = append(logArr, jobLog)
	}
	return logArr, nil
}
