# C2 系统开发进度记录

## 阶段 1：需求调研与架构设计

### 1.1 需求调研
- [x] 完成需求调研
  - 已创建 [需求分析文档](requirements/requirements_analysis.md)，包含业务场景、用户故事、功能模块及核心特性描述
- [x] 业务场景分析
  - 已分析红队测试、渗透测试与对抗演练三大业务场景
- [x] 模块划分
  - 已创建 [模块划分文档](requirements/module_division.md)，包含初步模块划分和接口设计

### 1.2 架构设计
- [ ] 接口设计
- [ ] 详细架构图绘制
## 阶段 2：核心模块开发与单元测试

### 2.1 Server 模块开发

#### 2.1.1 多协议监听器实现
- [x] 实现对 TCP、UDP、WS、ICMP、DNS 协议的监听，每个协议使用独立 goroutine 处理
- [x] 支持灵活配置监听端口和参数
- [x] 编写单元测试（go test）和编译测试（go build）

### 任务2.1.1：多协议监听器实现
- **完成时间**：2025-03-04
- **主要内容**：
  1. 实现TCP、UDP、WebSocket、ICMP、DNS五种协议的监听器
  2. 每个协议使用独立goroutine处理，支持灵活配置监听端口和参数
  3. 编写单元测试和编译测试，确保代码质量
- **配置参数**：
  - 通用配置：
    - Address: 监听地址和端口
    - BufferSize: 缓冲区大小
    - MaxConnections: 最大连接数
    - Timeout: 超时时间（秒）
  - 协议特定配置：
    - DNS: 支持域名解析和记录类型配置
    - WebSocket: 支持路径和TLS配置
    - ICMP: 支持ICMP类型和代码配置
- **代码实现**：
  - 基于接口设计，所有监听器实现统一的Listener接口
  - 使用BaseListener作为基础实现，各协议监听器继承并扩展
  - 实现ListenerManager统一管理所有监听器的启动、停止和状态
  - 使用context控制监听器生命周期，确保优雅关闭
  - 使用goroutine和channel实现异步处理
- **测试结果**：
  - 所有协议监听器单元测试通过
  - 编译测试通过，无编译错误
  - 验证了监听器的启动、停止、状态管理等功能
- **遇到的问题与解决方案**：
  - 问题：DNS监听器实现中存在方法重复定义
  - 解决方案：重构DNS监听器代码，移除重复方法定义
  - 问题：DNS连接包装器实现不完整
  - 解决方案：实现完整的net.Conn接口，包括Read、Write、Close等方法
### 任务2.1.1：多协议监听器测试增强
- **完成时间**：2025-03-05
- **主要内容**：
  1. 增强所有协议监听器的测试覆盖率
  2. 为每个协议监听器添加Start和Stop功能的全面测试
  3. 修复测试过程中发现的问题
- **测试增强**：
  - TCP监听器：添加Start和Stop功能的全面测试，验证连接建立和关闭
  - UDP监听器：添加Start和Stop功能的全面测试，验证数据包收发
  - WebSocket监听器：添加Start和Stop功能的全面测试，修复端口配置问题
  - ICMP监听器：添加Start和Stop功能的全面测试，添加root权限检查
  - DNS监听器：添加Start和Stop功能的全面测试，使用非标准端口避免冲突
- **测试结果**：
  - TCP监听器：所有测试通过
  - UDP监听器：所有测试通过
  - WebSocket监听器：所有测试通过
  - ICMP监听器：测试被跳过（需要root权限）
  - DNS监听器：所有测试通过
  - 编译测试：所有包编译成功
- **遇到的问题与解决方案**：
  - 问题：WebSocket监听器测试使用随机端口导致连接失败
  - 解决方案：使用固定端口（127.0.0.1:8082）进行测试
  - 问题：DNS连接包装器中的ErrTimeout错误定义重复
  - 解决方案：统一使用一个ErrTimeout定义
  - 问题：ICMP测试需要root权限
  - 解决方案：添加权限检查，非root环境下跳过测试
- **下一步计划**：
  - 进行任务2.1.2：客户端管理模块实现
  - 包括客户端注册、状态更新、心跳检测和异常上报功能

#### 2.1.2 客户端管理模块实现
- [x] 实现客户端注册功能，记录客户端信息和支持模块
- [x] 实现客户端状态更新功能，支持在线、离线、忙碌和错误状态
- [x] 实现心跳检测功能，支持1s～24h随机延时与超时切换
- [ ] 实现异常上报功能

