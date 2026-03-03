export default {
  'app.logrotate.enum.type.local': '本机配置',
  'app.logrotate.enum.type.global': '全局配置',
  'app.logrotate.enum.type.system': '系统配置',

  // 模式
  'app.logrotate.mode.overview': '结构化模式',
  'app.logrotate.mode.form': '表单模式',
  'app.logrotate.mode.raw': '文件模式',

  // 频率
  'app.logrotate.frequency.daily': '每日',
  'app.logrotate.frequency.weekly': '每周',
  'app.logrotate.frequency.monthly': '每月',
  'app.logrotate.frequency.yearly': '每年',

  // 分类
  'app.logrotate.category.title': '分类',
  'app.logrotate.category.all': '全部',
  'app.logrotate.category.tree.empty': '暂无分类，',
  'app.logrotate.category.tree.create': '立即创建',

  // 列表页面
  'app.logrotate.list.action.create': '创建配置',
  'app.logrotate.list.column.name': '名称',
  'app.logrotate.list.column.path': '日志路径',
  'app.logrotate.list.column.frequency': '轮转频率',
  'app.logrotate.list.column.count': '保留数量',
  'app.logrotate.list.column.status': '状态',
  'app.logrotate.list.column.updated_at': '更新时间',
  'app.logrotate.list.status.active': '已激活',
  'app.logrotate.list.status.inactive': '未激活',
  'app.logrotate.list.operation.activate': '激活',
  'app.logrotate.list.operation.deactivate': '停用',
  'app.logrotate.list.operation.history': '历史',
  'app.logrotate.list.delete.title': '确认删除',
  'app.logrotate.list.delete.content':
    '确定要删除配置 "{name}" 吗？此操作不可恢复。',
  'app.logrotate.list.message.fetch_failed': '获取配置列表失败',
  'app.logrotate.list.message.delete_success': '删除配置成功',
  'app.logrotate.list.message.delete_failed': '删除配置失败',
  'app.logrotate.list.message.activate_success': '激活配置成功',
  'app.logrotate.list.message.activate_failed': '激活配置失败',
  'app.logrotate.list.message.deactivate_success': '停用配置成功',
  'app.logrotate.list.message.deactivate_failed': '停用配置失败',

  // 表单
  'app.logrotate.form.create_title': '创建日志轮转配置',
  'app.logrotate.form.edit_title': '编辑日志轮转配置',
  'app.logrotate.form.name': '配置名称',
  'app.logrotate.form.name_placeholder': '请输入配置名称',
  'app.logrotate.form.name_required': '请输入配置名称',
  'app.logrotate.form.name_pattern':
    '配置名称只能包含字母、数字、下划线和连字符',
  'app.logrotate.form.name_help':
    '用于标识该轮转规则，建议使用有业务含义的短名称',
  'app.logrotate.form.category': '分类',
  'app.logrotate.form.category_placeholder':
    '请选择或输入分类，新分类会自动创建',
  'app.logrotate.form.category_required': '请输入分类名称',
  'app.logrotate.form.category_help':
    '用于区分业务规则；创建时可填写新分类，编辑时分类不可修改',
  'app.logrotate.form.category_create_failed': '创建分类失败',
  'app.logrotate.form.path': '日志路径',
  'app.logrotate.form.path_placeholder':
    '请输入日志文件路径，如：/var/log/nginx/*.log',
  'app.logrotate.form.path_required': '请输入日志文件路径',
  'app.logrotate.form.path_help':
    '支持直接输入，建议使用绝对路径；也可使用通配符，如 /var/log/nginx/*.log',
  'app.logrotate.form.frequency': '轮转频率',
  'app.logrotate.form.frequency_placeholder': '请选择轮转频率',
  'app.logrotate.form.frequency_required': '请选择轮转频率',
  'app.logrotate.form.frequency_help':
    '对应 logrotate 的时间指令，决定轮转触发周期',
  'app.logrotate.form.count': '保留数量',
  'app.logrotate.form.count_placeholder': '请输入保留数量，如：7',
  'app.logrotate.form.count_required': '请输入保留数量',
  'app.logrotate.form.count_min': '保留数量必须大于0',
  'app.logrotate.form.count_help': '示例：7 表示最多保留 7 个历史轮转文件',
  'app.logrotate.form.create': '文件权限',
  'app.logrotate.form.create_help':
    '轮转后新建日志文件时使用的权限、用户和用户组',
  'app.logrotate.form.rotate_options': '轮转选项',
  'app.logrotate.form.create_placeholder':
    '请输入文件权限，如：create 0644 root root',
  'app.logrotate.form.section.basic': '基础信息',
  'app.logrotate.form.section.basic_desc': '先定义规则名和日志文件路径',
  'app.logrotate.form.section.strategy': '轮转策略',
  'app.logrotate.form.section.strategy_desc':
    '配置触发频率、保留数量和轮转行为',
  'app.logrotate.form.section.permission': '文件与权限',
  'app.logrotate.form.section.permission_desc':
    '设置轮转后新日志文件的创建策略',
  'app.logrotate.form.section.script': '执行脚本',
  'app.logrotate.form.section.script_desc': '在轮转前后执行自定义 shell 命令',
  'app.logrotate.form.summary.title': '实时结构化摘要',
  'app.logrotate.form.summary.desc':
    '根据当前表单输入实时展示配置结构，便于快速校验',
  'app.logrotate.form.advanced.permission': '高级设置：文件与权限',
  'app.logrotate.form.advanced.script': '高级设置：执行脚本',
  'app.logrotate.form.preview.title': '生成配置预览',
  'app.logrotate.form.preview.desc': '保存后将以此内容写入配置文件',
  'app.logrotate.form.preview.empty': '请先填写日志路径以生成预览',
  'app.logrotate.form.clear': '清空',

  // 权限设置
  'app.logrotate.permission.owner': '所有者权限',
  'app.logrotate.permission.group': '用户组权限',
  'app.logrotate.permission.other': '其他用户权限',
  'app.logrotate.permission.read': '读取',
  'app.logrotate.permission.write': '写入',
  'app.logrotate.permission.execute': '执行',
  'app.logrotate.permission.mode': '权限码',
  'app.logrotate.permission.mode_placeholder': '如：0644',
  'app.logrotate.permission.user': '用户',
  'app.logrotate.permission.user_placeholder': '如：root',
  'app.logrotate.permission.group_name': '用户组',
  'app.logrotate.permission.group_placeholder': '如：root',
  'app.logrotate.permission.preview': '预览',
  'app.logrotate.permission.advanced_show': '显示高级权限',
  'app.logrotate.permission.advanced_hide': '隐藏高级权限',
  'app.logrotate.form.compress': '压缩旧文件',
  'app.logrotate.form.compress_help': '开启后会对轮转后的历史日志做压缩',
  'app.logrotate.form.delay_compress': '延迟压缩',
  'app.logrotate.form.delay_compress_help':
    '配合压缩使用，跳过最新一份轮转文件，下一次再压缩',
  'app.logrotate.form.missing_ok': '忽略缺失文件',
  'app.logrotate.form.missing_ok_help':
    '日志文件不存在时不报错，继续执行其它规则',
  'app.logrotate.form.not_if_empty': '忽略空文件',
  'app.logrotate.form.not_if_empty_help':
    '日志为空时不轮转，避免产生无意义历史文件',
  'app.logrotate.form.pre_rotate': '轮转前命令',
  'app.logrotate.form.pre_rotate_placeholder': '请输入轮转前执行的命令',
  'app.logrotate.form.pre_rotate_help':
    '用于轮转前执行准备动作，支持多行 shell 命令',
  'app.logrotate.form.post_rotate': '轮转后命令',
  'app.logrotate.form.post_rotate_placeholder': '请输入轮转后执行的命令',
  'app.logrotate.form.post_rotate_help':
    '常用于重载服务使新日志文件生效，支持多行 shell 命令',
  'app.logrotate.form.script_tpl.sharedscripts': '插入 sharedscripts',
  'app.logrotate.form.script_tpl.reload_nginx': '插入重载 nginx',
  'app.logrotate.form.script_tpl.reload_rsyslog': '插入重载 rsyslog',
  'app.logrotate.form.script_tpl.select_placeholder': '选择脚本模板',
  'app.logrotate.form.script_tpl.insert': '插入模板',
  'app.logrotate.form.raw_placeholder': '请输入完整的logrotate配置内容',
  'app.logrotate.form.raw_create_not_supported':
    '文件模式暂不支持创建，请使用表单模式',
  'app.logrotate.form.parse_raw_failed': '解析原始配置失败，请检查配置格式',
  'app.logrotate.form.create_success': '创建配置成功',
  'app.logrotate.form.create_failed': '创建配置失败',
  'app.logrotate.form.update_success': '更新配置成功',
  'app.logrotate.form.update_failed': '更新配置失败',
  'app.logrotate.form.load_content_failed': '加载配置内容失败',
  'app.logrotate.form.system_readonly': '系统配置为只读，不支持保存',
  'app.logrotate.form.readonly_tag': '只读',
  'app.logrotate.overview.tip_view':
    '当前为查看模式：用于确认结构化配置是否符合预期。',
  'app.logrotate.overview.tip_edit':
    '当前为编辑模式：可直接修改结构化配置并实时预览。',
  'app.logrotate.overview.basic': '基础信息',
  'app.logrotate.overview.strategy': '轮转策略',
  'app.logrotate.overview.script': '脚本内容',
  'app.logrotate.overview.script_empty': '未配置脚本',
  'app.logrotate.overview.raw_preview': '配置预览',
  'app.logrotate.overview.edit_button': '切换到编辑',
  'app.logrotate.overview.view_button': '返回查看',
  'app.logrotate.overview.edit_title': '结构化编辑',

  // 历史记录
  'app.logrotate.history.title': '配置历史',
  'app.logrotate.history.current': '当前',
  'app.logrotate.history.column.commit': '提交ID',
  'app.logrotate.history.column.message': '提交信息',
  'app.logrotate.history.column.author': '作者',
  'app.logrotate.history.column.date': '提交时间',
  'app.logrotate.history.operation.restore': '恢复',
  'app.logrotate.history.operation.diff': '对比',
  'app.logrotate.history.restore.title': '确认恢复',
  'app.logrotate.history.restore.content': '确定要恢复到提交 {commit} 吗？',
  'app.logrotate.history.restore.button': '恢复到此版本',
  'app.logrotate.history.diff.title': '文件对比',
  'app.logrotate.history.diff.current': '当前版本',
  'app.logrotate.history.diff.version': '历史版本 {commit}',
  'app.logrotate.history.diff.historical': '历史版本',
  'app.logrotate.history.diff.description':
    '绿色表示在历史版本中存在，当前版本中被删除的内容；红色表示历史版本中不存在，当前版本中新增的内容',
  'app.logrotate.history.message.load_failed': '加载历史记录失败',
  'app.logrotate.history.message.restore_success': '恢复配置成功',
  'app.logrotate.history.message.restore_failed': '恢复配置失败',
  'app.logrotate.history.message.diff_failed': '获取文件对比失败',

  // 分类管理
  'app.logrotate.category.manage.title': '分类管理',
  'app.logrotate.category.manage.create': '创建分类',
  'app.logrotate.category.manage.create_title': '创建分类',
  'app.logrotate.category.manage.edit_title': '编辑分类',
  'app.logrotate.category.manage.column.name': '分类名称',
  'app.logrotate.category.manage.column.count': '配置数量',
  'app.logrotate.category.manage.form.name': '分类名称',
  'app.logrotate.category.manage.form.name_placeholder': '请输入分类名称',
  'app.logrotate.category.manage.form.name_required': '请输入分类名称',
  'app.logrotate.category.manage.form.name_pattern':
    '分类名称只能包含字母、数字、下划线和连字符',
  'app.logrotate.category.manage.delete.title': '确认删除',
  'app.logrotate.category.manage.delete.content':
    '确定要删除分类 "{name}" 吗？此操作不可恢复。',
  'app.logrotate.category.manage.message.load_failed': '加载分类列表失败',
  'app.logrotate.category.manage.message.create_success': '创建分类成功',
  'app.logrotate.category.manage.message.create_failed': '创建分类失败',
  'app.logrotate.category.manage.message.update_success': '更新分类成功',
  'app.logrotate.category.manage.message.update_failed': '更新分类失败',
  'app.logrotate.category.manage.message.delete_success': '删除分类成功',
  'app.logrotate.category.manage.message.delete_failed': '删除分类失败',

  'app.logrotate.permission.invalid_mode': '无效的权限码',
  'app.logrotate.permission.no_permission': '无权限',
  'app.logrotate.permission.friendly_format':
    '用户 {user} 可{ownerPerms}，用户组 {group} 可{groupPerms}，其他用户可{otherPerms}',
};
