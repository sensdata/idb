export default {
  'app.ssh.pageTitle': 'SSH 管理',
  'app.ssh.tabs.config': 'SSH 配置',
  'app.ssh.tabs.keyPairs': '密钥对管理',
  'app.ssh.tabs.publicKeys': '公钥管理',
  'app.ssh.config.title': 'SSH 配置',
  'app.ssh.keyPairs.title': '密钥对管理',

  // 模式切换
  'app.ssh.mode.visual': '表单模式',
  'app.ssh.mode.source': '源文件模式',
  'app.ssh.mode.switchConfirmTitle': '切换模式确认',
  'app.ssh.mode.switchConfirmContent':
    '切换到表单模式将解析当前源文件配置并更新表单。未保存的更改可能会丢失。是否继续？',

  // 源文件模式
  'app.ssh.source.save': '保存',
  'app.ssh.source.reset': '重置',
  'app.ssh.source.placeholder': '在此编辑SSH配置文件内容',
  'app.ssh.source.info':
    '此文本编辑器允许直接编辑SSH配置文件。对文件进行修改后，点击"保存"按钮使更改生效。',
  'app.ssh.source.saveSuccess': '源文件配置已保存',
  'app.ssh.source.saveError': '保存源文件配置失败',
  'app.ssh.source.resetSuccess': '源文件配置已重置',
  'app.ssh.source.parseSuccess': '解析源文件配置成功',
  'app.ssh.source.parseError': '解析源文件配置失败',
  'app.ssh.source.emptyConfig': '配置内容为空',
  'app.ssh.source.noChanges': '未检测到有效的配置更改',
  'app.ssh.savingConfig': '正在保存SSH配置...',

  // 未保存更改弹窗
  'app.ssh.unsavedChanges.title': '未保存的更改',
  'app.ssh.unsavedChanges.content':
    '源文件编辑器中有未保存的更改。切换到表单模式将丢弃这些更改。是否继续？',
  'app.ssh.unsavedChanges.discard': '丢弃更改',
  'app.ssh.unsavedChanges.cancel': '取消',

  // 配置更新
  'app.ssh.config.updateSuccess': '配置更新成功',
  'app.ssh.config.updateError': '配置更新失败',

  // 错误提示
  'app.ssh.error.fetchFailed': '获取SSH配置失败',
  'app.ssh.error.noHost': '未选择主机',
  'app.ssh.error.emptyConfig': '收到的SSH配置为空',

  // 加载提示
  'app.ssh.loading': '正在加载SSH配置...',
  'app.ssh.savingChanges': '正在保存配置更改...',

  'app.ssh.port.label': '端口',
  'app.ssh.port.description': '指定 SSH 服务监听的端口号，默认为 22。',
  'app.ssh.listen.label': '监听地址',
  'app.ssh.listen.description': '指定 SSH 服务监听的 IP 地址。',
  'app.ssh.root.label': 'root 用户',
  'app.ssh.root.description': 'root 用户 SSH 登录方式，默认允许 SSH 登录。',
  'app.ssh.password.label': '密码认证',
  'app.ssh.password.description': '是否启用密码认证，默认启用。',
  'app.ssh.key.label': '公钥认证',
  'app.ssh.key.description': '是否启用公钥认证，默认启用。',
  'app.ssh.passwordInfo.label': '密码信息',
  'app.ssh.reverse.label': '反向解析',
  'app.ssh.reverse.description':
    '指定 SSH 服务是否对客户端 DNS 解析功能，从而加速连接建立的时间。',
  'app.ssh.sftp.label': 'SFTP 子系统',
  'app.ssh.sftp.description':
    '启用或禁用 SFTP（SSH 文件传输协议）子系统，用于安全文件传输。',
  'app.ssh.autostart.label': '自动启动',
  'app.ssh.btn.setting': '设置',

  // 端口设置弹窗
  'app.ssh.portModal.title': '端口设置',
  'app.ssh.portModal.port': '端口号',
  'app.ssh.portModal.description': '指定 SSH 服务监听的端口号，默认为 22。',
  'app.ssh.portModal.save': '保存',
  'app.ssh.portModal.cancel': '取消',
  'app.ssh.portModal.saveSuccess': '端口设置已保存',
  'app.ssh.portModal.saveError': '端口设置保存失败',

  // 监听地址设置弹窗
  'app.ssh.listenModal.title': '监听地址设置',
  'app.ssh.listenModal.address': '监听地址',
  'app.ssh.listenModal.description':
    '指定 SSH 服务监听的 IP 地址，默认为 0.0.0.0。',
  'app.ssh.listenModal.save': '保存',
  'app.ssh.listenModal.cancel': '取消',
  'app.ssh.listenModal.saveSuccess': '监听地址设置已保存',
  'app.ssh.listenModal.saveError': '监听地址设置保存失败',

  // Root用户设置弹窗
  'app.ssh.rootModal.title': 'Root用户设置',
  'app.ssh.rootModal.label': '允许 root 用户登录',
  'app.ssh.rootModal.description':
    '是否允许 root 用户通过 SSH 登录系统，默认启用。',
  'app.ssh.rootModal.save': '保存',
  'app.ssh.rootModal.cancel': '取消',
  'app.ssh.rootModal.allow': '允许SSH登录',
  'app.ssh.rootModal.deny': '禁止SSH登录',
  'app.ssh.rootModal.saveSuccess': 'Root用户设置已保存',
  'app.ssh.rootModal.saveError': 'Root用户设置保存失败',

  // 密码管理标签页
  'app.ssh.keyPairs.generateKey': '生成密钥',
  'app.ssh.keyPairs.hasPassword': '已设置密码',
  'app.ssh.keyPairs.noPassword': '未设置密码',
  'app.ssh.keyPairs.download': '下载',
  'app.ssh.keyPairs.set': '设置密码',
  'app.ssh.keyPairs.update': '更新密码',
  'app.ssh.keyPairs.clear': '清除密码',
  'app.ssh.keyPairs.delete': '删除',

  'app.ssh.keyPairs.columns.keyName': '密钥名称',
  'app.ssh.keyPairs.columns.encryptionMode': '加密方式',
  'app.ssh.keyPairs.columns.keyBits': '密钥位数',
  'app.ssh.keyPairs.columns.password': '密码',
  'app.ssh.keyPairs.columns.createTime': '创建时间',
  'app.ssh.keyPairs.columns.enabled': '启用状态',
  'app.ssh.keyPairs.columns.user': '用户',
  'app.ssh.keyPairs.columns.keyPath': '密钥路径',
  'app.ssh.keyPairs.columns.fingerprint': '指纹',
  'app.ssh.keyPairs.columns.status': '状态',
  'app.ssh.keyPairs.enabled': '已启用',
  'app.ssh.keyPairs.disabled': '已禁用',
  'app.ssh.keyPairs.enable': '启用',
  'app.ssh.keyPairs.disable': '禁用',

  'app.ssh.keyPairs.generateSuccess': 'SSH密钥生成成功',
  'app.ssh.keyPairs.enableSuccess': 'SSH密钥启用成功',
  'app.ssh.keyPairs.disableSuccess': 'SSH密钥禁用成功',
  'app.ssh.keyPairs.downloadSuccess': 'SSH密钥下载成功',
  'app.ssh.keyPairs.setSuccess': 'SSH密钥密码设置成功',
  'app.ssh.keyPairs.updateSuccess': 'SSH密钥密码更新成功',
  'app.ssh.keyPairs.clearSuccess': 'SSH密钥密码清除成功',
  'app.ssh.keyPairs.deleteSuccess': 'SSH密钥删除成功',
  'app.ssh.keyPairs.operationFailed': '操作失败',

  'app.ssh.keyPairs.clearConfirm': '确定要清除密钥 "{keyName}" 的密码吗？',
  'app.ssh.keyPairs.deleteConfirm': '确定要删除密钥 "{keyName}" 吗？',

  // 生成密钥弹窗
  'app.ssh.keyPairs.generateModal.title': '生成SSH密钥',
  'app.ssh.keyPairs.generateModal.keyName': '密钥名称',
  'app.ssh.keyPairs.generateModal.encryptionMode': '加密方式',
  'app.ssh.keyPairs.generateModal.keyBits': '密钥位数',
  'app.ssh.keyPairs.generateModal.password': '密码',
  'app.ssh.keyPairs.generateModal.enable': '生成后启用',
  'app.ssh.keyPairs.generateModal.keyNameRequired': '请输入密钥名称',
  'app.ssh.keyPairs.generateModal.encryptionModeRequired': '请选择加密方式',
  'app.ssh.keyPairs.generateModal.keyBitsRequired': '请选择密钥位数',

  // 设置密码弹窗
  'app.ssh.keyPairs.setModal.title': '设置密钥密码',
  'app.ssh.keyPairs.setModal.password': '密码',
  'app.ssh.keyPairs.setModal.passwordRequired': '请输入密码',

  // 更新密码弹窗
  'app.ssh.keyPairs.updateModal.title': '更新密钥密码',
  'app.ssh.keyPairs.updateModal.oldPassword': '旧密码',
  'app.ssh.keyPairs.updateModal.newPassword': '新密码',
  'app.ssh.keyPairs.updateModal.oldPasswordRequired': '请输入旧密码',
  'app.ssh.keyPairs.updateModal.newPasswordRequired': '请输入新密码',

  // 授权密钥标签页
  'app.ssh.publicKeys.add': '添加密钥',
  'app.ssh.publicKeys.remove': '删除',
  'app.ssh.publicKeys.addSuccess': 'SSH 密钥添加成功',
  'app.ssh.publicKeys.removeSuccess': 'SSH 密钥删除成功',
  'app.ssh.publicKeys.columns.algorithm': '算法',
  'app.ssh.publicKeys.columns.key': '密钥',
  'app.ssh.publicKeys.columns.comment': '备注',
  'app.ssh.publicKeys.columns.operations': '操作',
  'app.ssh.publicKeys.modal.title': '添加 SSH 密钥',
  'app.ssh.publicKeys.modal.content': '密钥内容',
  'app.ssh.publicKeys.modal.placeholder': '在此粘贴您的 SSH 公钥',
  'app.ssh.publicKeys.modal.description': '格式: ssh-rsa AAAAB3NzaC1... 备注',
  'app.ssh.publicKeys.modal.emptyError': '密钥内容不能为空',
  'app.ssh.publicKeys.modal.formatError': '无效的密钥格式',
  'app.ssh.publicKeys.removeModal.title': '删除 SSH 密钥',
  'app.ssh.publicKeys.removeModal.content': '确定要删除此 SSH 密钥吗？',

  // 状态组件
  'app.ssh.status.running': '运行中',
  'app.ssh.status.stopped': '已停止',
  'app.ssh.status.starting': '启动中...',
  'app.ssh.status.stopping': '停止中...',
  'app.ssh.status.unknown': '未知状态',
  'app.ssh.status.error': '错误',
  'app.ssh.status.unhealthy': '不健康',
  'app.ssh.status.stop': '停止',
  'app.ssh.status.reload': '重载',
  'app.ssh.status.restart': '重启',
  'app.ssh.status.stopSuccess': 'SSH 服务已成功停止',
  'app.ssh.status.stopFailed': 'SSH 服务停止失败',
  'app.ssh.status.reloadSuccess': 'SSH 服务已成功重载',
  'app.ssh.status.reloadFailed': 'SSH 服务重载失败',
  'app.ssh.status.restartSuccess': 'SSH 服务已成功重启',
  'app.ssh.status.restartFailed': 'SSH 服务重启失败',
  'app.ssh.status.autoStartEnabled': '自启动已启用',
  'app.ssh.status.autoStartDisabled': '自启动已禁用',
};
