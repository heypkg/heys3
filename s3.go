package s3

import (
	"io"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var defaultServerMutex sync.Mutex
var defaultServer *S3Server

func SetupMongoStorage(db *gorm.DB, mdb *mongo.Database, secret string) {
	defaultSecret = secret
	logger := zap.L().Named("s3")

	if db == nil {
		logger.Fatal("invalid database")
	}
	if mdb == nil {
		logger.Fatal("invalid mongo database")
	}

	defaultServerMutex.Lock()
	defer defaultServerMutex.Unlock()
	defaultServer = NewMongoStorageServer(db, mdb)
}

func SetupFileStorage(db *gorm.DB, dir string, secret string) {
	defaultSecret = secret
	logger := zap.L().Named("s3")

	if db == nil {
		logger.Fatal("invalid database")
	}

	defaultServerMutex.Lock()
	defer defaultServerMutex.Unlock()
	defaultServer = NewFileStorageServer(db, dir)
}

func getDefaultServer() *S3Server {
	return defaultServer
}

func PutObject(schema string, bucket string, key string, filename string, src io.Reader) (*S3Object, error) {
	return getDefaultServer().PutObject(schema, bucket, key, filename, src)
}

func GetObjectById(schema string, id uint) (*S3Object, error) {
	return getDefaultServer().GetObjectById(schema, id)
}

func GetObject(schema string, bucket string, key string) (*S3Object, error) {
	return getDefaultServer().GetObject(schema, bucket, key)
}

func RemoveObjectById(schema string, id uint) {
	getDefaultServer().RemoveObjectById(schema, id)
}

func RemoveObject(schema string, bucket string, key string) {
	getDefaultServer().RemoveObject(schema, bucket, key)
}

func GetObjectContent(obj *S3Object) ([]byte, error) {
	return getDefaultServer().GetObjectContent(obj)
}

func DropObjectsBeforeInterval(duration time.Duration) error {
	return getDefaultServer().DropObjectsBeforeInterval(duration)
}

func DropInvalidObjectFiles(schema string, bucket string) error {
	return getDefaultServer().DropInvalidObjectFiles(schema, bucket)
}
