export default {
  // 操作
  'app.rsync.action.createTask': '创建任务',

  // 表格列
  'app.rsync.columns.id': '任务ID',
  'app.rsync.columns.src': '源路径',
  'app.rsync.columns.dst': '目标路径',
  'app.rsync.columns.mode': '模式',
  'app.rsync.columns.status': '状态',
  'app.rsync.columns.progress': '进度',

  // 模式
  'app.rsync.mode.copy': '完整复制',
  'app.rsync.mode.incremental': '增量同步',

  // 状态
  'app.rsync.status.running': '运行中',
  'app.rsync.status.success': '成功',
  'app.rsync.status.failed': '失败',
  'app.rsync.status.unknown': '未知',

  // 创建抽屉
  'app.rsync.create.title': '创建同步任务',

  // 表单
  'app.rsync.form.srcHost': '源主机',
  'app.rsync.form.srcPath': '源路径',
  'app.rsync.form.dstHost': '目标主机',
  'app.rsync.form.dstPath': '目标路径',
  'app.rsync.form.mode': '同步模式',
  'app.rsync.form.placeholder.srcHost': '选择源主机',
  'app.rsync.form.placeholder.dstHost': '选择目标主机',
  'app.rsync.form.placeholder.srcPath': '/path/to/source',
  'app.rsync.form.placeholder.dstPath': '/path/to/destination',

  // 详情抽屉
  'app.rsync.detail.title': '任务详情 - {id}',
  'app.rsync.field.taskId': '任务ID',
  'app.rsync.field.src': '源路径',
  'app.rsync.field.dst': '目标路径',
  'app.rsync.field.cacheDir': '缓存目录',
  'app.rsync.field.mode': '模式',
  'app.rsync.field.status': '状态',
  'app.rsync.field.progress': '进度',
  'app.rsync.field.step': '当前步骤',
  'app.rsync.field.startTime': '开始时间',
  'app.rsync.field.endTime': '结束时间',
  'app.rsync.field.error': '错误信息',
  'app.rsync.field.lastLog': '最新日志',

  // 提示信息
  'app.rsync.message.fetchListFailed': '获取任务列表失败',
  'app.rsync.message.fetchHostFailed': '获取主机列表失败',
  'app.rsync.message.formIncomplete': '请填写完整信息',
  'app.rsync.message.createSuccess': '任务创建成功',
  'app.rsync.message.createFailed': '任务创建失败',
  'app.rsync.message.fetchDetailFailed': '获取任务详情失败',
  'app.rsync.message.cancelSuccess': '任务已取消',
  'app.rsync.message.cancelFailed': '取消任务失败',
  'app.rsync.message.retrySuccess': '任务已重试',
  'app.rsync.message.retryFailed': '重试任务失败',
  'app.rsync.message.deleteSuccess': '任务已删除',
  'app.rsync.message.deleteFailed': '删除任务失败',

  // ==================== 远程同步 ====================
  'app.rsync.client.title': '远程同步',
  'app.rsync.client.action.create': '创建配置',
  'app.rsync.client.action.edit': '编辑配置',
  'app.rsync.client.action.run': '执行',
  'app.rsync.client.action.cancel': '取消',

  // 表格列
  'app.rsync.client.columns.name': '配置名称',
  'app.rsync.client.columns.direction': '同步类型',
  'app.rsync.client.columns.localPath': '本地路径',
  'app.rsync.client.columns.remoteType': '远程类型',
  'app.rsync.client.columns.remoteHost': '远程主机',
  'app.rsync.client.columns.state': '状态',
  'app.rsync.client.columns.attempt': '尝试次数',
  'app.rsync.client.columns.createdAt': '创建时间',

  // 同步方向
  'app.rsync.client.direction.localToRemote': '本地->远程',
  'app.rsync.client.direction.remoteToLocal': '远程->本地',

  // 远程类型
  'app.rsync.client.remoteType.rsync': 'Rsync 服务',
  'app.rsync.client.remoteType.ssh': 'SSH 服务',

  // 认证方式
  'app.rsync.client.authMode.password': '密码',
  'app.rsync.client.authMode.anonymous': '匿名',
  'app.rsync.client.authMode.privateKey': '私钥',

  // 状态
  'app.rsync.client.state.pending': '待执行',
  'app.rsync.client.state.running': '运行中',
  'app.rsync.client.state.success': '成功',
  'app.rsync.client.state.succeeded': '成功',
  'app.rsync.client.state.failed': '失败',
  'app.rsync.client.state.canceled': '已取消',

  // 表单
  'app.rsync.client.form.name': '配置名称',
  'app.rsync.client.form.direction': '同步类型',
  'app.rsync.client.form.localPath': '本地文件/目录路径',
  'app.rsync.client.form.remoteType': '远程类型',
  'app.rsync.client.form.remoteHost': '主机',
  'app.rsync.client.form.remotePort': '端口',
  'app.rsync.client.form.authMode': '验证方式',
  'app.rsync.client.form.username': 'Rsync 帐号',
  'app.rsync.client.form.password': 'Rsync 密码',
  'app.rsync.client.form.sshPrivateKey': 'SSH 私钥',
  'app.rsync.client.form.remotePath': '远程路径',
  'app.rsync.client.form.remotePathInModule': '模块内路径',
  'app.rsync.client.form.module': '模块名',
  'app.rsync.client.form.enqueue': '立即执行',

  // 表单占位符
  'app.rsync.client.form.placeholder.name': '请输入配置名称',
  'app.rsync.client.form.placeholder.localPath': '/etc/nginx/ssl/',
  'app.rsync.client.form.placeholder.remoteHost': '请输入远程主机 IP 或域名',
  'app.rsync.client.form.placeholder.remotePort': '请输入端口',
  'app.rsync.client.form.placeholder.username': '请输入帐号',
  'app.rsync.client.form.placeholder.password': '请输入密码',
  'app.rsync.client.form.placeholder.sshPrivateKey': '请输入 SSH 私钥内容',
  'app.rsync.client.form.placeholder.remotePath':
    '请输入远程路径，如 /data/backup',
  'app.rsync.client.form.placeholder.remotePathInModule':
    '请输入模块内路径，如 /subdir（可选）',
  'app.rsync.client.form.placeholder.module': '请输入模块名',

  // 提示信息
  'app.rsync.client.message.fetchListFailed': '获取配置列表失败',
  'app.rsync.client.message.createSuccess': '配置创建成功',
  'app.rsync.client.message.createFailed': '配置创建失败',
  'app.rsync.client.message.updateSuccess': '配置更新成功',
  'app.rsync.client.message.updateFailed': '配置更新失败',
  'app.rsync.client.message.deleteSuccess': '配置已删除',
  'app.rsync.client.message.deleteFailed': '删除配置失败',
  'app.rsync.client.message.testSuccess': '测试成功',
  'app.rsync.client.message.testFailed': '测试失败',
  'app.rsync.client.message.retrySuccess': '已重试',
  'app.rsync.client.message.retryFailed': '重试失败',
  'app.rsync.client.message.runSuccess': '任务已启动',
  'app.rsync.client.message.runFailed': '启动任务失败',
  'app.rsync.client.message.cancelSuccess': '已取消',
  'app.rsync.client.message.cancelFailed': '取消失败',
  'app.rsync.client.message.confirmDelete': '确定要删除此配置吗？',
  'app.rsync.client.message.confirmCancel': '确定要取消此任务吗？',

  // 日志
  'app.rsync.client.logs.title': '任务日志',
  'app.rsync.client.logs.button': '日志',
  'app.rsync.client.logs.fetchFailed': '获取日志列表失败',
  'app.rsync.client.logs.loadFailed': '加载日志内容失败',
  'app.rsync.client.logs.loading': '加载中...',
  'app.rsync.client.logs.empty': '暂无日志',
  'app.rsync.client.logs.fileList': '日志文件',
  'app.rsync.client.logs.selectFile': '请选择日志文件',
};
