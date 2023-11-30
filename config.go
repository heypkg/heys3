package s3

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var (
	defaultDatabase      *gorm.DB        = nil
	defaultStorageType                   = "mongo"
	defaultMongoDatabase *mongo.Database = nil
	defaultStorageDir                    = "./s3"
	defaultSecret                        = "heypkg2023!!"
)

func SetupMongoStorage(db *gorm.DB, mdb *mongo.Database) {
	defaultDatabase = db
	db.AutoMigrate(&S3Object{})

	defaultStorageType = "mongo"
	defaultMongoDatabase = mdb
}

func SetupFileStorage(db *gorm.DB, dir string) {
	defaultDatabase = db
	db.AutoMigrate(&S3Object{})

	defaultStorageType = "fs"
	defaultStorageDir = dir
}
