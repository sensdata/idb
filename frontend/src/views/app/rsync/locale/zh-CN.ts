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
};
