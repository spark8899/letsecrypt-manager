# LetsEncrypt Manager

一个基于 Go 的 Let's Encrypt 泛域名证书申请管理系统，提供 REST API 接口。

## 功能

- JWT 认证（支持多账户，密码 SHA256 加密）
- 泛域名证书申请（`*.example.com` + `example.com`）
- DNS-01 验证流程（CNAME/TXT 记录）
- 证书信息查询（fullchain.cer + 私钥）
- 本地文件持久化（域名信息、证书）
- 请求日志 + 文件日志

## 目录结构

```
letsencrypt-manager/
├── main.go                 # 程序入口
├── config.json             # 配置文件
├── go.mod
├── Makefile
├── config/
│   └── config.go           # 配置加载
├── logger/
│   └── logger.go           # 日志系统
├── middleware/
│   └── auth.go             # JWT 认证中间件
├── models/
│   └── store.go            # 数据存储（文件系统）
├── acme/
│   └── client.go           # ACME/lego 客户端
├── handlers/
│   └── handlers.go         # HTTP 接口处理器
└── data/                   # 运行时数据目录（自动创建）
    ├── domains/            # 域名 JSON 文件
    ├── certs/              # 证书文件
    │   └── example.com/
    │       ├── fullchain.cer
    │       └── private.key
    ├── accounts/           # ACME 账号信息
    └── logs/               # 日志文件
```

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 编译

```bash
make build
```

### 3. 配置 config.json

```json
{
  "listen_addr": ":8080",
  "data_dir": "./data",
  "acme_email": "your@email.com",
  "acme_server": "staging",
  "jwt_secret": "your-random-secret",
  "token_expiry_hours": 24,
  "accounts": [
    {
      "username": "admin",
      "password": "sha256-hashed-password"
    }
  ]
}
```

**生成密码哈希：**
```bash
make hash-password PASSWORD=yourpassword
# 或
echo -n "yourpassword" | sha256sum
```

**acme_server 设置：**
- `"staging"` — 使用 Let's Encrypt 测试环境（推荐调试时使用）
- `"production"` — 使用 Let's Encrypt 生产环境

### 4. 运行

```bash
./letsencrypt-manager
# 或
make run
```

---

## API 接口

### 认证

#### POST /api/auth/login

获取访问 Token（有效期默认 24 小时）

**请求：**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应：**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400,
  "token_type": "Bearer"
}
```

后续所有接口需在 Header 中携带：
```
Authorization: Bearer <token>
```

---

### 域名管理

#### POST /api/domains

添加泛域名（输入 `abcd.com` 或 `*.abcd.com` 均可）

**请求：**
```json
{
  "domain": "abcd.com"
}
```

**响应：**
```json
{
  "domain": "abcd.com",
  "status": "pending",
  "message": "domain added successfully"
}
```

#### GET /api/domains

列出所有域名

---

### ACME 证书工作流

#### POST /api/domains/:domain/dns-challenge

向 Let's Encrypt 申请 DNS-01 验证信息，返回需要设置的 DNS 记录。

**响应：**
```json
{
  "domain": "abcd.com",
  "status": "verifying",
  "challenge": {
    "type": "dns-01",
    "record_name": "_acme-challenge.abcd.com",
    "cname_target": "_acme-challenge.abcd.com.acme.example.com.",
    "instructions": "Add a CNAME record: _acme-challenge.abcd.com -> ..."
  }
}
```

收到响应后，在 DNS 面板中添加对应的 CNAME 或 TXT 记录。

#### GET /api/domains/:domain/dns-verify

验证 DNS 解析是否已生效。

**响应：**
```json
{
  "domain": "abcd.com",
  "verified": true,
  "detail": "CNAME: _acme-challenge.abcd.com -> ...",
  "status": "verified"
}
```

#### POST /api/domains/:domain/issue

向 Let's Encrypt 申请并颁发证书（`*.abcd.com` + `abcd.com`）。

**响应：**
```json
{
  "domain": "abcd.com",
  "status": "issued",
  "cert_expiry": "2024-06-01T00:00:00Z",
  "message": "Certificate for *.abcd.com and abcd.com issued successfully"
}
```

#### GET /api/domains/:domain/cert

获取已颁发的证书内容。

**响应：**
```json
{
  "domain": "abcd.com",
  "status": "issued",
  "cert_expiry": "2024-06-01T00:00:00Z",
  "fullchain_cer": "-----BEGIN CERTIFICATE-----\n...",
  "private_key": "-----BEGIN EC PRIVATE KEY-----\n..."
}
```

---

## 完整使用示例

```bash
# 1. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .token)

# 2. 添加域名
curl -X POST http://localhost:8080/api/domains \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"domain":"abcd.com"}'

# 3. 获取 DNS 验证信息
curl -X POST http://localhost:8080/api/domains/abcd.com/dns-challenge \
  -H "Authorization: Bearer $TOKEN"

# 4. （在 DNS 面板设置 CNAME/TXT 记录后）验证 DNS
curl http://localhost:8080/api/domains/abcd.com/dns-verify \
  -H "Authorization: Bearer $TOKEN"

# 5. 申请证书
curl -X POST http://localhost:8080/api/domains/abcd.com/issue \
  -H "Authorization: Bearer $TOKEN"

# 6. 获取证书内容
curl http://localhost:8080/api/domains/abcd.com/cert \
  -H "Authorization: Bearer $TOKEN"
```

---

## 添加多账户

在 config.json 的 accounts 数组中添加多个账户：

```json
{
  "accounts": [
    {
      "username": "admin",
      "password": "sha256-of-admin-password"
    },
    {
      "username": "ops",
      "password": "sha256-of-ops-password"
    }
  ]
}
```

## 运行测试

```bash
# 运行全部测试
make test

# 详细输出
make test-verbose

# 生成覆盖率报告（HTML）
make test-cover

# 竞态检测
make test-race

# 指定包
make test-pkg PKG=./handlers
```

主要测试场景

认证（Login）：正确密码✅、错误密码❌、未知用户❌、缺少字段❌

域名管理：正常添加、通配符前缀自动剥离（*.a.com → a.com）、大写归一化、重复添加返回 409、无 token 返回 401

DNS 验证：无 challenge 时返回 400、域名不存在返回 404、challenge 注入后返回正确记录名

证书查询：未颁发返回 404、注入证书后正确返回 fullchain_cer、private_key、cert_expiry

端到端工作流：模拟完整链路（添加域名 → 注入 Challenge → 验证 DNS → 注入证书 → 查询证书）

## 注意事项

- 默认使用 **Let's Encrypt Staging 环境**，颁发的证书不受浏览器信任，仅用于测试
- 生产环境请将 `acme_server` 改为 `"production"`，并确保 DNS 记录正确设置
- Let's Encrypt 对 IP 有速率限制，同一域名每周最多申请 5 次
- 私钥文件权限为 0600，请妥善保管

/dns-challenge                    /issue
     │                               │
     │  Obtain() ──────────────────► goroutine 跑完
     │       │                            │
     │  Present() 被调用                  │
     │       ├─ challenge info → ready    │
     │       └─ 阻塞在 <-proceed          │
     │                               close(proceed)
     │                                    │
     │                               Present() 返回
     │                                    │
     │                               lego 做 DNS 验证
     │                                    │
     │                               签发证书 → resultCh
     │                               │
     └───────────────────────────────┘
                                     ProceedWithOrder 从 resultCh 读取
