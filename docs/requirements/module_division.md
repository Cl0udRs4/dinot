# C2 系统模块划分与接口设计

## 1. 模块划分

### 1.1 Server模块
- **监听器模块**：负责多协议监听和初始连接处理
- **客户端管理模块**：负责客户端注册、状态更新和心跳检测
- **控制接口模块**：提供Console和HTTP API两种控制方式
- **日志与监控模块**：负责日志记录和异常报警
- **加密模块**：处理通信加密与解密
- **模块管理模块**：负责模块下发、调用和卸载

### 1.2 Client模块
- **通信模块**：实现多协议支持和自动切换
- **加密模块**：处理通信加密与解密
- **心跳模块**：维持与Server的连接
- **模块加载器**：负责动态加载、调用和卸载功能模块
- **反馈模块**：处理指令执行结果的反馈

### 1.3 Builder工具
- **参数解析模块**：处理命令行参数
- **代码生成模块**：根据参数生成定制化客户端代码
- **编译打包模块**：编译生成最终可执行文件
- **签名校验模块**：确保生成文件的完整性与安全性

## 2. 接口设计

### 2.1 Server接口

#### 2.1.1 监听器接口
```go
// Listener 接口定义了所有协议监听器需要实现的方法
type Listener interface {
    // Start 启动监听器
    Start() error
    
    // Stop 停止监听器
    Stop() error
    
    // GetProtocol 获取监听器协议类型
    GetProtocol() string
    
    // GetStatus 获取监听器状态
    GetStatus() string
}
```

#### 2.1.2 客户端管理接口
```go
// ClientManager 接口定义了客户端管理模块需要实现的方法
type ClientManager interface {
    // RegisterClient 注册新客户端
    RegisterClient(client Client) error
    
    // UpdateClientStatus 更新客户端状态
    UpdateClientStatus(clientID string, status string) error
    
    // HeartbeatCheck 心跳检测
    HeartbeatCheck(clientID string) error
    
    // GetClient 获取客户端信息
    GetClient(clientID string) (Client, error)
    
    // GetAllClients 获取所有客户端信息
    GetAllClients() []Client
}
```

#### 2.1.3 控制接口
```go
// Controller 接口定义了控制接口需要实现的方法
type Controller interface {
    // Start 启动控制接口
    Start() error
    
    // Stop 停止控制接口
    Stop() error
    
    // ExecuteCommand 执行命令
    ExecuteCommand(command string, args []string) (string, error)
}
```

### 2.2 Client接口

#### 2.2.1 通信接口
```go
// Communicator 接口定义了通信模块需要实现的方法
type Communicator interface {
    // Connect 连接到服务器
    Connect() error
    
    // Disconnect 断开连接
    Disconnect() error
    
    // Send 发送数据
    Send(data []byte) error
    
    // Receive 接收数据
    Receive() ([]byte, error)
    
    // SwitchProtocol 切换协议
    SwitchProtocol(protocol string) error
}
```

#### 2.2.2 模块接口
```go
// Module 接口定义了功能模块需要实现的方法
type Module interface {
    // Init 初始化模块
    Init() error
    
    // Execute 执行模块功能
    Execute(command string, args []string) (string, error)
    
    // Cleanup 清理模块资源
    Cleanup() error
    
    // GetInfo 获取模块信息
    GetInfo() ModuleInfo
}
```

### 2.3 Builder接口

#### 2.3.1 代码生成接口
```go
// CodeGenerator 接口定义了代码生成模块需要实现的方法
type CodeGenerator interface {
    // GenerateCode 生成客户端代码
    GenerateCode(config BuilderConfig) error
    
    // Compile 编译生成可执行文件
    Compile() error
    
    // Sign 对生成的文件进行签名
    Sign() error
}
```
