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
        "license": {
            "name": "Open Government Licence v3.0",
            "url": "http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/files": {
            "get": {
                "description": "GETs metadata for a file",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "File upload started"
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "description": "POSTs metadata for a file when an upload has started.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "File upload started"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "name": "collection_id",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "etag",
                        "in": "formData"
                    },
                    {
                        "type": "boolean",
                        "name": "is_publishable",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "licence",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "licence_url",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "path",
                        "in": "formData"
                    },
                    {
                        "type": "integer",
                        "name": "size_in_bytes",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "state",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "title",
                        "in": "formData"
                    },
                    {
                        "type": "string",
                        "name": "type",
                        "in": "formData"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/files/{filepath}": {
            "patch": {
                "description": "PATCH metadata state for a file, (i.e. when an upload has completed).",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Patch metadata state"
                ],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Filepath of required file",
                        "name": "file_path",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Change the state of a file in the metadata",
                        "name": "patch_file",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/store.ChangeFileState"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "409": {
                        "description": "Conflict"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        }
    },
    "definitions": {
        "store.ChangeFileState": {
            "type": "object",
            "required": [
                "etag",
                "state"
            ],
            "properties": {
                "etag": {
                    "type": "string"
                },
                "state": {
                    "type": "string"
                }
            }
        }
    },
    "tags": [
        {
            "name": "private"
        }
    ]
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:26900",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "dp-files-api",
	Description:      "Digital Publishing API for handling file metadata.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
