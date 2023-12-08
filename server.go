package s3

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"time"

	"github.com/heypkg/store/utils"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type S3Server struct {
	logger *zap.Logger

	storageType string
	db          *gorm.DB
	mdb         *mongo.Database
	storageDir  string
}

func NewMongoStorageServer(db *gorm.DB, mdb *mongo.Database, secret string) *S3Server {
	s := S3Server{
		logger:      zap.L().Named("s3"),
		storageType: "mongo",
		db:          db,
		mdb:         mdb,
	}
	defaultSecret = secret
	return &s
}

func NewFileStorageServer(db *gorm.DB, dir string, secret string) *S3Server {
	s := S3Server{
		logger:      zap.L().Named("s3"),
		storageType: "fs",
		db:          db,
		storageDir:  dir,
	}
	defaultSecret = secret
	return &s
}

func (s *S3Server) GetDB() *gorm.DB {
	return s.db
}

func (s *S3Server) PutObject(schema string, bucket string, key string, filename string, src io.Reader) (*S3Object, error) {
	buffer, err := io.ReadAll(src)
	if err != nil {
		return nil, errors.Wrap(err, "read Object")
	}

	object, err := s.GetObject(schema, bucket, key)
	if err != nil {
		return nil, errors.Wrap(err, "get object")
	}

	if object == nil {
		object = &S3Object{
			Schema: schema,
			Bucket: bucket,
			Key:    key,
		}
		if err := s.putObjectFile(object, filename, buffer); err != nil {
			return nil, errors.Wrap(err, "put file")
		}
		if result := s.db.Save(object); result.Error != nil {
			return nil, errors.Wrap(result.Error, "save object")
		}
	} else {
		s.removeObjectFile(object)
		if err := s.putObjectFile(object, filename, buffer); err != nil {
			return nil, errors.Wrap(err, "put file")
		}
		updateColumns := []string{"file_name", "file_size", "signature", "sign_method"}
		result := s.db.Model(object).Where("id = ?", object.ID).Select(updateColumns).Updates(object)
		if result.Error != nil {
			return nil, errors.Wrap(result.Error, "update object")
		}
	}
	return object, nil
}

func (s *S3Server) GetObjectById(schema string, id uint) (*S3Object, error) {
	object := &S3Object{}
	result := s.db.Where("schema = ? AND id = ?", schema, id).First(object)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return object, nil
}

func (s *S3Server) GetObject(schema string, bucket string, key string) (*S3Object, error) {
	object := &S3Object{}
	result := s.db.Where("schema = ? AND bucket = ? AND key = ?", schema, bucket, key).First(object)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return object, nil
}

func (s *S3Server) RemoveObjectById(schema string, id uint) {
	obj, _ := s.GetObjectById(schema, id)
	if obj != nil {
		s.db.Unscoped().Delete(obj)
	}
}

func (s *S3Server) RemoveObject(schema string, bucket string, key string) {
	obj, _ := s.GetObject(schema, bucket, key)
	if obj != nil {
		s.db.Unscoped().Delete(obj)
		s.removeObjectFile(obj)
	}
}

func (s *S3Server) GetObjectContent(obj *S3Object) ([]byte, error) {
	if obj == nil {
		return []byte{}, errors.New("invalid object")
	}
	if s.storageType == "mongo" {
		coll := s.mdb.Collection(obj.Bucket)
		if coll == nil {
			return nil, errors.New("invalid collection")
		}
		var result bson.M
		err := coll.FindOne(context.Background(), bson.M{"key": obj.Key}).Decode(&result)
		if err != nil {
			return nil, err
		}
		content, ok := result["content"].(primitive.Binary)
		if !ok {
			return nil, errors.New("content not found")
		}
		return content.Data, nil
	} else {
		filename := filepath.Join(obj.Signature, obj.FileName)
		buffer, err := utils.ReadFileFromDir(s.storageDir, filename)
		return buffer, err
	}
}

func (s *S3Server) DropObjectsBeforeInterval(duration time.Duration) error {
	interval := duration + time.Hour*48
	threshold := time.Now().Add(-interval)
	result := s.db.Unscoped().Where("updated < ?", threshold).Delete(&S3Object{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "delete records")
	}

	return nil
}

func (s *S3Server) DropInvalidObjectFiles(schema string, bucket string) error {
	if s.storageType == "mongo" {
		coll := s.mdb.Collection(bucket)
		if coll == nil {
			return errors.New("invalid collection")
		}
		cur, err := coll.Find(context.Background(), bson.D{}, options.Find())
		if err != nil {
			return err
		}
		defer cur.Close(context.Background())

		deleteKeys := []string{}
		for cur.Next(context.Background()) {
			var result bson.M
			err := cur.Decode(&result)
			if err != nil {
				return err
			}
			key, ok := result["key"].(string)
			if !ok {
				deleteKeys = append(deleteKeys, key)
			} else if obj, err := s.GetObject(schema, bucket, key); err == nil && obj == nil {
				deleteKeys = append(deleteKeys, key)
			}
		}
		if err := cur.Err(); err != nil {
			return err
		}

		for _, key := range deleteKeys {
			_, err := coll.DeleteOne(context.Background(), bson.M{"key": key})
			if err != nil {
				return errors.Wrap(err, "delete object from collection")
			}
		}
	}
	return nil
}

func (s S3Server) removeObjectFile(obj *S3Object) error {
	if obj.Bucket == "" || obj.Key == "" {
		return nil
	}
	if s.storageType == "mongo" {
		coll := s.mdb.Collection(obj.Bucket)
		if coll == nil {
			return errors.New("invalid collection")
		}
		_, err := coll.DeleteOne(context.Background(), bson.M{"key": obj.Key})
		if err != nil {
			return errors.Wrap(err, "delete file from MongoDB")
		}
	} else {
		utils.RemoveFileFromDir(s.storageDir, filepath.Join(obj.Signature, obj.FileName))
	}
	return nil
}

func (s *S3Server) putObjectFile(obj *S3Object, filename string, buffer []byte) error {
	size := int64(len(buffer))
	sign, err := utils.GetSignatureSha256(bytes.NewReader(buffer), nil)
	if err != nil {
		return errors.Wrap(err, "get sign")
	}
	obj.FileName = filename
	obj.FileSize = size
	obj.SignMethod = "sha256"
	obj.Signature = sign

	if s.storageType == "mongo" {
		coll := s.mdb.Collection(obj.Bucket)
		if coll == nil {
			return errors.New("invalid collection")
		}
		fileData := bson.M{
			"key":         obj.Key,
			"content":     buffer,
			"filename":    obj.FileName,
			"sign":        obj.Signature,
			"sign_method": obj.SignMethod,
			"metadata":    bson.M{},
		}
		_, err = coll.InsertOne(context.Background(), fileData)
		if err != nil {
			return errors.Wrap(err, "save Object")
		}
	} else {
		n, err := utils.WriteFileToDir(s.storageDir, filepath.Join(obj.Signature, obj.FileName), bytes.NewReader(buffer))
		if err != nil || n != size {
			return errors.Wrap(err, "save Object")
		}
	}
	return nil
}
