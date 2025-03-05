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
- [x] 实现异常上报功能

#### 2.1.3 控制接口实现
- [x] 实现Console模式（命令行交互）
- [x] 实现HTTP API（RESTful接口）
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
  - 开发控制接口（Console模式和HTTP API）

### 任务2.1.2c：异常上报功能实现
- **完成时间**：2025-03-06
- **主要内容**：
  1. 实现异常上报功能，支持详细的异常信息记录和查询
  2. 提供API接口和控制台命令用于异常上报和查询
  3. 编写单元测试和编译测试，确保代码质量
- **代码实现**：
  - 异常报告模型（ExceptionReport）：
    - 包含ID、客户端ID、时间戳、消息、严重性级别等基本信息
    - 支持可选的模块名称、堆栈跟踪和附加信息
    - 使用JSON序列化支持API交互
  - 异常管理器（ExceptionManager）：
    - 提供异常报告的添加、查询和管理功能
    - 支持按ID和客户端ID查询异常报告
    - 使用并发安全的数据结构确保线程安全
  - 客户端管理器集成：
    - 扩展ClientManager以支持异常上报和查询
    - 实现ReportException方法用于创建新的异常报告
    - 实现GetExceptionReports和GetAllExceptionReports方法用于查询异常
    - 根据异常严重性自动更新客户端状态
  - 控制台集成：
    - 添加exception命令，支持list和report子命令
    - 支持按客户端ID筛选异常报告
    - 支持设置异常严重性级别（info、warning、error、critical）
  - API集成：
    - 添加/api/exceptions端点用于获取和报告异常
    - 添加/api/exceptions/{id}端点用于获取特定异常详情
    - 支持按客户端ID筛选异常报告
- **测试内容**：
  - 异常管理器功能测试：
    - 异常报告的添加和查询
    - 按ID和客户端ID查询异常报告
    - 空结果处理
  - 客户端管理器异常功能测试：
    - 异常上报和客户端状态更新
    - 异常查询和筛选
    - 错误处理（如客户端不存在）
  - API端点测试：
    - 获取所有异常报告
    - 按客户端ID筛选异常报告
    - 创建新的异常报告
    - 获取特定异常详情
    - 错误处理（如异常不存在）
- **实现特点**：
  - 异常严重性分级：info、warning、error、critical
  - 高严重性异常（error、critical）自动更新客户端状态
  - 支持详细的异常信息记录，包括模块、堆栈跟踪和附加信息
  - 提供完整的控制台命令和API接口用于异常管理
  - 使用互斥锁（sync.RWMutex）确保并发安全
  - 支持按ID和客户端ID高效查询异常报告
- **调用示例**：
  ```
  # 控制台命令
  > exception list
  All exceptions:
  ID                                   Client ID                 Severity   Timestamp           Message
  --------------------------------------------------------------------------------
  test-client-id-1709654695000000000   test-client-id           error      2025-03-06 05:44:55 Test exception message
  
  Total: 1 exceptions
  
  > exception report test-client-id critical "Critical system error" system-module "Stack trace..."
  Exception reported with ID: test-client-id-1709654700000000000
  
  # API调用
  GET /api/exceptions
  Response:
  [
    {
      "id": "test-client-id-1709654695000000000",
      "client_id": "test-client-id",
      "timestamp": "2025-03-06T05:44:55Z",
      "message": "Test exception message",
      "severity": "error",
      "module": "test-module",
      "stack_trace": "...",
      "additional_info": {"key": "value"}
    }
  ]
  
  POST /api/exceptions
  Request:
  {
    "clientId": "test-client-id",
    "message": "Test exception from API",
    "severity": "error",
    "module": "test-module-api"
  }
  Response:
  {
    "id": "test-client-id-1709654710000000000",
    "client_id": "test-client-id",
    "timestamp": "2025-03-06T05:45:10Z",
    "message": "Test exception from API",
    "severity": "error",
    "module": "test-module-api"
  }
  ```
