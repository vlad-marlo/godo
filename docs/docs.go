// Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/ping": {
            "get": {
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Ping"
                ],
                "summary": "Запрос состояния сервиса",
                "operationId": "ping",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users/register": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "UserCreate"
                ],
                "summary": "Создание пользователя",
                "operationId": "user_create",
                "parameters": [
                    {
                        "description": "User data",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.RegisterUserRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/model.Error"
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "schema": {
                            "$ref": "#/definitions/model.Error"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.Error"
                        }
                    }
                }
            }
        },
        "/users/token": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "CreateToken"
                ],
                "summary": "Создание JWT токена для пользователя.",
                "operationId": "login_jwt",
                "parameters": [
                    {
                        "description": "User data",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.CreateTokenRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/model.CreateTokenResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/model.Error"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.CreateGroupRequest": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "model.CreateTokenRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "description": "Email is user email",
                    "type": "string",
                    "example": "user@example.com"
                },
                "password": {
                    "description": "Password is password of user.",
                    "type": "string",
                    "example": "strong_password"
                },
                "token-type": {
                    "type": "string",
                    "example": "bearer"
                }
            }
        },
        "model.CreateTokenResponse": {
            "type": "object",
            "properties": {
                "access_token": {
                    "type": "string"
                },
                "token_type": {
                    "type": "string"
                }
            }
        },
        "model.Error": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string",
                    "example": "short summary info about error"
                },
                "field": {
                    "type": "string",
                    "example": "additional info about error"
                }
            }
        },
        "model.RegisterUserRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "description": "Email is user email",
                    "type": "string",
                    "example": "user@example.com"
                },
                "password": {
                    "description": "Password is password string",
                    "type": "string",
                    "example": "strong_password"
                }
            }
        },
        "model.User": {
            "type": "object",
            "properties": {
                "email": {
                    "description": "Email is string field of user's email addr.",
                    "type": "string",
                    "example": "user@example.com"
                },
                "id": {
                    "description": "ID is user uuid.",
                    "type": "string",
                    "example": "00000000-0000-0000-0000-000000000000"
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "GODO API",
	Description:      "This is a godo server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}