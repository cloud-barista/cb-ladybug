// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

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
            "name": "API Support",
            "url": "http://cloud-barista.github.io",
            "email": "contact-to-cloud-barista@googlegroups.com"
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
        "/healthy": {
            "get": {
                "description": "for health check",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Default"
                ],
                "summary": "Health Check",
                "operationId": "Healthy",
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/ns/{namespace}/clusters": {
            "get": {
                "description": "List all Clusters",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Cluster"
                ],
                "summary": "List all Clusters",
                "operationId": "ListCluster",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.ClusterList"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            },
            "post": {
                "description": "Create Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Cluster"
                ],
                "summary": "Create Cluster",
                "operationId": "CreateCluster",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Request Body to create cluster",
                        "name": "ClusterReq",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.ClusterReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Cluster"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            }
        },
        "/ns/{namespace}/clusters/{cluster}": {
            "get": {
                "description": "Get Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Cluster"
                ],
                "summary": "Get Cluster",
                "operationId": "GetCluster",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Cluster Name",
                        "name": "cluster",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Cluster"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Cluster"
                ],
                "summary": "Delete Cluster",
                "operationId": "DeleteCluster",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Cluster Name",
                        "name": "cluster",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Status"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            }
        },
        "/ns/{namespace}/clusters/{cluster}/nodes": {
            "get": {
                "description": "List all Nodes in specified Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Node"
                ],
                "summary": "List all Nodes in specified Cluster",
                "operationId": "ListNode",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Cluster Name",
                        "name": "cluster",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.NodeList"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            },
            "post": {
                "description": "Add Node in specified Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Node"
                ],
                "summary": "Add Node in specified Cluster",
                "operationId": "AddNode",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Cluster Name",
                        "name": "cluster",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Request Body to add node",
                        "name": "nodeReq",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.NodeReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Node"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            }
        },
        "/ns/{namespace}/clusters/{cluster}/nodes/{node}": {
            "get": {
                "description": "Get Node in specified Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Node"
                ],
                "summary": "Get Node in specified Cluster",
                "operationId": "GetNode",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Cluster Name",
                        "name": "cluster",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Node Name",
                        "name": "node",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Node"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            },
            "delete": {
                "description": "Remove Node in specified Cluster",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Node"
                ],
                "summary": "Remove Node in specified Cluster",
                "operationId": "RemoveNode",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Namespace ID",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Cluster Name",
                        "name": "cluster",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Node Name",
                        "name": "node",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Status"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/app.Status"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "app.Status": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "example": "Any message"
                }
            }
        },
        "model.Cluster": {
            "type": "object",
            "properties": {
                "clusterConfig": {
                    "type": "string"
                },
                "cpLeader": {
                    "type": "string"
                },
                "createdTime": {
                    "type": "string",
                    "example": "2022-01-02T12:00:00Z"
                },
                "description": {
                    "type": "string"
                },
                "kind": {
                    "type": "string"
                },
                "label": {
                    "type": "string"
                },
                "mcis": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "namespace": {
                    "type": "string"
                },
                "networkCni": {
                    "type": "string",
                    "enum": [
                        "canal",
                        "kilo"
                    ]
                },
                "nodes": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.Node"
                    }
                },
                "status": {
                    "type": "object",
                    "properties": {
                        "message": {
                            "type": "string"
                        },
                        "phase": {
                            "type": "string",
                            "enum": [
                                "Pending",
                                "Provisioning",
                                "Provisioned",
                                "Failed"
                            ]
                        },
                        "reason": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "model.ClusterList": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.Cluster"
                    }
                },
                "kind": {
                    "type": "string"
                }
            }
        },
        "model.ClusterReq": {
            "type": "object",
            "properties": {
                "config": {
                    "type": "object",
                    "$ref": "#/definitions/model.Config"
                },
                "controlPlane": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.NodeConfig"
                    }
                },
                "description": {
                    "type": "string"
                },
                "label": {
                    "type": "string"
                },
                "name": {
                    "type": "string",
                    "example": "cluster-01"
                },
                "worker": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.NodeConfig"
                    }
                }
            }
        },
        "model.Config": {
            "type": "object",
            "properties": {
                "kubernetes": {
                    "type": "object",
                    "$ref": "#/definitions/model.Kubernetes"
                }
            }
        },
        "model.Kubernetes": {
            "type": "object",
            "properties": {
                "networkCni": {
                    "type": "string",
                    "enum": [
                        "canal",
                        "kilo"
                    ],
                    "example": "canal"
                },
                "podCidr": {
                    "type": "string",
                    "example": "10.244.0.0/16"
                },
                "serviceCidr": {
                    "type": "string",
                    "example": "10.96.0.0/12"
                },
                "serviceDnsDomain": {
                    "type": "string",
                    "example": "cluster.local"
                }
            }
        },
        "model.Node": {
            "type": "object",
            "properties": {
                "createdTime": {
                    "type": "string",
                    "example": "2022-01-02T12:00:00Z"
                },
                "credential": {
                    "type": "string"
                },
                "csp": {
                    "type": "string",
                    "enum": [
                        "aws",
                        "gcp",
                        "azure",
                        "alibaba",
                        "tencent",
                        "openstack"
                    ]
                },
                "cspLabel": {
                    "type": "string"
                },
                "kind": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "publicIp": {
                    "type": "string"
                },
                "regionLabel": {
                    "type": "string"
                },
                "role": {
                    "type": "string",
                    "enum": [
                        "control-plane",
                        "worker"
                    ]
                },
                "spec": {
                    "type": "string"
                },
                "zoneLabel": {
                    "type": "string"
                }
            }
        },
        "model.NodeConfig": {
            "type": "object",
            "properties": {
                "connection": {
                    "type": "string",
                    "example": "config-aws-ap-northeast-2"
                },
                "count": {
                    "type": "integer",
                    "example": 3
                },
                "spec": {
                    "type": "string",
                    "example": "t2.medium"
                }
            }
        },
        "model.NodeList": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.Node"
                    }
                },
                "kind": {
                    "type": "string"
                }
            }
        },
        "model.NodeReq": {
            "type": "object",
            "properties": {
                "controlPlane": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.NodeConfig"
                    }
                },
                "worker": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.NodeConfig"
                    }
                }
            }
        },
        "model.Status": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "kind": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
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
	Version:     "latest",
	Host:        "localhost:1470",
	BasePath:    "/mcks",
	Schemes:     []string{},
	Title:       "CB-MCKS REST API",
	Description: "CB-MCKS REST API",
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