- **遇到的问题与解决方案**：
  - 问题：API测试中的ExceptionReport类型引用错误
  - 解决方案：使用map[string]interface{}替代直接引用ExceptionReport结构体
  - 问题：API处理器直接访问exceptionManager导致编译错误
  - 解决方案：通过ClientManager提供的方法间接访问异常报告
  - 问题：控制台输入流EOF错误处理不当
  - 解决方案：添加对io.EOF的专门处理，优雅退出控制台
- **下一步计划**：
  - 实现客户端模块管理API
  - 增强API安全性和JWT认证实现
  - 开发客户端模板和Builder工具

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

#### 2.1.4 日志与监控模块实现
- [x] 实现分级日志系统（调试、信息、告警、错误）
- [x] 实现异常检测与自动重连机制

### 任务2.1.4：日志与监控模块实现
- **完成时间**：2025-03-06
- **主要内容**：
  1. 实现分级日志系统（调试、信息、告警、错误、致命）
  2. 实现异常检测与自动重连机制
  3. 编写单元测试和编译测试，确保代码质量
- **代码实现**：
  - 日志系统（Logger）：
    - 基于logrus实现的日志接口
    - 支持不同级别的日志记录（调试、信息、告警、错误、致命）
    - 提供全局日志实例和线程安全的日志记录方法
    - 支持字段注入和格式化输出
    - 支持动态调整日志级别和输出目标
  - 监控管理器（MonitorManager）：
    - 定期检查客户端异常状态
    - 实现自动重连机制，支持配置重连间隔和最大尝试次数
    - 提供详细的异常日志记录
    - 使用context控制监控器生命周期
  - 服务器集成：
    - 在Server结构中添加日志和监控组件
    - 在启动和停止过程中记录详细日志
    - 为所有关键操作添加日志记录
    - 提供GetLogger和GetMonitorManager方法
- **测试内容**：
  - 日志系统功能验证：
    - 不同日志级别的记录和过滤
    - 字段注入和格式化输出
    - 日志级别动态调整
    - 线程安全性测试
  - 监控管理器测试：
    - 异常检测和客户端状态监控
    - 自动重连尝试和成功恢复
    - 配置参数验证
    - 生命周期管理（启动和停止）
- **实现特点**：
  - 基于logrus实现的高性能日志系统
  - 支持字段注入和结构化日志
  - 线程安全的日志记录和监控操作
  - 可配置的监控参数（检查间隔、重连间隔、最大尝试次数）
  - 与现有客户端管理和异常报告系统无缝集成
  - 使用context控制组件生命周期，确保优雅关闭
  - 提供详细的操作日志，便于问题诊断和系统监控
- **调用示例**：
  ```go
  // 初始化日志
  logger := logging.GetLogger()
  logger.SetLevel(logging.InfoLevel)
  
  // 记录不同级别的日志
  logger.Debug("调试信息", nil)
  logger.Info("操作信息", map[string]interface{}{
    "operation": "start_server",
    "time": time.Now().Format(time.RFC3339),
  })
  logger.Warn("警告信息", map[string]interface{}{
    "warning_type": "connection_delay",
    "delay_ms": 500,
  })
  logger.Error("错误信息", map[string]interface{}{
    "error_code": "ERR001",
    "client_id": "test-client-id",
  })
  
  // 创建监控管理器
  monitorConfig := logging.MonitorConfig{
    CheckInterval:        1 * time.Minute,
    ReconnectInterval:    10 * time.Second,
    MaxReconnectAttempts: 5,
  }
  monitor := logging.NewMonitorManager(logger, clientManager, monitorConfig)
  
  // 启动和停止监控
  monitor.Start()
  // ... 系统运行 ...
  monitor.Stop()
  ```