<<<<<<< Updated upstream
||||||| constructed merge base
#### 2.1.3 控制接口实现
- [x] 实现Console模式（命令行交互）
- [ ] 实现HTTP API（RESTful接口）

=======
#### 2.1.3 控制接口实现
- [x] 实现Console模式（命令行交互）
- [x] 实现HTTP API（RESTful接口）

>>>>>>> Stashed changes
### 任务2.1.2：客户端管理模块实现
- **完成时间**：2025-03-05
- **主要内容**：
  1. 实现客户端注册功能，记录客户端信息和支持模块
  2. 实现客户端状态更新功能，支持在线、离线、忙碌和错误状态
  3. 实现心跳检测功能，支持1s～24h随机延时与超时切换
- **代码实现**：
  - 客户端模型（Client）：
    - 包含ID、名称、IP地址、操作系统、架构等基本信息
    - 记录注册时间、最后活动时间、当前状态和错误信息
    - 支持模块列表和活动模块列表
    - 心跳间隔配置
    - 并发安全的状态更新方法
  - 客户端管理器（ClientManager）：
    - 支持客户端注册和注销
    - 提供按ID和状态查询客户端的方法
    - 支持批量状态更新和超时检测
    - 并发安全的客户端集合管理
  - 心跳监控器（HeartbeatMonitor）：
    - 定期检查客户端心跳状态
    - 支持配置检查间隔和超时时间
    - 支持启用/禁用随机心跳间隔
    - 支持1s～24h范围内的随机心跳间隔分配
    - 使用context控制监控器生命周期
- **测试内容**：
  - 客户端创建和属性验证
  - 状态更新和最后活动时间更新
  - 心跳间隔设置
  - 模块管理（添加/移除活动模块）
  - 客户端JSON序列化和反序列化
  - 客户端注册和注销
  - 客户端查询（按ID和状态）
  - 超时检测和离线标记
  - 心跳监控器配置和随机间隔分配
  - 心跳处理和状态恢复
- **实现特点**：
  - 使用互斥锁（sync.RWMutex）确保并发安全
  - 使用context控制心跳监控器的生命周期
  - 支持灵活配置心跳间隔和超时时间
  - 实现随机心跳间隔以增强隐蔽性
  - 提供完整的客户端生命周期管理
- **遇到的问题与解决方案**：
  - 问题：并发访问客户端状态可能导致竞态条件
  - 解决方案：使用读写锁（sync.RWMutex）保护客户端状态访问
  - 问题：心跳超时检测需要考虑不同客户端的心跳间隔
  - 解决方案：在检测超时时考虑客户端的个性化心跳间隔
- **下一步计划**：
  - 实现异常上报功能
<<<<<<< Updated upstream
  - 开发控制接口（Console模式）
||||||| constructed merge base
  - 开发控制接口（Console模式和HTTP API）

### 任务2.1.3a：控制接口实现 - Console模式
- **完成时间**：2025-03-05
- **主要内容**：
  1. 实现命令行交互式控制台界面
  2. 支持客户端管理、状态更新和心跳配置等功能
  3. 编写单元测试和编译测试，确保代码质量
- **代码实现**：
  - 控制台界面（Console）：
    - 支持命令注册和执行
    - 提供交互式命令行界面
    - 集成客户端管理器和心跳监控器
    - 支持命令历史和帮助信息
  - 服务器集成（Server）：
    - 整合监听器管理器、客户端管理器和心跳监控器
    - 提供统一的启动和停止接口
    - 支持信号处理和优雅关闭
  - 命令实现：
    - help：显示可用命令列表和使用说明
    - list：列出所有客户端或按状态筛选
    - info：显示客户端详细信息
    - status：设置客户端状态
    - heartbeat：配置心跳设置（检查间隔、超时时间、随机间隔）
    - exit：退出控制台
- **测试内容**：
  - 控制台命令执行和参数解析
  - 服务器创建和组件集成
  - 控制台启动和停止
  - 无效命令和参数处理
  - 客户端列表和信息显示
  - 状态更新和心跳配置
- **实现特点**：
  - 模块化设计，每个命令独立实现
  - 使用bufio.Reader读取用户输入
  - 支持命令参数解析和验证
  - 提供友好的错误提示和帮助信息
  - 支持按状态筛选客户端
  - 支持详细的客户端信息显示
  - 支持灵活的心跳配置
