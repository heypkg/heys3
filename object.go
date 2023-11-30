package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/heypkg/store/jsontype"
	"github.com/heypkg/store/utils"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type Tags map[string]any

type S3Object struct {
	ID      uint              `json:"Id" gorm:"primarykey"`
	Updated jsontype.JSONTime `json:"Updated" gorm:"autoUpdateTime"`
	Created jsontype.JSONTime `json:"Created" gorm:"autoCreateTime"`
	Deleted gorm.DeletedAt    `json:"Deleted" gorm:"index"`

	Schema      string                   `json:"Schema" gorm:"uniqueIndex:idx_s3_object_unique"`
	Bucket      string                   `json:"Bucket" gorm:"uniqueIndex:idx_s3_object_unique"`
	Key         string                   `json:"Key" gorm:"uniqueIndex:idx_s3_object_unique"`
	FileName    string                   `json:"FileName"`
	FileSize    int64                    `json:"FileSize"`
	Signature   string                   `json:"Signature"`
	SignMethod  string                   `json:"SignMethod"`
	MetaDataRaw jsontype.JSONType[*Tags] `json:"-" gorm:"column:meta_data"`
	MetaData    *Tags                    `json:"MetaData" gorm:"-"`
	DownloadUrl string                   `json:"DownloadUrl" gorm:"-"`
}

func (m *S3Object) BeforeSave(tx *gorm.DB) (err error) {
	if m.MetaData != nil {
		m.MetaDataRaw = jsontype.NewJSONType(m.MetaData)
	}
	return nil
}

func (m *S3Object) AfterFind(tx *gorm.DB) (err error) {
	m.MetaData = m.MetaDataRaw.Data
	m.DownloadUrl = m.makeDownloadUrl()
	return nil
}

func (m *S3Object) RemoveFile() error {
	if m.Bucket == "" || m.Key == "" {
		return nil
	}
	if defaultStorageType == "mongo" && defaultMongoDatabase != nil {
		mdb := defaultMongoDatabase
		coll := mdb.Collection(m.Bucket)
		if coll == nil {
			return errors.New("invalid collection")
		}
		_, err := coll.DeleteOne(context.Background(), bson.M{"key": m.Key})
		if err != nil {
			return errors.Wrap(err, "delete file from MongoDB")
		}
	} else {
		utils.RemoveFileFromDir(defaultStorageDir, filepath.Join(m.Signature, m.FileName))
	}
	return nil
}

func (m *S3Object) PutFile(filename string, buffer []byte) error {
	size := int64(len(buffer))
	sign, err := utils.GetSignatureSha256(bytes.NewReader(buffer), nil)
	if err != nil {
		return errors.Wrap(err, "get sign")
	}
	m.FileName = filename
	m.FileSize = size
	m.SignMethod = "sha256"
	m.Signature = sign

	if defaultStorageType == "mongo" && defaultMongoDatabase != nil {
		mdb := defaultMongoDatabase
		coll := mdb.Collection(m.Bucket)
		if coll == nil {
			return errors.New("invalid collection")
		}
		fileData := bson.M{
			"key":         m.Key,
			"content":     buffer,
			"filename":    m.FileName,
			"sign":        m.Signature,
			"sign_method": m.SignMethod,
			"metadata":    bson.M{},
		}
		_, err = coll.InsertOne(context.Background(), fileData)
		if err != nil {
			return errors.Wrap(err, "save Object")
		}
	} else {
		n, err := utils.WriteFileToDir(defaultStorageDir, filepath.Join(m.Signature, m.FileName), bytes.NewReader(buffer))
		if err != nil || n != size {
			return errors.Wrap(err, "save Object")
		}
	}
	return nil
}

func (m *S3Object) AfterDelete(tx *gorm.DB) (err error) {
	if m.ID == 0 {
		return nil
	}
	m.RemoveFile()
	return nil
}

func (m *S3Object) makeDownloadUrl() string {
	return MakeUrl(m.Schema, m.Bucket, m.Key)
}

func CreateAccessKey(schema string, bucket string, key string) string {
	raw := fmt.Sprintf("%v/%v@%v", bucket, key, schema)
	return utils.CRC32(raw)
}

func MakeUrl(schema string, bucket string, key string) string {
	accessKey := CreateAccessKey(schema, bucket, key)
	token, err := CreateAccessToken("", accessKey, time.Hour)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("/api/v1/s3/objects/%v/%v?token=%v", bucket, key, token)
}

func PutObject(schema string, bucket string, key string, filename string, src io.Reader) (*S3Object, error) {
	db := defaultDatabase
	buffer, err := io.ReadAll(src)
	if err != nil {
		return nil, errors.Wrap(err, "read Object")
	}

	object, err := GetObject(schema, bucket, key)
	if err != nil {
		return nil, errors.Wrap(err, "get object")
	}

	if object == nil {
		object = &S3Object{
			Schema: schema,
			Bucket: bucket,
			Key:    key,
		}
		if err := object.PutFile(filename, buffer); err != nil {
			return nil, errors.Wrap(err, "put file")
		}
		if result := db.Save(object); result.Error != nil {
			return nil, errors.Wrap(result.Error, "save object")
		}
	} else {
		object.RemoveFile()
		if err := object.PutFile(filename, buffer); err != nil {
			return nil, errors.Wrap(err, "put file")
		}
		updateColumns := []string{"file_name", "file_size", "signature", "sign_method"}
		result := db.Model(object).Where("id = ?", object.ID).Select(updateColumns).Updates(object)
		if result.Error != nil {
			return nil, errors.Wrap(result.Error, "update object")
		}
	}
	return object, nil
}

func GetObjectById(schema string, id uint) (*S3Object, error) {
	db := defaultDatabase
	object := &S3Object{}
	result := db.Where("schema = ? AND id = ?", schema, id).First(object)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return object, nil
}

func GetObject(schema string, bucket string, key string) (*S3Object, error) {
	db := defaultDatabase
	object := &S3Object{}
	result := db.Where("schema = ? AND bucket = ? AND key = ?", schema, bucket, key).First(object)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return object, nil
}

func RemoveObjectById(schema string, id uint) {
	db := defaultDatabase
	obj, _ := GetObjectById(schema, id)
	if obj != nil {
		db.Unscoped().Delete(obj)
	}
}

func RemoveObject(schema string, bucket string, key string) {
	db := defaultDatabase
	obj, _ := GetObject(schema, bucket, key)
	if obj != nil {
		db.Unscoped().Delete(obj)
	}
}

func GetObjectContent(obj *S3Object) ([]byte, error) {
	if obj == nil {
		return []byte{}, errors.New("invalid object")
	}
	if defaultStorageType == "mongo" && defaultMongoDatabase != nil {
		mdb := defaultMongoDatabase
		coll := mdb.Collection(obj.Bucket)
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
		buffer, err := utils.ReadFileFromDir(defaultStorageDir, filename)
		return buffer, err
	}
}

func DropObjectsBeforeInterval(duration time.Duration) error {
	db := defaultDatabase
	interval := duration + time.Hour*48
	threshold := time.Now().Add(-interval)
	result := db.Unscoped().Where("updated < ?", threshold).Delete(&S3Object{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "delete records")
	}

	return nil
}

func DropInvalidObjectFiles(schema string, bucket string) error {
	if defaultStorageType == "mongo" && defaultMongoDatabase != nil {
		mdb := defaultMongoDatabase
		coll := mdb.Collection(bucket)
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
			} else if obj, err := GetObject(schema, bucket, key); err == nil && obj == nil {
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