- **遇到的问题与解决方案**：
  - 问题：WithField和WithFields方法实现中的并发安全问题
  - 解决方案：重构实现，使用读写锁保护日志操作，并确保正确复制logger实例
  - 问题：监控器中的重连逻辑可能导致goroutine泄漏
  - 解决方案：使用context控制重连尝试的生命周期，确保在监控器停止时取消所有重连尝试
  - 问题：日志输出格式不一致
  - 解决方案：统一使用TextFormatter并配置一致的时间戳格式
- **下一步计划**：
  - 增强日志系统，支持文件输出和日志轮转
  - 实现更复杂的异常检测算法，支持模式识别
  - 添加系统资源监控功能（CPU、内存、网络）
  - 实现日志聚合和分析功能

### 任务2.1.4b：日志与监控模块增强
- **完成时间**：2025-03-06
- **主要内容**：
  1. 实现文件输出和日志轮转功能
  2. 修复WithField和WithFields方法的并发安全问题
  3. 实现系统资源监控（CPU、内存、网络）
  4. 实现异常模式检测和分析
  5. 实现日志聚合和分析功能
- **代码实现**：
  - 文件日志系统：
    - 支持配置日志文件目录、大小限制和轮转策略
    - 使用lumberjack实现日志轮转
    - 支持同时输出到控制台和文件
    - 提供EnableFileLogging和DisableFileLogging方法
  - 资源监控系统（ResourceMonitor）：
    - 监控CPU使用率、内存使用和网络流量
    - 提供历史数据存储和统计分析
    - 支持可配置的监控间隔和启用选项
    - 使用gopsutil库获取系统资源信息
  - 异常模式检测（PatternDetector）：
    - 分析异常报告中的模式和频率
    - 识别重复出现的异常信息
    - 提供按客户端和模块的异常模式分组
    - 支持配置时间窗口和最小频率阈值
  - 日志聚合与分析（LogAnalyzer）：
    - 收集和存储日志条目
    - 提供按级别和消息的统计分析
    - 识别最常见的日志消息和模式
    - 支持配置最大条目数和分析参数
  - 服务器集成：
    - 在Server结构中添加资源监控、模式检测和日志分析组件
    - 在启动时初始化和配置所有组件
    - 在停止时优雅关闭所有组件
    - 提供GetResourceMonitor、GetPatternDetector和GetLogAnalyzer方法
- **测试内容**：
  - 文件日志功能测试：
    - 验证日志文件创建和写入
    - 测试WithField方法的文件输出
    - 测试DisableFileLogging功能
  - 资源监控测试：
    - 验证CPU、内存和网络数据收集
    - 测试不同配置选项（启用/禁用特定监控）
    - 验证历史数据存储和检索
  - 异常模式检测测试：
    - 测试模式识别算法
    - 验证按客户端和模块的模式分组
    - 测试时间窗口和频率阈值配置
  - 日志分析测试：
    - 验证日志条目收集和统计
    - 测试最大条目数限制
    - 验证按级别筛选和统计功能
- **实现特点**：
  - 使用lumberjack实现高效的日志轮转
  - 基于gopsutil实现跨平台的资源监控
  - 线程安全的模式检测和日志分析
  - 与现有系统的无缝集成
  - 模块化设计，支持灵活配置
  - 完整的测试覆盖，确保代码质量
