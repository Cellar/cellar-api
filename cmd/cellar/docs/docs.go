// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2020-12-04 21:49:07.577935009 -0700 MST m=+31.401921901

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Parker Johansen",
            "email": "johansen.parker@gmail.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://gitlab.com/cellar-app/cellar-api/-/blob/148abea87dfbba32ab1aefc1ab36b2de1f652c9e/LICENSE.txt"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health-check": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "summary": "Health Check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.HealthResponse"
                        }
                    }
                }
            }
        },
        "/v1/secrets": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Create Secret",
                "parameters": [
                    {
                        "description": "Add secret",
                        "name": "secret",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.CreateSecretRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/models.SecretMetadataResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    }
                }
            }
        },
        "/v1/secrets/{id}": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get Secret Metadata",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Secret ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.SecretMetadataResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    }
                }
            },
            "delete": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Delete Secret",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Secret ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {},
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    }
                }
            }
        },
        "/v1/secrets/{id}/access": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Access Secret Content",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Secret ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.SecretContentResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/httputil.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "httputil.HTTPError": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer",
                    "example": 400
                },
                "message": {
                    "type": "string",
                    "example": "status bad request"
                }
            }
        },
        "models.CreateSecretRequest": {
            "type": "object",
            "properties": {
                "access_limit": {
                    "type": "integer",
                    "example": 10
                },
                "content": {
                    "type": "string",
                    "example": "my very secret text"
                },
                "expiration_epoch": {
                    "type": "integer",
                    "example": 1577836800
                }
            }
        },
        "models.Health": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "example": "Redis"
                },
                "status": {
                    "type": "string",
                    "example": "healthy"
                },
                "version": {
                    "type": "string",
                    "example": "1.0.0"
                }
            }
        },
        "models.HealthResponse": {
            "type": "object",
            "properties": {
                "datastore": {
                    "type": "object",
                    "$ref": "#/definitions/models.Health"
                },
                "encryption": {
                    "type": "object",
                    "$ref": "#/definitions/models.Health"
                },
                "host": {
                    "type": "string",
                    "example": "localhost"
                },
                "status": {
                    "type": "string",
                    "example": "healthy"
                }
            }
        },
        "models.SecretContentResponse": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string",
                    "example": "my very secret text"
                },
                "id": {
                    "type": "string",
                    "example": "22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"
                }
            }
        },
        "models.SecretMetadataResponse": {
            "type": "object",
            "properties": {
                "access_count": {
                    "type": "integer",
                    "example": 1
                },
                "access_limit": {
                    "type": "integer",
                    "example": 10
                },
                "expiration": {
                    "type": "string",
                    "example": "1970-01-01 00:00:00 UTC"
                },
                "id": {
                    "type": "string",
                    "example": "22b6fff1be15d1fd54b7b8ec6ad22e80e66275195c914c4b0f9652248a498680"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "Cellar",
	Description: "Simple secret sharing with the infrastructure you already trust",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