- **调用示例**：
  ```
  > help
  Available commands:
    list       - List all clients or filter by status
      Usage: list [status]
    info       - Show detailed information about a client
      Usage: info <client_id>
    status     - Set a client's status
      Usage: status <client_id> <online|offline|busy|error> [error_message]
    heartbeat  - Configure heartbeat settings
      Usage: heartbeat <check|timeout|random> [args...]
    exit       - Exit the console
      Usage: exit
    help       - Display available commands
      Usage: help
  
  > list
  All clients:
  ID                                   IP Address      Status     Last Seen      
  --------------------------------------------------------------------------------
  test-client-id                       192.168.1.100   online     0s ago
  
  Total: 1 clients
  
  > info test-client-id
  Client Information:
    ID:              test-client-id
    Name:            Test Client
    IP Address:      192.168.1.100
    OS:              Linux
    Architecture:    x86_64
    Status:          online
    Protocol:        tcp
    Registered At:   2025-03-05T04:14:31Z
    Last Seen:       2025-03-05T04:14:31Z (0s ago)
    Heartbeat:       1m0s
    Supported Modules:
      - shell
      - file
      - process
    Active Modules:
      None
  
  > status test-client-id busy
  Updated client test-client-id status to busy
  
  > heartbeat check 45
  Set heartbeat check interval to 45s
  
  > heartbeat random enable 5 300
  Enabled random heartbeat intervals (5s - 5m0s)
  ```
- **遇到的问题与解决方案**：
  - 问题：控制台测试中的输入模拟实现复杂
  - 解决方案：使用io.Pipe创建模拟输入流，简化测试代码
  - 问题：客户端状态比较逻辑错误
  - 解决方案：将客户端状态比较从枚举类型改为字符串比较
  - 问题：ListenerManager接口方法名不匹配
  - 解决方案：将StopAll方法调用更改为HaltAll以匹配实际接口
- **下一步计划**：
  - 实现HTTP API控制接口
  - 完善异常上报功能
=======
  - 开发控制接口（Console模式和HTTP API）

### 任务2.1.3a：控制接口实现 - Console模式
- **完成时间**：2025-03-05
- **主要内容**：
  1. 实现命令行交互式控制台界面
  2. 支持客户端管理、状态更新和心跳配置等功能
  3. 编写单元测试和编译测试，确保代码质量
- **代码实现**：
  - 控制台界面（Console）：
    - 支持命令注册和执行
    - 提供交互式命令行界面
    - 集成客户端管理器和心跳监控器
    - 支持命令历史和帮助信息
  - 服务器集成（Server）：
    - 整合监听器管理器、客户端管理器和心跳监控器
    - 提供统一的启动和停止接口
    - 支持信号处理和优雅关闭
  - 命令实现：
    - help：显示可用命令列表和使用说明
    - list：列出所有客户端或按状态筛选
    - info：显示客户端详细信息
    - status：设置客户端状态
    - heartbeat：配置心跳设置（检查间隔、超时时间、随机间隔）
    - exit：退出控制台
- **测试内容**：
  - 控制台命令执行和参数解析
  - 服务器创建和组件集成
  - 控制台启动和停止
  - 无效命令和参数处理
  - 客户端列表和信息显示
  - 状态更新和心跳配置
- **实现特点**：
  - 模块化设计，每个命令独立实现
  - 使用bufio.Reader读取用户输入
  - 支持命令参数解析和验证
  - 提供友好的错误提示和帮助信息
  - 支持按状态筛选客户端
  - 支持详细的客户端信息显示
  - 支持灵活的心跳配置
- **调用示例**：
  ```
  > help
  Available commands:
    list       - List all clients or filter by status
      Usage: list [status]
    info       - Show detailed information about a client
      Usage: info <client_id>
    status     - Set a client's status
      Usage: status <client_id> <online|offline|busy|error> [error_message]
    heartbeat  - Configure heartbeat settings
      Usage: heartbeat <check|timeout|random> [args...]
    exit       - Exit the console
      Usage: exit
    help       - Display available commands
      Usage: help
  
  > list
  All clients:
  ID                                   IP Address      Status     Last Seen      
  --------------------------------------------------------------------------------
  test-client-id                       192.168.1.100   online     0s ago
  
  Total: 1 clients
  
  > info test-client-id
  Client Information:
    ID:              test-client-id
    Name:            Test Client
    IP Address:      192.168.1.100
    OS:              Linux
    Architecture:    x86_64
    Status:          online
    Protocol:        tcp
    Registered At:   2025-03-05T04:14:31Z
    Last Seen:       2025-03-05T04:14:31Z (0s ago)
    Heartbeat:       1m0s
    Supported Modules:
      - shell
      - file
      - process
    Active Modules:
      None
  
  > status test-client-id busy
  Updated client test-client-id status to busy
  
  > heartbeat check 45
  Set heartbeat check interval to 45s
  
  > heartbeat random enable 5 300
  Enabled random heartbeat intervals (5s - 5m0s)
  ```