- **调用示例**：
  ```go
  // 启用文件日志
  fileConfig := logging.FileLogConfig{
    Directory:  "logs",
    MaxSize:    10, // 10 MB
    MaxAge:     7,  // 7 days
    MaxBackups: 5,
    Compress:   true,
  }
  logger.EnableFileLogging(fileConfig)
  
  // 创建资源监控器
  resourceConfig := logging.ResourceMonitorConfig{
    Interval:      30 * time.Second,
    EnableCPU:     true,
    EnableMemory:  true,
    EnableNetwork: true,
  }
  resourceMonitor := logging.NewResourceMonitor(logger, resourceConfig)
  resourceMonitor.Start()
  
  // 获取资源统计
  stats := resourceMonitor.GetLatestStats()
  fmt.Printf("CPU: %.2f%%, Memory: %.2f%%\n", stats.CPUUsage, stats.MemoryUsage)
  
  // 创建模式检测器
  patternConfig := logging.PatternDetectorConfig{
    TimeWindow:         60 * 60, // 1 hour
    MinFrequency:       3,
    SimilarityThreshold: 0.8,
  }
  detector := logging.NewPatternDetector(logger, clientManager, patternConfig)
  
  // 检测异常模式
  patterns := detector.DetectPatterns()
  for _, pattern := range patterns {
    fmt.Printf("Pattern: %s, Frequency: %d\n", pattern.MessagePattern, pattern.Frequency)
  }
  
  // 创建日志分析器
  analyzerConfig := logging.LogAnalyzerConfig{
    MaxEntries:      1000,
    TopMessageCount: 10,
  }
  analyzer := logging.NewLogAnalyzer(analyzerConfig)
  
  // 添加日志条目并获取统计
  analyzer.AddEntry(logging.LogEntry{
    Timestamp: time.Now(),
    Level:     logging.InfoLevel,
    Message:   "Test message",
    Fields:    map[string]interface{}{"key": "value"},
  })
  
  stats := analyzer.GetStats()
  fmt.Printf("Total entries: %d\n", stats.TotalEntries)
  ```
- **遇到的问题与解决方案**：
  - 问题：文件日志路径创建失败
  - 解决方案：添加目录自动创建功能，确保日志目录存在
  - 问题：资源监控数据收集的性能开销
  - 解决方案：使用可配置的监控间隔和选择性启用监控项，减少资源消耗
  - 问题：模式检测算法的准确性
  - 解决方案：实现基于消息相似度的模式检测，支持配置相似度阈值
  - 问题：日志分析器内存占用过高
  - 解决方案：实现最大条目数限制，自动清理最旧的日志条目
- **下一步计划**：
  - 实现更高级的日志分析算法，支持异常检测
  - 添加日志可视化功能，提供图表和仪表板
  - 实现分布式日志收集和聚合
  - 增强资源监控，支持自定义告警阈值

### 任务2.2.2：加密通信模块实现
- **完成时间**：2025-03-05
- **主要内容**：
  1. 实现AES与ChaCha20加密/解密函数
  2. 设计密钥协商与定期更新机制
  3. 编写单元测试验证不同密钥及加密方式的正确性
- **代码实现**：
  - 加密接口（Encrypter）：
    - 定义通用加密接口，支持加密、解密、密钥轮换等操作
    - 提供GetType和GetKeyID方法用于识别加密类型和密钥版本
    - 支持多种加密类型（AES、ChaCha20）和无加密模式
  - AES加密实现：
    - 基于AES-GCM实现的认证加密
    - 支持128位、192位和256位密钥长度
    - 实现密钥轮换和版本管理
    - 使用随机生成的nonce确保加密安全性
  - ChaCha20加密实现：
    - 基于ChaCha20-Poly1305实现的认证加密
    - 使用32字节密钥和12字节nonce
    - 实现密钥轮换和版本管理
    - 提供高性能的加密和解密操作
  - 密钥交换机制：
    - 基于ECDH（椭圆曲线Diffie-Hellman）实现安全的密钥交换
    - 使用P-256曲线生成密钥对
    - 支持计算共享密钥和密钥协商
    - 提供密钥轮换时间配置
  - 消息格式：
    - 设计包含版本、加密类型、密钥ID和时间戳的消息头
    - 支持JSON序列化和反序列化
    - 提供加密元数据，便于接收方正确解密
  - 协议集成：
    - 更新Protocol接口，添加加密相关方法
    - 实现BaseProtocol中的加密支持
    - 在ProtocolManager中添加加密类型管理
  - 服务器端加密处理：
    - 实现服务器端加密管理器
    - 支持客户端特定的加密状态
    - 实现加密监听器包装器
    - 提供密钥协商和更新机制
