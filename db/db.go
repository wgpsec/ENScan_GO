package db

import (
	"github.com/adjust/rmq/v4"
	"github.com/go-redis/redis"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/wgpsec/ENScan/common"
	"github.com/wgpsec/ENScan/common/utils/gologger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
	"log"
)

var RedisBb *redis.Client
var MongoBb *mongo.Client
var RmqC rmq.Connection

func InitMongo(enOptions *common.ENOptions) {
	clientOptions := options.Client().ApplyURI(enOptions.ENConfig.Api.Mongodb)
	var ctx = context.TODO()
	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	MongoBb = client
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = MongoBb.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	databases, err := MongoBb.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	gologger.Infof("MongoDB! Server %v", databases)
}

func InitRedis(enOptions *common.ENOptions) {
	//连接服务器
	RedisBb = redis.NewClient(&redis.Options{
		Addr:     enOptions.ENConfig.Api.Server + ":6379", // use default Addr
		Password: enOptions.ENConfig.Api.Redis,            // no password set
		DB:       0,                                       // use default DB
	})
}

func InitQueue(enOptions *common.ENOptions) {
	var err error
	redisClient := redis2.NewClient(&redis2.Options{Network: "tcp", Addr: enOptions.ENConfig.Api.Server + ":6379", DB: 1, Password: enOptions.ENConfig.Api.Redis})
	RmqC, err = rmq.OpenConnectionWithRedisClient("enscan-service", redisClient, nil)
	if err != nil {
		panic(err)
	}
}
