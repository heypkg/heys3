package s3

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/heypkg/store/echohandler"
	"github.com/heypkg/store/utils"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// @title S3 API
// @version 1.0
// @host dev.netdoop.com
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func SetupEchoGroup(group *echo.Group) *echo.Group {
	group.GET("/objects", HandleListObjects)
	group.POST("/objects/:bucket/:key", HandlePutObject)
	group.PUT("/objects/:bucket/:key", HandlePutObject)
	group.GET("/objects/:bucket/:key/info", HandleGetObjectInfo, S3ObjectHandler)
	group.GET("/objects/:bucket/:key", HandleGetObject, S3ObjectHandler)
	group.DELETE("/objects/:bucket/:key", HandleDeleteObject, S3ObjectHandler)
	return group
}

func getS3ObjectFromEchoContext(c echo.Context) *S3Object {
	if v := c.Get(utils.GetRawTypeName(S3Object{})); v != nil {
		if object, ok := v.(*S3Object); ok {
			return object
		}
	}
	return nil
}

func S3ObjectHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		schema := cast.ToString(c.Get("schema"))
		bucket := cast.ToString(c.Param("bucket"))
		key := cast.ToString(c.Param("key"))
		object, err := GetObject(schema, bucket, key)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		if object == nil {
			return echo.NewHTTPError(http.StatusNotFound, "not found")

		}
		c.Set(utils.GetRawTypeName(object), object)
		return next(c)
	}
}

type listS3ObjectsData struct {
	Data  []S3Object `json:"Data"`
	Total int64      `json:"Total"`
}

// @Summary List objects
// @Tags Objects
// @ID list-objects
// @Accept json
// @Produce json
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param order_by query string false "Sort order" default()
// @Param q query string false "Query" default()
// @Security Bearer
// @Success 200 {object} listS3ObjectsData"List of objects"
// @Failure 401 {object} echo.HTTPError "Unauthorized"
// @Failure 500 {object} echo.HTTPError "Internal server error"
// @Router /s3/objects [GET]
func HandleListObjects(c echo.Context) error {
	db := getDefaultServer().GetDB()
	data, total, err := echohandler.ListObjects[S3Object](db, c, nil, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set("X-Total", fmt.Sprintf("%v", total))
	return c.JSON(http.StatusOK, listS3ObjectsData{Data: data, Total: total})
}

// @Summary Put an object to S3
// @Tags Objects
// @ID put-object
// @Accept multipart/form-data
// @Produce json
// @Param bucket path string true "S3 objet bucket" default()
// @Param key path string true "S3 object key" default()
// @Param file formData file true "File to upload"
// @Security Bearer
// @Success 200 {object} s3.S3Object "Object information"
// @Failure 400 {object} echo.HTTPError "Bad request"
// @Failure 500 {object} echo.HTTPError "Internal server error"
// @Router /s3/objects/{bucket}/{key} [POST]
func HandlePutObject(c echo.Context) error {
	schema := cast.ToString(c.Get("schema"))
	bucket := cast.ToString(c.Param("bucket"))
	key := cast.ToString(c.Param("key"))

	req := c.Request()
	contentType := req.Header.Get("Content-Type")

	var (
		src      io.Reader
		filename string
	)

	if contentType == "" || strings.HasPrefix(contentType, "text/plain") {
		filename = key
		defer req.Body.Close()
		src = req.Body
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Wrap(err, "read Object").Error())
		}
		filename = file.Filename
		f, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "open Object").Error())
		}
		defer f.Close()
		src = f
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content type "+contentType)
	}

	object, err := PutObject(schema, bucket, key, filename, src)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "put object").Error())
	}
	return c.JSON(http.StatusOK, object)
}

// @Summary Delete an object from S3
// @Tags Objects
// @ID delete-object
// @Accept json
// @Produce json
// @Security Bearer
// @Param bucket path string true "S3 objet bucket" default()
// @Param key path string true "S3 object key" default()
// @Success 204
// @Failure 401 {object} echo.HTTPError "Unauthorized"
// @Failure 500 {object} echo.HTTPError "Internal server error"
// @Router /s3/objects/{bucket}/{key} [DELETE]
func HandleDeleteObject(c echo.Context) error {
	object := getS3ObjectFromEchoContext(c)
	RemoveObjectById(object.Schema, object.ID)
	return c.NoContent(http.StatusNoContent)
}

// @Summary Get information for an object from S3
// @Tags Objects
// @ID get-object-info
// @Accept json
// @Produce json
// @Security Bearer
// @Param bucket path string true "S3 objet bucket" default()
// @Param key path string true "S3 object key" default()
// @Success 200 {object} s3.S3Object "Object information"
// @Failure 401 {object} echo.HTTPError "Unauthorized"
// @Failure 500 {object} echo.HTTPError "Internal server error"
// @Router /s3/objects/{bucket}/{key}/info [GET]
func HandleGetObjectInfo(c echo.Context) error {
	object := getS3ObjectFromEchoContext(c)
	return c.JSON(http.StatusOK, object)
}

// @Summary Download an object from S3
// @Tags Objects
// @ID download-object
// @Accept json
// @Produce octet-stream
// @Security Bearer
// @Param bucket path string true "S3 objet bucket" default()
// @Param key path string true "S3 object key" default()
// @Param token query string true "Token for the object"
// @Success 200 {object} nil "Empty response"
// @Failure 401 {object} echo.HTTPError "Unauthorized"
// @Failure 500 {object} echo.HTTPError "Internal server error"
// @Router /s3/objects/{bucket}/{key} [GET]
func HandleGetObject(c echo.Context) error {
	object := getS3ObjectFromEchoContext(c)
	accessKey := cast.ToString(c.Get("accessKey"))
	if tmp := CreateAccessKey(object.Schema, object.Bucket, object.Key); accessKey != tmp {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	// Set response headers
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; Filename="+object.FileName)
	c.Response().Header().Set(echo.HeaderAccessControlExposeHeaders, "Content-Disposition")

	buffer, err := GetObjectContent(object)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "read Object").Error())
	}

	// Stream Object to response body
	if _, err = io.Copy(c.Response().Writer, bytes.NewReader(buffer)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "read Object").Error())
	}
	return nil
}