- **测试内容**：
  - 随机字节生成测试：
    - 验证不同长度的随机字节生成
    - 确保生成的字节具有随机性
  - 加密类型测试：
    - 验证EncryptionType常量定义
    - 测试字符串表示和比较
  - AES加密测试：
    - 测试不同密钥长度（128位、192位、256位）
    - 验证加密和解密操作
    - 测试密钥轮换功能
    - 验证无效输入处理
  - ChaCha20加密测试：
    - 测试加密器创建和密钥验证
    - 验证加密和解密操作
    - 测试密钥轮换功能
    - 验证无效输入处理
  - 密钥交换测试：
    - 测试ECDH密钥对生成
    - 验证共享密钥计算
    - 测试密钥交换消息格式
    - 验证无效输入处理
  - 消息格式测试：
    - 测试消息创建和属性验证
    - 验证JSON序列化和反序列化
    - 测试与加密集成
    - 验证消息头解析
- **实现特点**：
  - 模块化设计，支持多种加密算法
  - 基于接口的实现，便于扩展和替换
  - 完整的密钥生命周期管理
  - 安全的密钥交换机制
  - 高性能的加密和解密操作
  - 与协议层无缝集成
  - 全面的测试覆盖
- **调用示例**：
  ```go
  // 创建AES加密器
  aesEncrypter, err := encryption.NewAESEncrypter(32) // 256位密钥
  if err != nil {
      log.Fatalf("Failed to create AES encrypter: %v", err)
  }
  
  // 加密数据
  plaintext := []byte("Hello, world!")
  ciphertext, err := aesEncrypter.Encrypt(plaintext)
  if err != nil {
      log.Fatalf("Encryption failed: %v", err)
  }
  
  // 解密数据
  decrypted, err := aesEncrypter.Decrypt(ciphertext)
  if err != nil {
      log.Fatalf("Decryption failed: %v", err)
  }
  
  fmt.Printf("Decrypted: %s\n", string(decrypted))
  
  // 创建ChaCha20加密器
  chaCha20Encrypter, err := encryption.NewChaCha20Encrypter()
  if err != nil {
      log.Fatalf("Failed to create ChaCha20 encrypter: %v", err)
  }
  
  // 密钥交换
  keyExchanger, err := encryption.NewECDHKeyExchanger()
  if err != nil {
      log.Fatalf("Failed to create key exchanger: %v", err)
  }
  
  publicKey := keyExchanger.GetPublicKey()
  
  // 创建密钥交换消息
  keyExchangeMsg := encryption.NewKeyExchangeMessage(
      encryption.EncryptionAES,
      publicKey,
      time.Now().Add(24 * time.Hour).Unix(),
  )
  
  // 序列化为JSON
  jsonData, err := keyExchangeMsg.ToJSON()
  if err != nil {
      log.Fatalf("Failed to serialize message: %v", err)
  }
  
  fmt.Printf("Key exchange message: %s\n", string(jsonData))
  ```
- **遇到的问题与解决方案**：
  - 问题：ChaCha20-Poly1305实现依赖缺失
  - 解决方案：添加golang.org/x/crypto/chacha20poly1305依赖
  - 问题：密钥轮换后旧密钥解密失败
  - 解决方案：实现密钥ID管理，确保使用正确的密钥进行解密
  - 问题：ECDH密钥交换中的公钥格式不一致
  - 解决方案：统一使用X509编码格式的公钥
  - 问题：消息格式中的时间戳处理不当
  - 解决方案：使用int64类型存储Unix时间戳，确保跨平台兼容性
- **下一步计划**：
  - 实现客户端心跳模块（含随机延时、超时切换）
  - 确保特殊协议（如DNS）支持心跳
  - 实现模块化设计，支持动态加载、调用和卸载
  - 实现通信反馈机制

### 任务2.2.3：心跳与模块化设计
- **完成时间**：2025-03-05
- **主要内容**：
  1. 实现客户端心跳模块（含随机延时、超时切换）
  2. 实现模块化设计，支持动态加载、调用和卸载模块
  3. 编写单元测试验证心跳机制和模块系统功能
