export default {
  'app.nftables.title': 'NFTables 防火墙管理',
  'app.nftables.description':
    '管理服务器防火墙规则配置，支持全局和本地两种配置模式',

  // 菜单项
  'app.nftables.menu.config': '应用配置',
  'app.nftables.menu.ports': '端口配置',
  'app.nftables.menu.ipBlacklist': 'IP黑名单',
  'app.nftables.menu.ping': 'Ping配置',

  // 页面标题
  'app.nftables.ports.pageTitle': '端口配置',
  'app.nftables.ipBlacklist.pageTitle': 'IP黑名单管理',
  'app.nftables.ping.pageTitle': 'Ping配置',

  // 防火墙类型选择
  'app.nftables.firewall.title': '防火墙类型',
  'app.nftables.firewall.nftables.desc':
    '现代化的防火墙管理工具，支持复杂的规则配置和高级功能',
  'app.nftables.firewall.iptables.desc':
    '传统的防火墙管理工具，广泛支持但功能相对简单',
  'app.nftables.firewall.iptables.notSupported': '不支持 IPTables 配置',
  'app.nftables.firewall.iptables.notSupportedDesc':
    '本系统仅支持 NFTables 的可视化配置管理。如需使用 IPTables，请通过终端或其他工具进行配置。',

  // 服务管理
  'app.nftables.service.title': 'NFTables 服务状态',
  'app.nftables.service.control': '服务控制',

  // 按钮
  'app.nftables.button.refresh': '刷新状态',
  'app.nftables.button.install': '立即安装',
  'app.nftables.button.start': '启动服务',
  'app.nftables.button.stop': '停止服务',
  'app.nftables.button.restart': '重启服务',
  'app.nftables.button.test': '测试配置',
  'app.nftables.button.load': '加载配置',
  'app.nftables.button.save': '保存配置',
  'app.nftables.button.apply': '应用配置',
  'app.nftables.button.format': '格式化',
  'app.nftables.button.history': '历史记录',
  'app.nftables.button.switchToNftables': '切换到 NFTables',
  'app.nftables.button.switchToIptables': '切换到 IPTables',
  'app.nftables.button.activateConfig': '激活此配置',
  'app.nftables.button.addRule': '新建规则',
  'app.nftables.button.configPort': '配置端口',
  'app.nftables.button.scan': '扫描',
  'app.nftables.button.addIP': '添加IP',
  'app.nftables.button.configIPBlacklist': '配置IP黑名单',

  // 状态
  'app.nftables.status.installed': '已安装',
  'app.nftables.status.notInstalled': '未安装',
  'app.nftables.status.running': '运行状态',
  'app.nftables.status.configured': '配置状态',
  'app.nftables.status.ruleActive': '规则已生效',
  'app.nftables.status.needConfig': '需要创建配置',
  'app.nftables.status.stopped': '已停止',
  'app.nftables.status.unknown': '未知',
  'app.nftables.status.nftablesActive': 'NFTables 已激活',
  'app.nftables.status.iptablesLegacyActive': 'IPTables (传统) 已激活',
  'app.nftables.status.iptablesNftActive': 'IPTables-NFT (兼容层) 已激活',
  'app.nftables.status.noFirewall': '无防火墙系统',
  'app.nftables.status.uncertain': '状态不确定',
  'app.nftables.status.currentActive': '当前激活',
  'app.nftables.status.inactive': '非激活状态',
  'app.nftables.status.stop': '停止',
  'app.nftables.status.reload': '重载',
  'app.nftables.status.restart': '重启',
  'app.nftables.status.starting': '启动中',
  'app.nftables.status.stopping': '停止中',
  'app.nftables.status.error': '错误',
  'app.nftables.status.unhealthy': '异常',
  'app.nftables.status.autoStartEnabled': '自动启动已启用',
  'app.nftables.status.autoStartDisabled': '自动启动已禁用',

  // 安装指南
  'app.nftables.installation.title': '安装 NFTables',
  'app.nftables.installation.step1.title': '检查系统依赖',
  'app.nftables.installation.step1.desc':
    '检查系统是否具备安装 NFTables 的必要条件',
  'app.nftables.installation.step2.title': '下载并安装',
  'app.nftables.installation.step2.desc': '从软件仓库下载并安装 NFTables 包',
  'app.nftables.installation.step3.title': '配置服务',
  'app.nftables.installation.step3.desc': '启用 NFTables 服务并进行基本配置',
  'app.nftables.installation.log.title': '安装日志',
  'app.nftables.installation.log.clear': '清除',

  // 配置管理
  'app.nftables.config.title': '应用配置',
  'app.nftables.config.firewallStatus': '防火墙状态',
  'app.nftables.config.installStatus': '安装状态',
  'app.nftables.config.activeSystem': '当前激活系统',
  'app.nftables.config.modeChanged': '已切换到 {mode} 模式',
  'app.nftables.config.mode': '配置模式',
  'app.nftables.config.configScope': '配置范围',
  'app.nftables.config.editMode': '编辑模式',

  // 可视化配置
  'app.nftables.config.visual.title': '可视化配置',
  'app.nftables.config.visual.desc': '通过表单界面配置防火墙规则，操作简单直观',
  'app.nftables.config.visual.developing': '功能开发中',
  'app.nftables.config.visual.developingDesc':
    '可视化配置界面正在开发中，请暂时使用文件模式进行配置。',

  // 文件配置
  'app.nftables.config.file.title': '文件配置',
  'app.nftables.config.file.desc': '直接编辑 NFTables 配置文件，适合高级用户',
  'app.nftables.config.file.exists': '已存在',
  'app.nftables.config.file.new': '新建',

  // 表单配置
  'app.nftables.config.form.title': '表单配置',
  'app.nftables.config.form.desc': '通过表单界面配置端口规则，操作简单直观',
  'app.nftables.form.portHelp': '单个端口：80，端口列表：80, 443, 9918',
  'app.nftables.form.portRangeHelp': '支持端口段，例如 8000-9000',

  'app.nftables.form.updateRule': '更新规则',
  'app.nftables.form.cancel': '取消',
  'app.nftables.form.sourceOptional': '来源（可选）',
  'app.nftables.form.sourcePlaceholder': '例如: 192.168.1.0/24 或 0.0.0.0/0',

  // 高级规则
  'app.nftables.form.advancedRules': '高级规则',
  'app.nftables.form.rulesList': '规则列表',
  'app.nftables.form.addRule': '添加规则',
  'app.nftables.form.noRules': '暂无规则',
  'app.nftables.form.rule': '规则',
  'app.nftables.form.ruleType': '规则类型',
  'app.nftables.form.basicRule': '基本规则',
  'app.nftables.form.rateLimit': '速率限制',
  'app.nftables.form.concurrentLimit': '并发限制',
  'app.nftables.form.rateValue': '速率值',
  'app.nftables.form.ratePlaceholder': '例如: 100/second',
  'app.nftables.form.concurrentCount': '并发数量',
  'app.nftables.form.concurrentPlaceholder': '例如: 10',
  'app.nftables.form.srcIp': '源IP限制',
  'app.nftables.form.srcIpPlaceholder': '例如: 192.168.1.100 或 192.168.1.0/24',
  'app.nftables.form.srcIpHelp':
    '指定允许访问此端口的IP地址或网段，留空表示允许所有IP',

  // 配置类型
  'app.nftables.config.type.local': '本地',
  'app.nftables.config.type.global': '全局',
  'app.nftables.config.globalDescription': '全局配置将应用到所有主机',
  'app.nftables.config.localDescription': '本地配置仅应用到当前主机',
  'app.nftables.config.activateHint':
    '当前查看的是 {viewing} 配置，但激活的是 {active} 配置',

  // 控制栏
  'nftables.controlBar.title': '配置管理',
  'nftables.controlBar.currentStatus':
    '当前配置: {configType}, 当前模式: {configMode}',

  // 配置类型
  'nftables.configType.local': '本地',
  'nftables.configType.global': '全局',

  // 配置模式
  'nftables.configMode.form': '表单',
  'nftables.configMode.file': '文件',

  // 进程状态
  'app.nftables.config.processStatusList': '进程状态列表',
  'app.nftables.config.columns.process': '进程',
  'app.nftables.config.columns.port': '端口',
  'app.nftables.config.columns.addresses': '监听地址',
  'app.nftables.config.accessible': '可访问',
  'app.nftables.config.fullyAccessible': '完全可访问',
  'app.nftables.config.rejected': '已拒绝',
  'app.nftables.config.restricted': '受限',
  'app.nftables.config.unknown': '未知',
  'app.nftables.config.notAccessible': '受限',
  'app.nftables.config.localOnly': '仅本地',

  // 配置模式
  'app.nftables.config.formMode': '表单模式',
  'app.nftables.config.visualMode': '可视化模式',
  'app.nftables.config.fileMode': '文件模式',

  // 基础策略
  'app.nftables.config.policy.title': '基础策略',
  'app.nftables.config.policy.input': '输入策略',
  'app.nftables.config.policy.output': '输出策略',
  'app.nftables.config.policy.accept': 'ACCEPT (允许)',
  'app.nftables.config.policy.drop': 'DROP (丢弃)',

  // 端口规则
  'app.nftables.config.rules.title': '端口规则',
  'app.nftables.config.rules.protocol': '协议',
  'app.nftables.config.rules.port': '端口',
  'app.nftables.config.rules.action': '动作',
  'app.nftables.config.rules.source': '来源',
  'app.nftables.config.rules.description': '描述',
  'app.nftables.config.rules.allow': '允许',
  'app.nftables.config.rules.deny': '拒绝',
  'app.nftables.config.rules.unknown': '未知',
  'app.nftables.config.rules.portPlaceholder': '80',
  'app.nftables.config.rules.descPlaceholder': '规则描述',
  'app.nftables.config.rules.deleteConfirm': '确定要删除端口 {port} 的规则吗？',
  'app.nftables.config.rules.empty': '暂无端口规则',

  // Ping配置
  'app.nftables.ping.title': 'Ping配置',
  'app.nftables.ping.description': '配置服务器是否响应ICMP ping请求',
  'app.nftables.ping.currentStatus': '当前状态',
  'app.nftables.ping.allowed': 'Ping允许',
  'app.nftables.ping.blocked': 'Ping阻止',
  'app.nftables.ping.allowPing': '允许Ping',
  'app.nftables.ping.blockPing': '阻止Ping',
  'app.nftables.ping.statusDescription.allowed':
    '服务器当前响应来自外部的ICMP ping请求。',
  'app.nftables.ping.statusDescription.blocked':
    '服务器当前阻止来自外部的ICMP ping请求。',
  'app.nftables.ping.configureTitle': 'Ping设置',
  'app.nftables.ping.enableLabel': '允许外部ping请求',
  'app.nftables.ping.enableHelp':
    '启用时，服务器将响应ICMP ping请求。禁用时，ping请求将被防火墙阻止。',
  'app.nftables.ping.applySettings': '应用设置',
  'app.nftables.ping.loading': '正在加载ping状态...',
  'app.nftables.ping.saving': '正在保存设置...',
  'app.nftables.ping.saveSuccess': 'Ping设置更新成功',
  'app.nftables.ping.saveFailed': '更新Ping设置失败',
  'app.nftables.ping.loadFailed': '加载Ping状态失败',

  // 编辑器
  'app.nftables.config.editor.placeholder': '请输入 NFTables 配置...',
  'app.nftables.config.editor.lineCount': '{count} 行',
  'app.nftables.config.editor.reloadTip':
    '提示：配置保存后需要重新加载防火墙规则才能生效',

  // 模式切换
  'app.nftables.config.modeSwitch.form': '已切换到表单模式',
  'app.nftables.config.modeSwitch.file': '已切换到文件模式',

  // 编辑器
  'app.nftables.editor.placeholder': '请输入 nftables 配置...',

  // 通用状态
  'app.nftables.common.installed': '已安装',
  'app.nftables.common.notInstalled': '未安装',
  'app.nftables.common.running': '运行中',
  'app.nftables.common.stopped': '已停止',
  'app.nftables.common.configured': '已配置',
  'app.nftables.common.notConfigured': '待配置',

  // Tab标签
  'app.nftables.tabs.overview': '总览',
  'app.nftables.tabs.service': '服务管理',
  'app.nftables.tabs.config': '配置管理',
  'app.nftables.tabs.monitoring': '监控日志',

  // 总览页面
  'app.nftables.overview.status': '系统状态',
  'app.nftables.overview.quickActions': '快速操作',
  'app.nftables.overview.startConfig': '开始配置',
  'app.nftables.overview.editConfig': '编辑配置',

  // 监控日志
  'app.nftables.monitoring.title': '监控中心',
  'app.nftables.monitoring.developing': '功能开发中',
  'app.nftables.monitoring.developingDesc':
    '监控和日志功能正在开发中，敬请期待。',

  // 端口配置
  'app.nftables.ports.title': '端口配置',

  // IP黑名单配置
  'app.nftables.ipBlacklist.title': 'IP黑名单',
  'app.nftables.ipBlacklist.description': '管理IP黑名单规则，阻止恶意IP访问',
  'app.nftables.ipBlacklist.form.title': '可视化配置',
  'app.nftables.ipBlacklist.form.desc':
    '通过表单界面配置IP黑名单规则，操作简单直观',
  'app.nftables.ipBlacklist.file.title': '文件配置',
  'app.nftables.ipBlacklist.file.desc': '直接编辑NFTables IP黑名单配置文件',
  'app.nftables.ipBlacklist.rules.title': 'IP黑名单规则',
  'app.nftables.ipBlacklist.rules.empty': '暂无IP黑名单规则',
  'app.nftables.ipBlacklist.rules.ip': 'IP地址',
  'app.nftables.ipBlacklist.rules.type': '类型',
  'app.nftables.ipBlacklist.rules.description': '描述',
  'app.nftables.ipBlacklist.rules.action': '动作',
  'app.nftables.ipBlacklist.rules.createdAt': '创建时间',
  'app.nftables.ipBlacklist.rules.ipPlaceholder':
    '192.168.1.100 或 192.168.1.0/24',
  'app.nftables.ipBlacklist.rules.descPlaceholder': '规则描述',
  'app.nftables.ipBlacklist.rules.deleteConfirm':
    '确定要删除IP {ip} 的黑名单规则吗？',
  'app.nftables.ipBlacklist.rules.dropHint':
    '所有添加到黑名单的IP地址都将被丢弃（DROP），阻止其访问服务器',
  'app.nftables.ipBlacklist.rules.invalidIPFormat': '请输入正确的IP地址格式',
  'app.nftables.ipBlacklist.rules.ipFormatHint':
    '支持单个IP（192.168.1.100）、CIDR网段（192.168.1.0/24）或IP范围（192.168.1.1-192.168.1.10）',
  'app.nftables.ipBlacklist.type.single': '单个IP',
  'app.nftables.ipBlacklist.type.cidr': 'CIDR网段',
  'app.nftables.ipBlacklist.type.range': 'IP范围',
  'app.nftables.ipBlacklist.action.drop': 'DROP (丢弃)',
  'app.nftables.ipBlacklist.action.reject': 'REJECT (拒绝)',

  // 策略和规则
  'app.nftables.policy.title': '基础策略',
  'app.nftables.rules.title': '端口规则',
  'app.nftables.rules.empty': '暂无规则',
  'app.nftables.rules.deleteConfirm': '确定要删除端口 {port} 的规则吗？',
  'app.nftables.services.title': '应用服务',
  'app.nftables.templates.title': '预设模板',

  // 通用文本
  'app.nftables.loading': '加载中...',
  'app.nftables.savingChanges': '正在保存更改...',

  // 消息提示
  'app.nftables.message.statusRefreshed': '状态刷新完成',
  'app.nftables.message.refreshFailed': '状态刷新失败',
  'app.nftables.message.fetchSuccess': '进程数据加载成功，共{count}条记录',
  'app.nftables.message.fetchFailed': '数据获取失败',
  'app.nftables.message.statusFetchSuccess': '防火墙状态获取成功',
  'app.nftables.message.statusFetchFailed': '防火墙状态获取失败',
  'app.nftables.message.switchToNftablesSuccess': '成功切换到 NFTables',
  'app.nftables.message.switchToIptablesSuccess': '成功切换到 IPTables',
  'app.nftables.message.switchFailed': '防火墙系统切换失败',
  'app.nftables.message.installingNftables': '正在安装 NFTables...',
  'app.nftables.message.installSuccess': 'NFTables 安装成功',
  'app.nftables.message.installFailed': 'NFTables 安装失败',
  'app.nftables.message.serviceStarted': '服务启动成功',
  'app.nftables.message.startFailed': '服务启动失败',
  'app.nftables.message.serviceStopped': '服务停止成功',
  'app.nftables.message.stopFailed': '服务停止失败',
  'app.nftables.message.serviceRestarted': '服务重启成功',
  'app.nftables.message.restartFailed': '服务重启失败',
  'app.nftables.message.configReloaded': '配置重载成功',
  'app.nftables.message.reloadFailed': '配置重载失败',
  'app.nftables.message.configLoaded': '配置加载成功',
  'app.nftables.message.loadFailed': '配置加载失败',
  'app.nftables.message.configSaved': '配置保存成功',
  'app.nftables.message.saveFailed': '配置保存失败',
  'app.nftables.message.testPassed': '配置测试通过',
  'app.nftables.message.testFailed': '配置测试失败',
  'app.nftables.message.configFormatted': '配置格式化完成',
  'app.nftables.message.formatFailed': '配置格式化失败',
  'app.nftables.message.configApplied': '配置应用成功',
  'app.nftables.message.applyFailed': '配置应用失败',
  'app.nftables.message.initializingDefaultConfig': '正在初始化默认配置...',
  'app.nftables.message.defaultConfigInitialized': '默认配置已创建',
  'app.nftables.message.defaultConfigInitFailed': '默认配置初始化失败',
  'app.nftables.message.ruleAdded': '规则添加成功',
  'app.nftables.message.ruleUpdated': '规则更新成功',
  'app.nftables.message.ruleDeleted': '规则删除成功',
  'app.nftables.message.configRefreshed': '配置刷新成功',
  'app.nftables.message.operationFailed': '操作失败',
  'app.nftables.message.actionRequired': '请选择动作',
  'app.nftables.message.protocolRequired': '请选择协议类型',

  'app.nftables.message.batchNotSupported':
    '当前版本暂不支持批量端口提交，请选择单个端口',

  'app.nftables.message.noHost': '没有可用的主机',

  // 编辑器相关
  'app.nftables.config.editor.modified': '已修改',
  'app.nftables.config.editor.tips': '按 Ctrl+S 保存，或使用上方的保存按钮',
  'app.nftables.config.editor.emptyContent': '配置内容不能为空',
  'app.nftables.config.editor.unsavedChanges':
    '您有未保存的更改。确定要离开吗？',
  'app.nftables.config.editor.confirmRefresh': '确认刷新',
  'app.nftables.config.editor.confirmRefreshContent':
    '您有未保存的更改。刷新将丢弃这些更改。是否继续？',

  // 错误信息
  'app.nftables.error.fetchConfigFailed': '获取配置失败',

  // 进程描述
  'app.nftables.process.unknown': '未知进程',
  'app.nftables.process.nginx': 'Nginx 高性能Web服务器',
  'app.nftables.process.mysql': 'MySQL 关系型数据库',
  'app.nftables.process.mysqld': 'MySQL 数据库守护进程',
  'app.nftables.process.postgresql': 'PostgreSQL 开源数据库',
  'app.nftables.process.postgres': 'PostgreSQL 数据库服务',
  'app.nftables.process.redis': 'Redis 内存键值数据库',
  'app.nftables.process.mongodb': 'MongoDB 文档型数据库',
  'app.nftables.process.apache': 'Apache HTTP Web服务器',
  'app.nftables.process.httpd': 'Apache HTTP 守护进程',
  'app.nftables.process.docker': 'Docker 容器平台',
  'app.nftables.process.dockerd': 'Docker 守护进程',
  'app.nftables.process.dockerProxy': 'Docker 网络代理',
  'app.nftables.process.dockerContainerd': 'Docker containerd 运行时',
  'app.nftables.process.containerd': 'containerd 容器运行时',
  'app.nftables.process.containerdShim': 'containerd shim 进程',
  'app.nftables.process.kubernetes': 'Kubernetes 容器编排系统',
  'app.nftables.process.kubectl': 'Kubernetes 命令行工具',
  'app.nftables.process.ssh': 'SSH 安全Shell连接',
  'app.nftables.process.sshd': 'SSH 服务器守护进程',
  'app.nftables.process.systemd': '系统和服务管理器',
  'app.nftables.process.nodejs': 'Node.js JavaScript运行时',
  'app.nftables.process.node': 'Node.js 应用程序',
  'app.nftables.process.java': 'Java 虚拟机应用',
  'app.nftables.process.tomcat': 'Apache Tomcat Web服务器',
  'app.nftables.process.php': 'PHP 脚本解释器',
  'app.nftables.process.phpFpm': 'PHP FastCGI 进程管理器',
  'app.nftables.process.podman': 'Podman 无守护进程容器引擎',
  'app.nftables.process.mariadb': 'MariaDB 数据库服务器',
  'app.nftables.process.bind': 'BIND DNS 服务器',
  'app.nftables.process.named': 'BIND DNS 守护进程',
  'app.nftables.process.fail2ban': 'Fail2ban 入侵防护系统',
  'app.nftables.process.iptables': 'iptables 防火墙规则',
  'app.nftables.process.firewall': '防火墙服务',
  'app.nftables.process.selinux': 'SELinux 安全模块',

  // 进程分类描述
  'app.nftables.process.category.database': '数据库服务',
  'app.nftables.process.category.webServer': 'Web服务器',
  'app.nftables.process.category.container': '容器化服务',
  'app.nftables.process.category.network': '网络服务',
  'app.nftables.process.category.security': '安全服务',
  'app.nftables.process.category.development': '开发工具',
  'app.nftables.process.category.system': '系统服务',
  'app.nftables.process.category.application': '应用程序',

  // 表单相关
  'app.nftables.form.source': '源地址',
  'app.nftables.form.basicConfig': '基本配置',
  'app.nftables.form.configMode': '配置模式',
  'app.nftables.form.simpleMode': '简单模式',
  'app.nftables.form.simpleModeDesc':
    '适合新手，只需选择端口和动作即可快速配置',
  'app.nftables.form.advancedMode': '高级模式',
  'app.nftables.form.advancedModeDesc':
    '适合高级用户，支持速率限制、并发限制等复杂规则',
  'app.nftables.form.accessControl': '访问控制',
  'app.nftables.form.allowDesc': '允许访问此端口',
  'app.nftables.form.denyDesc': '拒绝访问此端口，不返回任何响应',
  'app.nftables.form.rejectDesc': '拒绝访问此端口，返回拒绝响应',
  'app.nftables.form.noRulesHint': '还没有配置任何高级规则',
  'app.nftables.form.addFirstRule': '添加第一个规则',
  'app.nftables.form.configPreview': '配置预览',
  'app.nftables.form.noRulesPreview': '# 暂无规则配置',
  'app.nftables.form.rateHelpText':
    '格式：数量/时间单位，如 100/second, 50/minute',
  'app.nftables.form.concurrentHelpText': '同时允许的最大连接数',

  // 表单验证
  'app.nftables.validation.portOrRange':
    '请输入端口或端口段，例如 8080 或 8000-9000',

  'app.nftables.validation.portRequired': '请输入端口号',
  'app.nftables.validation.portRangeOrder': '端口段起始值需小于或等于结束值',

  'app.nftables.validation.portRange': '端口号必须在 1-65535 之间',
  'app.nftables.validation.protocolRequired': '请选择协议类型',
  'app.nftables.validation.actionRequired': '请选择动作',
  'app.nftables.validation.sourceFormat': '请输入有效的IP地址或CIDR格式',

  // 抽屉标题
  'app.nftables.drawer.addRule': '添加端口规则',
  'app.nftables.drawer.editRule': '编辑端口规则',
  'app.nftables.drawer.addIPRule': '添加IP黑名单规则',

  // 其他错误信息
  'app.nftables.error.initializeFailed': '初始化失败',
  'app.nftables.error.deleteRuleFailed': '删除规则失败',
  'app.nftables.error.refreshConfigFailed': '刷新配置失败',
  'app.nftables.error.saveConfigFailed': '保存配置失败',

  // 基础规则管理
  'app.nftables.baseRules.title': '基础规则',
  'app.nftables.baseRules.tooltip':
    '配置防火墙的基础策略，决定默认的流量处理方式',
  'app.nftables.baseRules.inputPolicy': '入站策略',
  'app.nftables.baseRules.inputPolicyDescription':
    '设置默认的入站流量处理策略，影响所有未明确配置的端口',

  // 策略选项
  'app.nftables.baseRules.accept': '允许',
  'app.nftables.baseRules.acceptDesc': '默认允许所有入站流量（较低安全性）',
  'app.nftables.baseRules.drop': '丢弃',
  'app.nftables.baseRules.dropDesc': '默认丢弃所有入站流量（高安全性）',
  'app.nftables.baseRules.reject': '拒绝',
  'app.nftables.baseRules.rejectDesc': '默认拒绝所有入站流量并返回错误信息',

  // 安全警告
  'app.nftables.baseRules.warningTitle': '安全提示',
  'app.nftables.baseRules.warningDescription':
    '使用"丢弃"策略时，请确保已正确配置必要的端口规则，避免无法访问服务器',

  // 基础规则相关消息
  'app.nftables.message.baseRulesSaved': '基础规则保存成功',
  'app.nftables.message.saveBaseRulesFailed': '基础规则保存失败',
  'app.nftables.message.fetchBaseRulesFailed': '获取基础规则失败',
};
