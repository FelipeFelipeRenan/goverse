// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "Felipe Renan",
            "email": "feliperenanqwerty@gmail.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Realiza login com email e senha",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Login",
                "parameters": [
                    {
                        "description": "Credenciais de login",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.LoginResponse"
                        }
                    },
                    "401": {
                        "description": "Credenciais inválidas",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Listar salas",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/swagger.RoomResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Criar nova sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Dados da nova sala",
                        "name": "room",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.CreateRoomRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/swagger.RoomResponse"
                        }
                    },
                    "400": {
                        "description": "Dados inválidos",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms/mine": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retorna todas as salas onde o usuário autenticado é o proprietário (owner_id)",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Listar salas criadas pelo usuário autenticado",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Lista de salas",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/swagger.RoomResponse"
                            }
                        }
                    },
                    "401": {
                        "description": "Não autorizado",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Erro interno no servidor",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Buscar sala por ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.RoomResponse"
                        }
                    },
                    "404": {
                        "description": "Sala não encontrada",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Excluir sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Sala excluída",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Sala não encontrada",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Atualizar sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Dados atualizados da sala",
                        "name": "room",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.UpdateRoomRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.RoomResponse"
                        }
                    },
                    "400": {
                        "description": "Dados inválidos",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms/{roomID}/join": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "Member"
                ],
                "summary": "Entrar em uma sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "roomID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.MemberResponse"
                        }
                    },
                    "403": {
                        "description": "Acesso negado",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms/{roomID}/members": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "Member"
                ],
                "summary": "Listar membros da sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "roomID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/swagger.MemberResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Member"
                ],
                "summary": "Adicionar membro à sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "roomID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Dados do membro",
                        "name": "member",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.AddMemberRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/swagger.MemberResponse"
                        }
                    },
                    "400": {
                        "description": "Dados inválidos",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms/{roomID}/members/{memberID}": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "tags": [
                    "Member"
                ],
                "summary": "Remover membro da sala",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "roomID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID do membro",
                        "name": "memberID",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Removido com sucesso",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rooms/{roomID}/members/{memberID}/role": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Member"
                ],
                "summary": "Atualizar papel do membro",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID da sala",
                        "name": "roomID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID do membro",
                        "name": "memberID",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Novo papel",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.UpdateRoleRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.MemberResponse"
                        }
                    }
                }
            }
        },
        "/user": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Criar novo usuário",
                "parameters": [
                    {
                        "description": "Dados do novo usuário",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.CreateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/swagger.UserResponse"
                        }
                    },
                    "400": {
                        "description": "Dados inválidos",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/me": {
            "put": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Usuário"
                ],
                "summary": "Atualiza os dados do usuário autenticado",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "description": "Dados do usuário",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/swagger.UpdateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.UserResponse"
                        }
                    },
                    "400": {
                        "description": "Requisição inválida",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "401": {
                        "description": "Não autorizado",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "BearerAuth": []
                    }
                ],
                "tags": [
                    "Usuário"
                ],
                "summary": "Remove (soft delete) o usuário autenticado",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Usuário deletado com sucesso",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "401": {
                        "description": "Não autorizado",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/rooms": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retorna todas as salas das quais o usuário participa",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Room"
                ],
                "summary": "Listar salas na qual um usuário é membro",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/swagger.RoomResponse"
                            }
                        }
                    },
                    "401": {
                        "description": "Não autorizado",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Erro interno",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Obter usuário por ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Token de autenticação (Bearer token)",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID do usuário",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/swagger.UserResponse"
                        }
                    },
                    "401": {
                        "description": "Não autorizado",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Listar todos os usuários",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/swagger.UserResponse"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "swagger.AddMemberRequest": {
            "type": "object",
            "required": [
                "user_id"
            ],
            "properties": {
                "role": {
                    "type": "string",
                    "example": "member"
                },
                "user_id": {
                    "type": "integer",
                    "example": 3
                }
            }
        },
        "swagger.CreateRoomRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "description": {
                    "type": "string",
                    "example": "Sala para estudo de algoritmos"
                },
                "name": {
                    "type": "string",
                    "example": "Sala de Estudos"
                }
            }
        },
        "swagger.CreateUserRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "swagger.LoginRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string",
                    "example": "joao@email.com"
                },
                "password": {
                    "type": "string",
                    "example": "senha123"
                },
                "type": {
                    "type": "string",
                    "example": "password"
                }
            }
        },
        "swagger.LoginResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "swagger.MemberResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "example": 5
                },
                "joined_at": {
                    "type": "string",
                    "example": "2025-06-05T20:00:00Z"
                },
                "role": {
                    "type": "string",
                    "example": "admin"
                },
                "room_id": {
                    "type": "integer",
                    "example": 1
                },
                "user_id": {
                    "type": "integer",
                    "example": 2
                },
                "username": {
                    "type": "string",
                    "example": "joaogate"
                }
            }
        },
        "swagger.RoomResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string",
                    "example": "2025-06-06T18:30:00Z"
                },
                "description": {
                    "type": "string",
                    "example": "Sala para estudo de algoritmos"
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "name": {
                    "type": "string",
                    "example": "Sala de Estudos"
                },
                "owner_id": {
                    "type": "integer",
                    "example": 1
                },
                "updated_at": {
                    "type": "string",
                    "example": "2025-06-06T19:00:00Z"
                }
            }
        },
        "swagger.UpdateRoleRequest": {
            "type": "object",
            "required": [
                "role"
            ],
            "properties": {
                "role": {
                    "type": "string",
                    "example": "admin"
                }
            }
        },
        "swagger.UpdateRoomRequest": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string",
                    "example": "Descrição atualizada"
                },
                "name": {
                    "type": "string",
                    "example": "Sala Atualizada"
                }
            }
        },
        "swagger.UpdateUserRequest": {
            "type": "object",
            "properties": {
                "picture": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "swagger.UserResponse": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string",
                    "example": "joao@email.com"
                },
                "id": {
                    "type": "integer",
                    "example": 1
                },
                "picture": {
                    "type": "string"
                },
                "username": {
                    "type": "string",
                    "example": "joaogate"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8088",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Goverse API GAteway",
	Description:      "Documentação unificada dos serviços do Goverse",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