- **代码实现**：
  - 客户端心跳模块：
    - 扩展Client结构体，添加心跳相关字段：
      - randomHeartbeatEnabled：是否启用随机心跳间隔
      - minHeartbeatInterval：最小心跳间隔（默认1秒）
      - maxHeartbeatInterval：最大心跳间隔（默认24小时）
      - heartbeatFailCount：心跳失败计数
      - heartbeatTimeout：心跳超时时间
      - lastHeartbeatTime：上次心跳时间
    - 实现随机心跳功能：
      - EnableRandomHeartbeat：启用随机心跳间隔
      - DisableRandomHeartbeat：禁用随机心跳间隔
      - updateHeartbeatInterval：更新随机心跳间隔
    - 增强心跳循环和发送功能：
      - 更新heartbeatLoop方法，支持心跳失败检测和协议切换
      - 更新sendHeartbeat方法，添加心跳状态信息
      - 实现心跳失败计数和协议切换逻辑
  - 模块化设计：
    - 增强ModuleManager功能：
      - 添加LoadModuleFromBytes方法，支持从字节数组动态加载模块
      - 添加IsModuleLoaded方法，检查模块是否已加载
      - 使用Go的plugin包实现动态模块加载
    - 实现Shell模块作为示例：
      - 创建shell模块包和结构体
      - 实现Module接口的所有方法（Init、Execute、Cleanup）
      - 支持跨平台的shell命令执行（Windows和Unix-like系统）
      - 提供结构化的参数和结果处理
    - 更新客户端模块处理：
      - 增强handleLoadModule方法，支持从字节数组加载模块
      - 添加模块加载结果的详细反馈
      - 实现模块加载错误处理
- **测试内容**：
  - 客户端心跳测试：
    - 测试基本心跳功能和间隔设置
    - 验证随机心跳启用和禁用功能
    - 测试心跳失败处理和协议切换
    - 验证心跳状态信息的正确性
  - 模块管理器测试：
    - 测试模块加载和卸载功能
    - 验证模块执行和参数传递
    - 测试模块查询和列表功能
    - 验证错误处理（如重复加载、模块不存在）
  - Shell模块测试：
    - 测试命令执行功能
    - 验证跨平台支持
    - 测试错误处理和结果格式
    - 验证初始化和清理功能
- **实现特点**：
  - 灵活的心跳配置，支持1秒到24小时的随机间隔
  - 智能的协议切换机制，基于心跳失败计数
  - 安全的动态模块加载，使用Go的plugin系统
  - 统一的模块接口，确保一致的行为
  - 跨平台的Shell模块实现
  - 详细的错误处理和状态反馈
  - 全面的测试覆盖，验证所有核心功能
- **调用示例**：
  ```go
  // 创建客户端
  config := client.Config{
      ID:                     "test-client",
      Name:                   "Test Client",
      ServerAddresses:        map[string]string{"tcp": "tcp://localhost:8080"},
      HeartbeatInterval:      60 * time.Second,
      ProtocolSwitchThreshold: 3,
  }
  
  c, err := client.NewClient(config)
  if err != nil {
      log.Fatalf("Failed to create client: %v", err)
  }
  
  // 启动客户端
  err = c.Start()
  if err != nil {
      log.Fatalf("Failed to start client: %v", err)
  }
  
  // 启用随机心跳
  c.EnableRandomHeartbeat(5*time.Second, 5*time.Minute)
  
  // 加载Shell模块
  shellModuleBytes := []byte{...} // 模块字节码
  err = c.moduleMgr.LoadModuleFromBytes("shell", shellModuleBytes)
  if err != nil {
      log.Fatalf("Failed to load shell module: %v", err)
  }
  
  // 执行Shell模块
  params := json.RawMessage(`{"command": "echo hello"}`)
  result, err := c.moduleMgr.ExecuteModule(context.Background(), "shell", params)
  if err != nil {
      log.Fatalf("Failed to execute shell module: %v", err)
  }
  
  var shellResult struct {
      Success bool   `json:"success"`
      Output  string `json:"output"`
      Error   string `json:"error,omitempty"`
  }
  
  err = json.Unmarshal(result, &shellResult)
  if err != nil {
      log.Fatalf("Failed to parse shell result: %v", err)
  }
  
  fmt.Printf("Shell output: %s\n", shellResult.Output)
  
  // 禁用随机心跳
  c.DisableRandomHeartbeat()
  
  // 停止客户端
  err = c.Stop()
  if err != nil {
      log.Fatalf("Failed to stop client: %v", err)
  }
  ```
