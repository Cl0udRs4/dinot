# C2系统 API 文档

## 目录
1. [概述](#概述)
2. [认证](#认证)
3. [API端点](#API端点)
4. [客户端管理](#客户端管理)
5. [模块管理](#模块管理)
6. [命令执行](#命令执行)
7. [系统状态](#系统状态)
8. [错误处理](#错误处理)
9. [示例](#示例)

## 概述

C2系统提供了一套RESTful API，用于管理和控制客户端、模块和命令执行。API服务器默认监听在8090端口，支持HTTP和HTTPS协议。

所有API响应均使用JSON格式，状态码遵循HTTP标准：
- 200: 成功
- 400: 请求错误
- 401: 认证失败
- 404: 资源不存在
- 500: 服务器内部错误

## 认证

API支持两种认证方式：

### 基本认证 (Basic Authentication)

使用HTTP基本认证，在请求头中包含`Authorization`字段：

```
Authorization: Basic base64(username:password)
```

示例：
```bash
curl -u admin:password http://localhost:8090/api/status
```

### JWT认证 (JSON Web Token)

在请求头中包含`Authorization`字段，使用Bearer模式：

```
Authorization: Bearer <token>
```

获取JWT令牌：

```
POST /api/auth/token
```

请求体：
```json
{
  "username": "admin",
  "password": "password"
}
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-03-06T13:27:09Z"
}
```

示例：
```bash
# 获取令牌
TOKEN=$(curl -s -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}' http://localhost:8090/api/auth/token | jq -r .token)

# 使用令牌访问API
curl -H "Authorization: Bearer $TOKEN" http://localhost:8090/api/status
```

## API端点

### 系统状态

#### 获取系统状态

```
GET /api/status
```

响应：
```json
{
  "status": "running",
  "uptime": "3h 24m 15s",
  "version": "1.0.0",
  "clients_count": 42,
  "protocols": {
    "tcp": {
      "status": "listening",
      "port": 8080,
      "connections": 15
    },
    "udp": {
      "status": "listening",
      "port": 8081,
      "connections": 10
    },
    "ws": {
      "status": "listening",
      "port": 8082,
      "connections": 8
    },
    "dns": {
      "status": "listening",
      "port": 8053,
      "connections": 5
    },
    "icmp": {
      "status": "listening",
      "connections": 4
    }
  }
}
```

## 客户端管理

### 列出所有客户端

```
GET /api/clients
```

查询参数：
- `limit`: 限制返回的客户端数量 (默认: 100)
- `offset`: 分页偏移量 (默认: 0)
- `protocol`: 按协议筛选 (可选)
- `status`: 按状态筛选 (可选: active, inactive)

响应：
```json
{
  "clients": [
    {
      "id": "client-123456",
      "name": "Client-123456",
      "ip": "192.168.1.100",
      "os": "Windows 10",
      "arch": "x64",
      "protocol": "tcp",
      "last_seen": "2025-03-05T13:15:22Z",
      "status": "active",
      "modules": ["shell", "system", "file"]
    },
    {
      "id": "client-789012",
      "name": "Client-789012",
      "ip": "192.168.1.101",
      "os": "Ubuntu 20.04",
      "arch": "x64",
      "protocol": "udp",
      "last_seen": "2025-03-05T13:10:45Z",
      "status": "active",
      "modules": ["shell", "system"]
    }
  ],
  "total": 42,
  "limit": 100,
  "offset": 0
}
```

### 获取客户端详情

```
GET /api/clients/{client_id}
```

响应：
```json
{
  "id": "client-123456",
  "name": "Client-123456",
  "ip": "192.168.1.100",
  "os": "Windows 10",
  "arch": "x64",
  "protocol": "tcp",
  "last_seen": "2025-03-05T13:15:22Z",
  "status": "active",
  "modules": ["shell", "system", "file"],
  "heartbeat": {
    "interval": 10,
    "last_heartbeat": "2025-03-05T13:15:22Z",
    "failures": 0
  },
  "system_info": {
    "hostname": "DESKTOP-ABC123",
    "username": "user",
    "cpu": "Intel Core i7-10700K",
    "memory_total": 16384,
    "memory_used": 8192,
    "disk_total": 512000,
    "disk_used": 256000
  }
}
```

### 向客户端发送命令

```
POST /api/clients/{client_id}/command
```

请求体：
```json
{
  "command": "whoami",
  "timeout": 30
}
```

响应：
```json
{
  "id": "cmd-123456",
  "client_id": "client-123456",
  "command": "whoami",
  "status": "sent",
  "sent_at": "2025-03-05T13:20:15Z"
}
```

### 获取命令执行结果

```
GET /api/clients/{client_id}/command/{command_id}
```

响应：
```json
{
  "id": "cmd-123456",
  "client_id": "client-123456",
  "command": "whoami",
  "status": "completed",
  "sent_at": "2025-03-05T13:20:15Z",
  "completed_at": "2025-03-05T13:20:16Z",
  "exit_code": 0,
  "output": "administrator\n"
}
```

## 模块管理

### 列出所有可用模块

```
GET /api/modules
```

响应：
```json
{
  "modules": [
    {
      "name": "shell",
      "description": "执行shell命令",
      "version": "1.0.0",
      "author": "C2 Team",
      "commands": ["execute", "interactive"]
    },
    {
      "name": "system",
      "description": "系统信息收集",
      "version": "1.0.0",
      "author": "C2 Team",
      "commands": ["info", "processes", "users"]
    },
    {
      "name": "file",
      "description": "文件操作",
      "version": "1.0.0",
      "author": "C2 Team",
      "commands": ["list", "read", "write", "delete", "download", "upload"]
    }
  ]
}
```

### 获取模块详情

```
GET /api/modules/{module_name}
```

响应：
```json
{
  "name": "shell",
  "description": "执行shell命令",
  "version": "1.0.0",
  "author": "C2 Team",
  "commands": [
    {
      "name": "execute",
      "description": "执行单条shell命令",
      "usage": "execute <command>",
      "parameters": [
        {
          "name": "command",
          "type": "string",
          "required": true,
          "description": "要执行的命令"
        }
      ]
    },
    {
      "name": "interactive",
      "description": "启动交互式shell会话",
      "usage": "interactive",
      "parameters": []
    }
  ]
}
```

### 向客户端加载模块

```
POST /api/clients/{client_id}/modules
```

请求体：
```json
{
  "module": "file"
}
```

响应：
```json
{
  "client_id": "client-123456",
  "module": "file",
  "status": "loaded",
  "loaded_at": "2025-03-05T13:25:30Z"
}
```

### 从客户端卸载模块

```
DELETE /api/clients/{client_id}/modules/{module_name}
```

响应：
```json
{
  "client_id": "client-123456",
  "module": "file",
  "status": "unloaded",
  "unloaded_at": "2025-03-05T13:30:45Z"
}
```

### 执行模块命令

```
POST /api/clients/{client_id}/modules/{module_name}/execute
```

请求体：
```json
{
  "command": "list",
  "parameters": {
    "path": "/home/user/documents"
  },
  "timeout": 30
}
```

响应：
```json
{
  "id": "mod-123456",
  "client_id": "client-123456",
  "module": "file",
  "command": "list",
  "parameters": {
    "path": "/home/user/documents"
  },
  "status": "sent",
  "sent_at": "2025-03-05T13:35:10Z"
}
```

### 获取模块命令执行结果

```
GET /api/clients/{client_id}/modules/{module_name}/execute/{execution_id}
```

响应：
```json
{
  "id": "mod-123456",
  "client_id": "client-123456",
  "module": "file",
  "command": "list",
  "parameters": {
    "path": "/home/user/documents"
  },
  "status": "completed",
  "sent_at": "2025-03-05T13:35:10Z",
  "completed_at": "2025-03-05T13:35:11Z",
  "exit_code": 0,
  "result": {
    "files": [
      {
        "name": "document1.txt",
        "type": "file",
        "size": 1024,
        "modified": "2025-03-01T10:15:30Z"
      },
      {
        "name": "document2.pdf",
        "type": "file",
        "size": 102400,
        "modified": "2025-03-02T14:20:45Z"
      },
      {
        "name": "subfolder",
        "type": "directory",
        "modified": "2025-03-03T09:30:15Z"
      }
    ]
  }
}
```

## 系统状态

### 获取系统资源使用情况

```
GET /api/system/resources
```

响应：
```json
{
  "cpu": {
    "usage": 35.2,
    "cores": 8
  },
  "memory": {
    "total": 16384,
    "used": 8192,
    "free": 8192,
    "usage": 50.0
  },
  "disk": {
    "total": 512000,
    "used": 256000,
    "free": 256000,
    "usage": 50.0
  },
  "network": {
    "rx_bytes": 1024000,
    "tx_bytes": 512000,
    "connections": 42
  }
}
```

### 获取系统日志

```
GET /api/system/logs
```

查询参数：
- `level`: 日志级别 (debug, info, warn, error)
- `limit`: 限制返回的日志条数 (默认: 100)
- `offset`: 分页偏移量 (默认: 0)
- `from`: 开始时间 (ISO 8601格式)
- `to`: 结束时间 (ISO 8601格式)

响应：
```json
{
  "logs": [
    {
      "timestamp": "2025-03-05T13:00:00Z",
      "level": "info",
      "message": "Server started",
      "source": "server"
    },
    {
      "timestamp": "2025-03-05T13:01:15Z",
      "level": "info",
      "message": "Client connected: client-123456",
      "source": "tcp_listener"
    },
    {
      "timestamp": "2025-03-05T13:05:30Z",
      "level": "warn",
      "message": "Client heartbeat missed: client-789012",
      "source": "heartbeat_monitor"
    }
  ],
  "total": 1542,
  "limit": 100,
  "offset": 0
}
```

## 错误处理

所有API错误响应均使用以下格式：

```json
{
  "error": {
    "code": "client_not_found",
    "message": "Client with ID client-999999 not found",
    "details": {
      "client_id": "client-999999"
    }
  }
}
```

常见错误代码：

| 错误代码 | 描述 |
|---------|------|
| invalid_request | 请求格式或参数无效 |
| authentication_failed | 认证失败 |
| authorization_failed | 授权失败 |
| client_not_found | 客户端不存在 |
| module_not_found | 模块不存在 |
| command_not_found | 命令不存在 |
| execution_failed | 命令执行失败 |
| timeout | 操作超时 |
| internal_error | 服务器内部错误 |

## 示例

### 使用curl进行API调用

#### 获取系统状态

```bash
curl -u admin:password http://localhost:8090/api/status
```

#### 列出所有客户端

```bash
curl -u admin:password http://localhost:8090/api/clients
```

#### 向客户端发送命令

```bash
curl -u admin:password -X POST -H "Content-Type: application/json" -d '{"command":"whoami","timeout":30}' http://localhost:8090/api/clients/client-123456/command
```

#### 加载模块并执行命令

```bash
# 加载模块
curl -u admin:password -X POST -H "Content-Type: application/json" -d '{"module":"file"}' http://localhost:8090/api/clients/client-123456/modules

# 执行模块命令
curl -u admin:password -X POST -H "Content-Type: application/json" -d '{"command":"list","parameters":{"path":"/home/user/documents"},"timeout":30}' http://localhost:8090/api/clients/client-123456/modules/file/execute
```

### 使用Python进行API调用

```python
import requests
import json

# 配置
API_URL = "http://localhost:8090/api"
USERNAME = "admin"
PASSWORD = "password"

# 获取系统状态
response = requests.get(f"{API_URL}/status", auth=(USERNAME, PASSWORD))
status = response.json()
print(f"系统状态: {status['status']}")
print(f"在线客户端数: {status['clients_count']}")

# 列出所有客户端
response = requests.get(f"{API_URL}/clients", auth=(USERNAME, PASSWORD))
clients = response.json()["clients"]
print(f"客户端列表:")
for client in clients:
    print(f"  - {client['id']} ({client['name']}): {client['status']}")

# 向客户端发送命令
client_id = "client-123456"
command_data = {
    "command": "whoami",
    "timeout": 30
}
response = requests.post(
    f"{API_URL}/clients/{client_id}/command",
    auth=(USERNAME, PASSWORD),
    json=command_data
)
command = response.json()
print(f"命令已发送: {command['id']}")

# 获取命令执行结果
command_id = command["id"]
response = requests.get(
    f"{API_URL}/clients/{client_id}/command/{command_id}",
    auth=(USERNAME, PASSWORD)
)
result = response.json()
print(f"命令执行结果: {result['output']}")
```

### 使用Go进行API调用

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	APIURL    = "http://localhost:8090/api"
	USERNAME  = "admin"
	PASSWORD  = "password"
	CLIENT_ID = "client-123456"
)

func main() {
	// 获取系统状态
	req, _ := http.NewRequest("GET", APIURL+"/status", nil)
	req.SetBasicAuth(USERNAME, PASSWORD)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var status map[string]interface{}
	json.Unmarshal(body, &status)
	
	fmt.Printf("系统状态: %v\n", status["status"])
	fmt.Printf("在线客户端数: %v\n", status["clients_count"])
	
	// 向客户端发送命令
	commandData := map[string]interface{}{
		"command": "whoami",
		"timeout": 30,
	}
	jsonData, _ := json.Marshal(commandData)
	
	req, _ = http.NewRequest(
		"POST",
		fmt.Sprintf("%s/clients/%s/command", APIURL, CLIENT_ID),
		bytes.NewBuffer(jsonData),
	)
	req.SetBasicAuth(USERNAME, PASSWORD)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ = ioutil.ReadAll(resp.Body)
	var command map[string]interface{}
	json.Unmarshal(body, &command)
	
	fmt.Printf("命令已发送: %v\n", command["id"])
}
```

## 安全建议

1. **始终使用HTTPS**
   - 在生产环境中配置TLS/SSL
   - 使用有效的证书

2. **实施强认证**
   - 使用复杂密码
   - 启用JWT认证并设置合理的过期时间
   - 考虑实施双因素认证

3. **限制API访问**
   - 使用防火墙限制API端口访问
   - 实施IP白名单
   - 考虑使用VPN或SSH隧道访问API

4. **监控API使用**
   - 记录所有API访问
   - 设置异常访问告警
   - 定期审查访问日志
