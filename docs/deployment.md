# C2系统部署文档

## 目录
1. [系统要求](#系统要求)
2. [安装步骤](#安装步骤)
3. [配置选项](#配置选项)
4. [安全建议](#安全建议)
5. [监控与维护](#监控与维护)
6. [故障排除](#故障排除)
7. [常见问题](#常见问题)

## 系统要求

### 服务器要求
- **操作系统**: Linux (推荐 Ubuntu 20.04 LTS 或更高版本)
- **CPU**: 至少4核心 (推荐8核心或更多用于高负载环境)
- **内存**: 至少4GB RAM (推荐8GB或更多用于高负载环境)
- **存储**: 至少20GB可用空间
- **网络**: 稳定的网络连接，支持所有需要的协议端口开放

### 客户端要求
- **操作系统**: 支持Windows、Linux、macOS
- **内存**: 至少256MB RAM
- **存储**: 至少50MB可用空间
- **网络**: 能够连接到服务器的网络环境

### 软件依赖
- Go 1.23.5 或更高版本
- Git
- 防火墙配置允许所需的端口通信

## 安装步骤

### 1. 准备环境

```bash
# 安装Go
wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile

# 安装Git
sudo apt update
sudo apt install -y git

# 克隆仓库
git clone https://github.com/Cl0udRs4/dinot.git
cd dinot
```

### 2. 编译服务器和构建工具

```bash
# 创建bin目录
mkdir -p bin

# 编译服务器
go build -o bin/server cmd/server/main.go

# 编译构建工具
go build -o bin/builder cmd/builder/main.go
```

### 3. 生成客户端

```bash
# 使用脚本生成各种协议的客户端
./scripts/generate_clients.sh ./bin/clients
```

### 4. 配置服务器

创建配置文件目录并设置基本配置：

```bash
mkdir -p configs
cat > configs/server.json << EOF
{
  "log_level": "info",
  "enable_console": true,
  "enable_api": true,
  "api_port": 8090,
  "tcp_port": 8080,
  "udp_port": 8081,
  "ws_port": 8082,
  "dns_port": 8053,
  "dns_domain": "example.com",
  "heartbeat_check_interval": 10,
  "heartbeat_timeout": 60,
  "max_connections": 1000,
  "buffer_size": 4096,
  "auth_enabled": true,
  "auth_user": "admin",
  "auth_password": "CHANGE_THIS_PASSWORD",
  "jwt_enabled": true,
  "jwt_secret": "CHANGE_THIS_SECRET_KEY"
}
EOF
```

**重要**: 请务必修改默认的认证凭据和JWT密钥！

## 配置选项

### 服务器配置参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| log_level | 日志级别 (debug, info, warn, error) | info |
| enable_console | 启用控制台模式 | true |
| enable_api | 启用API模式 | true |
| api_port | API服务器端口 | 8090 |
| tcp_port | TCP监听端口 | 8080 |
| udp_port | UDP监听端口 | 8081 |
| ws_port | WebSocket监听端口 | 8082 |
| dns_port | DNS监听端口 | 8053 |
| dns_domain | DNS域名 | example.com |
| heartbeat_check_interval | 心跳检查间隔(秒) | 10 |
| heartbeat_timeout | 心跳超时(秒) | 60 |
| max_connections | 每个协议的最大连接数 | 1000 |
| buffer_size | 缓冲区大小(字节) | 4096 |
| auth_enabled | 启用HTTP基本认证 | true |
| auth_user | HTTP基本认证用户名 | admin |
| auth_password | HTTP基本认证密码 | - |
| jwt_enabled | 启用JWT认证 | true |
| jwt_secret | JWT密钥 | - |

### 客户端构建参数

使用Builder工具构建客户端时可用的参数：

```bash
./bin/builder build -h
```

| 参数 | 描述 | 示例 |
|------|------|------|
| -p, --protocol | 通信协议 | tcp,udp,ws,dns,icmp |
| -d, --domain | 服务器域名 | example.com |
| -s, --servers | 服务器地址列表 | tcp:10.0.0.1:8080,udp:10.0.0.1:8081 |
| -m, --modules | 要包含的模块 | shell,system,file |
| -e, --encryption | 加密类型 | aes,chacha20 |
| -o, --output | 输出文件路径 | ./clients/client_tcp |
| --debug | 启用调试模式 | - |

## 安全建议

### 网络安全

1. **防火墙配置**
   - 仅开放必要的端口
   - 使用iptables或ufw限制访问来源
   - 示例配置:
     ```bash
     # 允许API端口仅从特定IP访问
     sudo ufw allow from 10.0.0.0/24 to any port 8090
     
     # 允许C2通信端口
     sudo ufw allow 8080/tcp
     sudo ufw allow 8081/udp
     sudo ufw allow 8082/tcp
     sudo ufw allow 8053/udp
     ```

2. **使用TLS/SSL**
   - 为API和WebSocket连接配置TLS
   - 使用Let's Encrypt获取免费证书
   - 示例配置:
     ```bash
     # 安装certbot
     sudo apt install -y certbot
     
     # 获取证书
     sudo certbot certonly --standalone -d c2.example.com
     
     # 配置服务器使用证书
     # 在configs/server.json中添加:
     # "tls_enabled": true,
     # "tls_cert_file": "/etc/letsencrypt/live/c2.example.com/fullchain.pem",
     # "tls_key_file": "/etc/letsencrypt/live/c2.example.com/privkey.pem"
     ```

### 认证与授权

1. **API认证**
   - 启用JWT认证
   - 定期轮换JWT密钥
   - 使用强密码和长令牌
   - 限制令牌有效期

2. **客户端认证**
   - 使用双向认证
   - 实施客户端签名验证
   - 启用前向保密

### 数据安全

1. **加密通信**
   - 使用AES-256-GCM或ChaCha20-Poly1305加密
   - 定期轮换加密密钥
   - 实施消息完整性验证

2. **敏感数据处理**
   - 不在日志中记录敏感信息
   - 使用内存安全的方式处理密钥
   - 实施安全的密钥存储

## 监控与维护

### 日志管理

1. **日志轮换**
   ```bash
   # 创建日志目录
   mkdir -p /var/log/c2server
   
   # 配置logrotate
   cat > /etc/logrotate.d/c2server << EOF
   /var/log/c2server/*.log {
       daily
       missingok
       rotate 7
       compress
       delaycompress
       notifempty
       create 640 root root
   }
   EOF
   ```

2. **日志监控**
   - 使用ELK Stack或Graylog进行集中式日志管理
   - 设置关键事件的告警

### 性能监控

1. **系统监控**
   - 使用Prometheus + Grafana监控系统资源
   - 监控CPU、内存、磁盘和网络使用情况

2. **应用监控**
   - 监控客户端连接数
   - 监控命令执行成功率
   - 监控心跳状态

### 备份策略

1. **配置备份**
   ```bash
   # 创建备份脚本
   cat > backup.sh << EOF
   #!/bin/bash
   BACKUP_DIR="/backup/c2server/\$(date +%Y%m%d)"
   mkdir -p \$BACKUP_DIR
   cp -r configs/* \$BACKUP_DIR/
   cp -r bin/clients/* \$BACKUP_DIR/clients/
   EOF
   
   chmod +x backup.sh
   
   # 添加到crontab
   echo "0 2 * * * /path/to/backup.sh" | crontab -
   ```

## 故障排除

### 常见问题与解决方案

1. **服务器无法启动**
   - 检查端口是否被占用: `netstat -tulpn | grep <port>`
   - 检查日志文件中的错误信息
   - 验证配置文件格式是否正确

2. **客户端无法连接**
   - 检查网络连接和防火墙设置
   - 验证客户端配置中的服务器地址是否正确
   - 检查服务器日志中的连接尝试

3. **API请求失败**
   - 验证认证凭据是否正确
   - 检查API端口是否开放
   - 查看服务器日志中的API请求记录

### 诊断命令

```bash
# 检查服务器状态
ps aux | grep server

# 检查端口监听状态
netstat -tulpn | grep server

# 检查日志文件
tail -f /var/log/c2server/server.log

# 测试API连接
curl -u admin:password http://localhost:8090/api/status

# 检查DNS解析
dig @localhost -p 8053 example.com
```

## 常见问题

### Q: 如何增加新的模块?
A: 在`internal/modules`目录下创建新模块，实现`Module`接口，然后在Builder配置中包含该模块。

### Q: 如何扩展系统支持更多客户端?
A: 考虑以下扩展策略:
1. 增加`max_connections`参数值
2. 部署多个服务器实例并使用负载均衡
3. 优化内存使用和连接处理逻辑

### Q: 如何实现高可用部署?
A: 使用以下方法实现高可用:
1. 部署多个服务器实例
2. 使用负载均衡器分发流量
3. 实现服务器状态同步
4. 使用数据库存储客户端状态信息

### Q: 如何安全地部署在公网环境?
A: 公网部署安全建议:
1. 使用反向代理(如Nginx)保护API端口
2. 实施IP白名单限制
3. 启用所有安全功能(TLS, JWT, 加密通信)
4. 定期更新系统和依赖
5. 实施入侵检测系统监控

---

## 生产环境部署示例

以下是一个完整的生产环境部署示例:

```bash
#!/bin/bash
# 生产环境部署脚本

# 1. 创建部署目录
mkdir -p /opt/c2server
mkdir -p /opt/c2server/bin
mkdir -p /opt/c2server/configs
mkdir -p /opt/c2server/clients
mkdir -p /var/log/c2server

# 2. 编译服务器和构建工具
cd /path/to/dinot
go build -o /opt/c2server/bin/server cmd/server/main.go
go build -o /opt/c2server/bin/builder cmd/builder/main.go

# 3. 创建配置文件
cat > /opt/c2server/configs/server.json << EOF
{
  "log_level": "info",
  "enable_console": false,
  "enable_api": true,
  "api_port": 8090,
  "tcp_port": 8080,
  "udp_port": 8081,
  "ws_port": 8082,
  "dns_port": 8053,
  "dns_domain": "example.com",
  "heartbeat_check_interval": 10,
  "heartbeat_timeout": 60,
  "max_connections": 1000,
  "buffer_size": 4096,
  "auth_enabled": true,
  "auth_user": "admin",
  "auth_password": "$(openssl rand -base64 16)",
  "jwt_enabled": true,
  "jwt_secret": "$(openssl rand -base64 32)",
  "log_file": "/var/log/c2server/server.log",
  "tls_enabled": true,
  "tls_cert_file": "/etc/letsencrypt/live/c2.example.com/fullchain.pem",
  "tls_key_file": "/etc/letsencrypt/live/c2.example.com/privkey.pem"
}
EOF

# 4. 生成客户端
/opt/c2server/bin/builder build -p tcp,udp,ws,dns,icmp -d c2.example.com -s tcp:c2.example.com:8080,udp:c2.example.com:8081,ws:c2.example.com:8082,dns:c2.example.com:8053,icmp:c2.example.com -m shell,system,file -e aes -o /opt/c2server/clients/client

# 5. 创建systemd服务
cat > /etc/systemd/system/c2server.service << EOF
[Unit]
Description=C2 Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/c2server
ExecStart=/opt/c2server/bin/server --config /opt/c2server/configs/server.json
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# 6. 启动服务
systemctl daemon-reload
systemctl enable c2server
systemctl start c2server

# 7. 配置防火墙
ufw allow 8080/tcp
ufw allow 8081/udp
ufw allow 8082/tcp
ufw allow 8053/udp
ufw allow from 10.0.0.0/24 to any port 8090

# 8. 显示服务状态
systemctl status c2server
echo "API认证信息保存在 /opt/c2server/configs/server.json"
```

## 安全检查清单

在部署到生产环境前，请确保完成以下安全检查:

- [ ] 修改所有默认密码和密钥
- [ ] 启用TLS/SSL加密
- [ ] 配置防火墙限制访问
- [ ] 启用API认证
- [ ] 配置日志轮换
- [ ] 设置备份策略
- [ ] 更新系统和依赖到最新版本
- [ ] 验证所有协议的连接性
- [ ] 测试API功能
- [ ] 验证模块加载和执行
- [ ] 检查资源使用情况
- [ ] 配置监控和告警