- **遇到的问题与解决方案**：
  - 问题：Go plugin系统的平台限制
  - 解决方案：添加平台检测，在不支持的平台上提供备选实现
  - 问题：心跳随机间隔可能导致过于频繁的心跳
  - 解决方案：实现最小间隔限制，确保心跳不会过于频繁
  - 问题：模块加载时的临时文件管理
  - 解决方案：使用ioutil.TempFile创建临时文件，并在加载后自动删除
  - 问题：Shell模块在不同操作系统上的命令差异
  - 解决方案：使用runtime.GOOS检测操作系统，并使用相应的shell命令（cmd或sh）
- **下一步计划**：
  - 实现更多内置模块（文件操作、进程管理、网络工具）
  - 增强模块安全性，添加签名验证
  - 实现模块版本管理和兼容性检查
  - 开发Builder工具，支持客户端定制化构建

### 任务2.2.4：通信反馈机制实现
- **完成时间**：2025-03-07
- **主要内容**：
  1. 实现客户端在接收 Server 下发指令后，返回执行结果与状态反馈
  2. 实现错误处理及自动重试逻辑
  3. 编写单元测试验证反馈流程，记录测试结果
- **代码实现**：
  - 反馈配置（FeedbackConfig）：
    - 最大重试次数（MaxRetries）：配置最大允许的重试次数
    - 重试间隔（RetryInterval）：初始重试间隔时间
    - 最大重试间隔（MaxRetryInterval）：重试间隔的上限
    - 重试退避因子（RetryBackoffFactor）：每次重试后间隔增加的倍数
  - 反馈响应（FeedbackResponse）：
    - 类型（Type）：模块执行结果、模块加载结果、模块卸载结果
    - 客户端ID（ClientID）：当前客户端的唯一标识
    - 命令ID（CommandID）：服务器下发命令的唯一标识
    - 模块名称（Module）：相关模块的名称
    - 成功标志（Success）：操作是否成功
    - 结果数据（Result）：操作的结果数据（JSON格式）
    - 错误信息（Error）：失败时的错误信息
    - 重试计数（RetryCount）：已尝试的重试次数
    - 状态信息（Status）：处理中、重试中、已完成、失败
    - 时间戳（Timestamp）：反馈生成的时间戳
  - 反馈发送（sendFeedback）：
    - 实现指数退避重试逻辑，初始间隔为RetryInterval
    - 每次重试后，间隔时间乘以RetryBackoffFactor
    - 间隔时间上限为MaxRetryInterval
    - 支持最大重试次数限制（MaxRetries）
    - 支持上下文取消，确保资源正确释放
  - 命令处理增强：
    - 模块执行处理（handleExecuteModule）：
      - 发送初始"processing"状态反馈
      - 执行模块并处理结果
      - 对可重试错误实现自动重试
      - 发送最终状态反馈（completed或failed）
    - 模块加载处理（handleLoadModule）：
      - 发送初始"processing"状态反馈
      - 加载模块并处理结果
      - 对可重试错误实现自动重试
      - 发送最终状态反馈
    - 模块卸载处理（handleUnloadModule）：
      - 发送初始"processing"状态反馈
      - 卸载模块并处理结果
      - 对可重试错误实现自动重试
      - 发送最终状态反馈
  - 错误处理：
    - 实现isRetryableError函数，区分可重试和不可重试错误
    - 可重试错误包括网络相关错误（如连接重置、超时）
    - 不可重试错误包括上下文取消、参数错误等
    - 记录重试次数和详细错误信息
    - 提供完整的状态反馈，便于服务器端诊断
