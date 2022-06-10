package common

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoConfig struct {
	ConnectTimeOut time.Duration
	Uri            string
}

func InitMongo(config MongoConfig) (*mongo.Client, error) {
	opt := &options.ClientOptions{}
	opt.SetConnectTimeout(time.Duration(config.ConnectTimeOut) * time.Millisecond)
	opt.ApplyURI(config.Uri)
	// 建立mongodb连接
	client, err := mongo.Connect(context.TODO(), opt)
	if err != nil {
		return nil, err
	}
	return client, nil
}
