# iDB

面向服务器、容器、文件、服务、防火墙规则与常见运维操作的自托管基础设施管理平台。

[English](README.md) | [简体中文](README.zh-CN.md)

[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/sensdb/idb)](https://hub.docker.com/r/sensdb/idb)

iDB 适合开发者和小型团队，用一套控制台统一管理 Linux 主机，而不必引入一整套复杂运维平台。项目由 Web 控制台、控制端 `center` 和宿主机侧执行端 `agent` 组成。

## 功能特性

- 统一管理多台主机
- 实时查看 CPU、内存、磁盘、网络和进程状态
- 提供 Web 终端、文件管理、服务管理
- 支持 Docker 与 Docker Compose 管理
- 基于 `nftables` 的防火墙管理
- SSH 设置与证书管理
- 计划任务、日志查看、文件同步
- 内置常用应用的部署能力

## 架构说明

- `center`：控制平面、API 服务、Web 控制台宿主、任务与日志协调
- `agent`：运行在主机侧的执行组件，负责终端、文件、服务与系统操作
- `frontend`：基于 Vue 的前端控制台
- `plugins`：可选插件与应用集成

## 安装

### 快速安装

直接通过 GitHub 安装：

```bash
sudo curl -fsSL https://raw.githubusercontent.com/sensdata/idb/main/scripts/install.sh | sudo bash
```

显式使用 iDB 提供的镜像/代理源安装：

```bash
curl -fsSL https://dl.idb.net/github-raw/sensdata/idb/main/scripts/install.sh | sudo IDB_GITHUB_PROXY=https://dl.idb.net bash
```

如果 GitHub 访问不稳定，也可以先下载脚本再执行：

```bash
curl -fsSL https://dl.idb.net/github-raw/sensdata/idb/main/scripts/install.sh -o install.sh
sudo IDB_GITHUB_PROXY=https://dl.idb.net bash install.sh
```

说明：

- 安装脚本本身已经带 GitHub 连通性检测，必要时会自动回退到 `https://dl.idb.net`
- 需要 `root` 或 `sudo` 权限执行
- 安装完成后会在终端输出一次初始 `admin` 密码

### Docker 安装

使用 GitHub Release 资源：

```bash
VERSION=$(curl -s https://api.github.com/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
mkdir -p /var/lib/idb && cd /var/lib/idb
curl -fsSL "https://github.com/sensdata/idb/releases/download/${VERSION}/idb.env" -o .env
curl -fsSL "https://github.com/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml" -o docker-compose.yaml
docker compose up -d
```

使用 iDB 镜像/代理源拉取 Release 资源：

```bash
VERSION=$(curl -s https://dl.idb.net/github-api/repos/sensdata/idb/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
mkdir -p /var/lib/idb && cd /var/lib/idb
curl -fsSL "https://dl.idb.net/github-releases/sensdata/idb/releases/download/${VERSION}/idb.env" -o .env
curl -fsSL "https://dl.idb.net/github-releases/sensdata/idb/releases/download/${VERSION}/docker-compose.yaml" -o docker-compose.yaml
docker compose up -d
```

### 从源码构建部署

```bash
git clone --recurse-submodules https://github.com/sensdata/idb.git
cd idb
make deploy
```

`make deploy` 会基于当前工作区代码直接构建并安装本机的 `idb` 与 `idb-agent`。

它会执行：

- 编译前后端产物
- 准备密钥、JWT 材料和证书
- 安装或更新 `center` 服务
- 安装或更新 `agent` 服务

注意：

- `make deploy` 不要求必须处于 `main` 分支，它部署的是你当前 checkout 的代码
- 会覆盖本机已安装的 `idb` / `idb-agent` 服务
- 涉及重要升级时，建议先备份 `/var/lib/idb/data` 与 `/etc/idb`

初始管理员密码同时会写入：

```bash
/etc/idb/idb.env
```

可通过以下命令查看：

```bash
sudo grep '^PASSWORD=' /etc/idb/idb.env
```

## 升级

通过发布版升级脚本升级。GitHub 直连方式：

```bash
curl -fsSL https://raw.githubusercontent.com/sensdata/idb/main/scripts/upgrade.sh -o upgrade.sh
sudo bash upgrade.sh
```

通过 iDB 镜像/代理源升级：

```bash
curl -fsSL https://dl.idb.net/github-raw/sensdata/idb/main/scripts/upgrade.sh -o upgrade.sh
sudo IDB_GITHUB_PROXY=https://dl.idb.net bash upgrade.sh
```

当前升级行为：

- 先升级 `center`
- `center` 重启后会自动检查默认主机上的本机 `agent`
- 如果本机 `agent` 版本落后，会自动升级，不需要用户再手动点击一次主机升级

## 快速开始

1. 浏览器访问控制台。
   默认部署下通常是 `http://your-server-ip:9918`
2. 使用以下信息登录：
   用户名：`admin`
   密码：安装脚本输出的密码，或 `/etc/idb/idb.env` 中记录的密码
3. 在主机管理页面添加远程主机
4. 按需安装或升级主机 `agent`
5. 开始使用文件管理、终端、服务管理、Docker 管理和应用部署功能

## 仓库结构

```text
idb/
├── agent/       # Agent 运行时
├── center/      # Center 控制端与 API
├── core/        # 公共常量、模型、工具
├── frontend/    # Web 前端
├── plugins/     # 插件与应用集成
├── scripts/     # 安装 / 升级 / 卸载脚本
└── README.md
```

## 开发

### 环境要求

- Go 1.25+
- Node.js 24+
- npm

### 后端开发

```bash
cd center
go run main.go start
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

### 测试源码部署

如果要在测试机验证某个分支：

```bash
git fetch origin
git switch your-branch
git pull
make deploy
```

建议部署后至少检查：

```bash
sudo systemctl status idb --no-pager -l
sudo systemctl status idb-agent --no-pager -l
sudo journalctl -u idb -n 100 --no-pager
sudo tail -n 100 /var/log/idb-agent/idb-agent.log
```

如果要确认当前部署的是哪个提交：

```bash
git branch --show-current
git rev-parse --short HEAD
```

## 参与贡献

欢迎提交 Issue 和 Pull Request。

基本流程：

1. Fork 仓库
2. 创建功能分支
3. 提交聚焦、清晰的改动
4. 在合适的地方补充或更新测试
5. 发起 Pull Request

约定建议：

- 遵循 Go 标准格式和现有项目风格
- 保持提交粒度清晰
- 功能行为发生变化时同步更新文档

## 文档

- 官网：https://idb.net
- 文档：https://idb.net/docs
- API 文档：`http://your-server-ip:9918/api/v1/swagger/index.html`
- 插件仓库：https://github.com/sensdata/idb-plugins

重新生成 Swagger 文档：

```bash
cd center
go generate .
```

## 许可证

本项目基于 Apache License 2.0 发布。详见 [LICENSE](LICENSE)。