- **测试内容**：
  - 反馈机制基本功能测试：
    - 验证初始"processing"状态反馈正确发送
    - 验证最终状态反馈（completed或failed）正确发送
    - 验证命令ID和模块名称正确传递
    - 验证结果数据和错误信息正确包含
  - 重试逻辑测试：
    - 验证可重试错误的自动重试功能
    - 验证重试计数正确记录和传递
    - 验证指数退避算法正确实现
    - 验证最大重试次数限制有效
    - 验证重试间隔上限正确应用
  - 错误处理测试：
    - 验证不同类型错误的正确分类（可重试vs不可重试）
    - 验证错误信息正确传递到反馈响应
    - 验证状态信息正确更新（processing→retrying→completed/failed）
    - 验证上下文取消时的正确处理
- **实现特点**：
  - 标准化的反馈格式，便于服务器端处理和分析
  - 完整的状态跟踪，提供操作全生命周期的可见性
  - 智能的重试机制，提高操作成功率
  - 指数退避算法，避免频繁重试导致的资源浪费
  - 详细的错误信息，便于问题诊断
  - 与现有模块系统无缝集成
  - 线程安全的实现，支持并发操作
- **调用示例**：
  ```json
  // 初始反馈（处理中）
  {
    "type": "module_result",
    "client_id": "test-client",
    "command_id": "test-command-1",
    "module": "shell",
    "success": false,
    "status": "processing",
    "timestamp": 1709654800
  }
  
  // 重试反馈
  {
    "type": "module_result",
    "client_id": "test-client",
    "command_id": "test-command-1",
    "module": "shell",
    "success": false,
    "error": "connection reset by peer",
    "retry_count": 1,
    "status": "retrying",
    "timestamp": 1709654801
  }
  
  // 最终反馈（成功）
  {
    "type": "module_result",
    "client_id": "test-client",
    "command_id": "test-command-1",
    "module": "shell",
    "success": true,
    "result": {"output": "command output"},
    "retry_count": 1,
    "status": "completed",
    "timestamp": 1709654802
  }
  
  // 最终反馈（失败）
  {
    "type": "module_result",
    "client_id": "test-client",
    "command_id": "test-command-1",
    "module": "shell",
    "success": false,
    "error": "maximum retry attempts exceeded: connection reset by peer",
    "retry_count": 3,
    "status": "failed",
    "timestamp": 1709654805
  }
  ```
- **测试结果**：
  - 所有单元测试通过
  - 验证了反馈机制的正确性和可靠性
  - 验证了重试逻辑的有效性，包括指数退避算法
  - 验证了错误处理的完整性，包括可重试和不可重试错误的区分
  - 验证了与现有模块系统的兼容性
  - 验证了并发安全性，确保多个命令同时处理时的正确行为
- **遇到的问题与解决方案**：
  - 问题：重试逻辑中的goroutine泄漏
  - 解决方案：使用context.Done()确保在客户端停止时取消所有重试
  - 问题：JSON序列化时的循环引用
  - 解决方案：使用json.RawMessage类型存储结果数据，避免循环引用
  - 问题：重试间隔计算中的溢出风险
  - 解决方案：添加最大重试间隔限制，防止间隔时间过长
  - 问题：并发发送反馈时的竞态条件
  - 解决方案：使用互斥锁保护关键操作，确保线程安全
- **下一步计划**：
  - 增强反馈机制，支持批量命令处理
  - 实现反馈压缩，减少网络流量
  - 添加反馈优先级，确保关键反馈优先发送
  - 实现反馈缓存，处理网络暂时不可用的情况
  - 开发Builder工具，支持客户端定制化构建