- **遇到的问题与解决方案**：
  - 问题：控制台测试中的输入模拟实现复杂
  - 解决方案：使用io.Pipe创建模拟输入流，简化测试代码
  - 问题：客户端状态比较逻辑错误
  - 解决方案：将客户端状态比较从枚举类型改为字符串比较
  - 问题：ListenerManager接口方法名不匹配
  - 解决方案：将StopAll方法调用更改为HaltAll以匹配实际接口
- **下一步计划**：
  - 实现HTTP API控制接口
  - 完善异常上报功能

### 任务2.1.3b：控制接口实现 - HTTP API
- **完成时间**：2025-03-05
- **主要内容**：
  1. 实现RESTful API接口，支持客户端管理、状态更新和心跳配置
  2. 支持基本认证和JWT认证
  3. 编写单元测试和编译测试，确保代码质量
- **代码实现**：
  - API处理器（APIHandler）：
    - 集成客户端管理器和心跳监控器
    - 支持配置认证方式（Basic Auth和JWT）
    - 提供RESTful API端点
    - 支持JSON格式的请求和响应
  - 服务器集成：
    - 在Server结构中添加API处理器
    - 在单独的goroutine中启动API服务器
    - 支持配置API监听地址和端口
  - API端点实现：
    - /api/clients：获取所有客户端或按状态筛选
    - /api/clients/{id}：获取特定客户端的详细信息
    - /api/status：更新客户端状态
    - /api/heartbeat：获取和更新心跳配置
  - 认证中间件：
    - 支持禁用认证（开发环境）
    - 支持Basic Auth认证
    - 支持JWT认证（预留接口）
- **测试内容**：
  - API端点功能验证
  - 认证中间件测试
  - 客户端管理和心跳配置API测试
  - 错误处理和边界条件测试
- **实现特点**：
  - RESTful设计风格
  - JSON格式的请求和响应
  - 模块化的API端点实现
  - 灵活的认证配置
  - 完整的错误处理
  - 与现有客户端管理和心跳监控模块无缝集成
- **调用示例**：
  ```
  # 获取所有客户端
  GET /api/clients
  Response:
  [
    {
      "id": "test-client-id",
      "name": "Test Client",
      "ip_address": "192.168.1.100",
      "os": "Linux",
      "architecture": "x86_64",
      "registered_at": "2025-03-05T04:44:55Z",
      "last_seen": "2025-03-05T04:44:55Z",
      "status": "online",
      "supported_modules": ["shell", "file", "process"],
      "active_modules": [],
      "protocol": "tcp",
      "heartbeat_interval": 60000000000
    }
  ]
  
  # 获取特定客户端
  GET /api/clients/test-client-id
  Response:
  {
    "id": "test-client-id",
    "name": "Test Client",
    "ip_address": "192.168.1.100",
    "os": "Linux",
    "architecture": "x86_64",
    "registered_at": "2025-03-05T04:44:55Z",
    "last_seen": "2025-03-05T04:44:55Z",
    "status": "online",
    "supported_modules": ["shell", "file", "process"],
    "active_modules": [],
    "protocol": "tcp",
    "heartbeat_interval": 60000000000
  }
  
  # 更新客户端状态
  POST /api/status
  Request:
  {
    "clientId": "test-client-id",
    "status": "busy"
  }
  Response:
  Client status updated
  
  # 获取心跳配置
  GET /api/heartbeat
  Response:
  {
    "checkInterval": 30,
    "timeout": 60,
    "randomEnabled": false,
    "randomMinInterval": 0,
    "randomMaxInterval": 0
  }
  
  # 更新心跳配置
  POST /api/heartbeat
  Request:
  {
    "checkInterval": 45,
    "timeout": 90,
    "randomEnabled": true,
    "randomMinInterval": 5,
    "randomMaxInterval": 300
  }
  Response:
  Heartbeat settings updated
  ```
- **遇到的问题与解决方案**：
  - 问题：API测试中的客户端状态比较错误
  - 解决方案：修正测试代码，确保正确访问客户端状态常量
  - 问题：DNS连接包装器重复定义导致编译错误
  - 解决方案：删除重复的DNS连接包装器实现文件
  - 问题：控制台输入流EOF错误处理不当
  - 解决方案：添加对io.EOF的专门处理，优雅退出控制台
- **下一步计划**：
  - 完善异常上报功能
  - 实现客户端模块管理API
  - 增强API安全性和JWT认证实现
>>>>>>> Stashed changes
