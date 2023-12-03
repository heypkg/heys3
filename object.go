package s3

import (
	"fmt"
	"time"

	"github.com/heypkg/store/jsontype"
	"github.com/heypkg/store/utils"
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
