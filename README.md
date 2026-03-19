# iDB

**基于 Go 语言构建的轻量级自托管运维平台**  
为开发者和小型团队而设计：一站式服务器管理与数据库快速部署工具

[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/sensdb/idb)](https://hub.docker.com/r/sensdb/idb)

---

## ✨ 特性

### 服务器管理
- **系统监控**：实时查看 CPU、内存、磁盘、网络等资源占用
- **进程管理**：查看和管理系统进程
- **文件管理**：浏览器式文件管理，支持上传、下载、编辑、压缩等操作
- **服务管理**：管理系统服务，支持启动、停止、重启等操作

### 网络与安全
- **防火墙管理**：基于 nftables 的防火墙配置
- **SSH 管理**：SSH 配置、密钥管理、登录日志
- **证书管理**：生成和管理 SSL/TLS 证书

### 容器管理
- **Docker 管理**：容器、镜像、网络、卷的完整管理
- **Docker Compose**：支持 Docker Compose 应用管理
- **一键部署**：快速部署 MySQL、PostgreSQL、Redis 等常用服务

### 工具集
- **日志管理**：日志查看、搜索、实时跟踪
- **计划任务**：定时任务管理
- **命令终端**：Web 终端，支持多主机切换
- **文件同步**：基于 rsync 的文件同步功能

### 多主机管理
- **统一管理**：支持管理多台服务器
- **跨主机操作**：支持跨服务器的文件传输、命令执行

---

## 📦 安装

### 一键安装
```bash
sudo curl -fsSL https://raw.githubusercontent.com/sensdata/idb/main/scripts/install.sh | sudo bash

```

### Docker 安装
```bash
# 下载配置文件
VERSION=$(curl -s https://api.github.com/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
mkdir -p /var/lib/idb && cd /var/lib/idb
curl -fsSL "https://github.com/sensdata/idb/releases/download/${VERSION}/idb.env" -o .env
curl -fsSL "https://github.com/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml" -o docker-compose.yaml

# 启动
docker compose up -d
```

### 手动编译安装

1. 克隆仓库
```bash
git clone --recurse-submodules https://github.com/sensdata/idb.git
cd idb
```

2. 一键编译并安装
```bash
make deploy
```

`make deploy` 会自动执行：
- 编译前端并产出 `frontend/dist`
- 生成并分发 key / jwt-key / certs（已存在则复用）
- 安装并重启 `center` 服务
- 安装并重启 `agent` 服务

首次部署时会生成管理员密码并在终端输出一次，同时写入：
`/etc/idb/idb.env`（`PASSWORD=...`，权限 `600`）

如果错过首次输出，可执行：
```bash
sudo grep '^PASSWORD=' /etc/idb/idb.env
```

---

## 🚀 快速开始

1. **访问控制台**
   - 安装完成后，访问 `http://your-server-ip:9918`
   - 默认用户名：`admin`
   - 密码：
     - 一键脚本/Docker 安装：安装完成后的终端输出可见
     - 手动编译安装（`make deploy`）：首次部署时终端输出，并可在 `/etc/idb/idb.env` 查看 `PASSWORD`

2. **添加服务器**
   - 登录后，点击左侧菜单栏的「主机管理」
   - 点击「添加主机」，填写服务器信息
   - 选择安装方式（一键安装或手动安装）

3. **管理服务器**
   - 服务器添加成功后，即可在控制台中管理该服务器
   - 点击服务器卡片，进入服务器详情页面
   - 可查看系统信息、管理进程、文件、服务等

4. **部署应用**
   - 点击左侧菜单栏的「应用市场」
   - 选择要部署的应用（如 MySQL）
   - 填写配置信息，点击「部署」

---

## 📁 项目结构

```
idb/
├── agent/           # Agent 客户端代码
│   ├── agent/       # Agent 核心实现
│   ├── config/      # 配置管理
│   ├── db/          # 本地数据库
│   └── main.go      # Agent 入口
├── center/          # Center 服务器代码
│   ├── config/      # 配置管理
│   ├── core/        # 核心功能实现
│   ├── db/          # 数据库模型和仓库
│   ├── plugin/      # 插件系统
│   └── main.go      # Center 入口
├── core/            # 公共库和定义
│   ├── constant/    # 常量定义
│   ├── model/       # 数据模型
│   └── utils/       # 工具函数
├── plugins/         # 插件集合（git submodule → sensdata/idb-plugins）
│   ├── mysqlmanager/
│   ├── postgresql/
│   ├── redis/
│   ├── pma/
│   ├── rsync/
│   └── scriptmanager/
├── frontend/        # 前端代码
│   ├── src/         # 前端源码
│   └── package.json # 前端依赖
├── scripts/         # 安装/升级/卸载脚本
├── LICENSE          # 许可证文件
└── README.md        # 项目说明
```

---

## 🔧 开发环境

### 后端开发

1. **环境要求**
   - Go 1.25+ 

2. **启动开发服务器**
   ```bash
   cd center
   go run main.go start
   ```

### 前端开发

1. **环境要求**
   - Node.js 18+ 
   - npm 或 yarn

2. **安装依赖**
   ```bash
   cd frontend
   npm install
   ```

3. **启动开发服务器**
   ```bash
   npm run dev
   ```

---

## 🤝 贡献

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发规范

- 代码风格：遵循 Go 官方代码风格
- 提交信息：清晰、简洁，使用英文
- 测试：为新功能添加测试
- 文档：更新相关文档

### 代码结构

- 新增功能请遵循现有代码结构
- 核心功能添加到 `core/` 目录
- 服务器功能添加到 `center/` 目录
- 客户端功能添加到 `agent/` 目录

---

## 📄 许可证

本项目采用 Apache License 2.0 许可证。详情请查看 [LICENSE](LICENSE) 文件。

---

## 📚 文档

- **官网**：https://idb.net
- **项目文档**：https://idb.net/docs
- **API 文档**：部署后访问 `http://your-server-ip:9918/api/v1/swagger/index.html`
- **插件仓库**：[sensdata/idb-plugins](https://github.com/sensdata/idb-plugins)

重新生成 Swagger 文档：在 `center/` 目录执行 `go generate .`

---

## 📬 联系

- 技术支持：support@idb.net
- 问题反馈：[GitHub Issues](https://github.com/sensdata/idb/issues)
- 讨论交流：[GitHub Discussions](https://github.com/sensdata/idb/discussions)

---

## 📊 数据

- **Docker 下载量**：1 万+
- **支持系统**：Ubuntu、Debian、CentOS、Rocky Linux、Kylin、UnionTech 等
- **云平台支持**：AWS、Azure、Google Cloud、阿里云、腾讯云、华为云、UCloud 等

---

## 🔗 相关链接

- [官网](https://idb.net)
- [文档](https://idb.net/docs)
- [Docker Hub](https://hub.docker.com/r/sensdb/idb)
- [GitHub](https://github.com/sensdata/idb)

---

**感谢使用 iDB！** 🎉
