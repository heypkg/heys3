// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/s3/objects": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Objects"
                ],
                "summary": "List objects",
                "operationId": "list-objects",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 1,
                        "description": "Page",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "default": 20,
                        "description": "Page size",
                        "name": "page_size",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "",
                        "description": "Sort order",
                        "name": "order_by",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "",
                        "description": "Query",
                        "name": "q",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "List of objects",
                        "schema": {
                            "$ref": "#/definitions/s3.listS3ObjectsData"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/s3/objects/{bucket}/{key}": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/octet-stream"
                ],
                "tags": [
                    "Objects"
                ],
                "summary": "Download an object from S3",
                "operationId": "download-object",
                "parameters": [
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 objet bucket",
                        "name": "bucket",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 object key",
                        "name": "key",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Token for the object",
                        "name": "token",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Empty response"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Objects"
                ],
                "summary": "Put an object to S3",
                "operationId": "put-object",
                "parameters": [
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 objet bucket",
                        "name": "bucket",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 object key",
                        "name": "key",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "file",
                        "description": "File to upload",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Object information",
                        "schema": {
                            "$ref": "#/definitions/s3.S3Object"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Objects"
                ],
                "summary": "Delete an object from S3",
                "operationId": "delete-object",
                "parameters": [
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 objet bucket",
                        "name": "bucket",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 object key",
                        "name": "key",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/s3/objects/{bucket}/{key}/info": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Objects"
                ],
                "summary": "Get information for an object from S3",
                "operationId": "get-object-info",
                "parameters": [
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 objet bucket",
                        "name": "bucket",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "default": "",
                        "description": "S3 object key",
                        "name": "key",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Object information",
                        "schema": {
                            "$ref": "#/definitions/s3.S3Object"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "echo.HTTPError": {
            "type": "object",
            "properties": {
                "message": {}
            }
        },
        "gorm.DeletedAt": {
            "type": "object",
            "properties": {
                "time": {
                    "type": "string"
                },
                "valid": {
                    "description": "Valid is true if Time is not NULL",
                    "type": "boolean"
                }
            }
        },
        "s3.S3Object": {
            "type": "object",
            "properties": {
                "Bucket": {
                    "type": "string"
                },
                "Created": {
                    "type": "string"
                },
                "Deleted": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "DownloadUrl": {
                    "type": "string"
                },
                "FileName": {
                    "type": "string"
                },
                "FileSize": {
                    "type": "integer"
                },
                "Id": {
                    "type": "integer"
                },
                "Key": {
                    "type": "string"
                },
                "MetaData": {
                    "$ref": "#/definitions/s3.Tags"
                },
                "Schema": {
                    "type": "string"
                },
                "SignMethod": {
                    "type": "string"
                },
                "Signature": {
                    "type": "string"
                },
                "Updated": {
                    "type": "string"
                }
            }
        },
        "s3.Tags": {
            "type": "object",
            "additionalProperties": {}
        },
        "s3.listS3ObjectsData": {
            "type": "object",
            "properties": {
                "Data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/s3.S3Object"
                    }
                },
                "Total": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "description": "Type \"Bearer\" followed by a space and JWT token.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "dev.netdoop.com",
	BasePath:         "/api/v1",
	Schemes:          []string{"http"},
	Title:            "S3 API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
