export default {
  'app.logrotate.enum.type.local': 'Host Config',
  'app.logrotate.enum.type.global': 'Global Config',
  'app.logrotate.enum.type.system': 'System Config',

  // Mode
  'app.logrotate.mode.overview': 'Structured Mode',
  'app.logrotate.mode.form': 'Form Mode',
  'app.logrotate.mode.raw': 'Raw Mode',

  // Frequency
  'app.logrotate.frequency.daily': 'Daily',
  'app.logrotate.frequency.weekly': 'Weekly',
  'app.logrotate.frequency.monthly': 'Monthly',
  'app.logrotate.frequency.yearly': 'Yearly',

  // Category
  'app.logrotate.category.title': 'Categories',
  'app.logrotate.category.all': 'All',
  'app.logrotate.category.tree.empty': 'No categories, ',
  'app.logrotate.category.tree.create': 'create now',

  // List page
  'app.logrotate.list.action.create': 'Create Config',
  'app.logrotate.list.column.name': 'Name',
  'app.logrotate.list.column.path': 'Log Path',
  'app.logrotate.list.column.frequency': 'Frequency',
  'app.logrotate.list.column.count': 'Rotate Count',
  'app.logrotate.list.column.status': 'Status',
  'app.logrotate.list.column.updated_at': 'Updated At',
  'app.logrotate.list.status.active': 'Active',
  'app.logrotate.list.status.inactive': 'Inactive',
  'app.logrotate.list.operation.activate': 'Activate',
  'app.logrotate.list.operation.deactivate': 'Deactivate',
  'app.logrotate.list.operation.history': 'History',
  'app.logrotate.list.delete.title': 'Confirm Delete',
  'app.logrotate.list.delete.content':
    'Are you sure you want to delete config "{name}"? This action cannot be undone.',
  'app.logrotate.list.message.fetch_failed': 'Failed to fetch config list',
  'app.logrotate.list.message.delete_success': 'Config deleted successfully',
  'app.logrotate.list.message.delete_failed': 'Failed to delete config',
  'app.logrotate.list.message.activate_success':
    'Config activated successfully',
  'app.logrotate.list.message.activate_failed': 'Failed to activate config',
  'app.logrotate.list.message.deactivate_success':
    'Config deactivated successfully',
  'app.logrotate.list.message.deactivate_failed': 'Failed to deactivate config',

  // Form
  'app.logrotate.form.create_title': 'Create Logrotate Config',
  'app.logrotate.form.edit_title': 'Edit Logrotate Config',
  'app.logrotate.form.name': 'Config Name',
  'app.logrotate.form.name_placeholder': 'Please enter config name',
  'app.logrotate.form.name_required': 'Please enter config name',
  'app.logrotate.form.name_pattern':
    'Config name can only contain letters, numbers, underscores and hyphens',
  'app.logrotate.form.name_help':
    'Used to identify this rotation rule. A short business-oriented name is recommended',
  'app.logrotate.form.category': 'Category',
  'app.logrotate.form.category_placeholder':
    'Please select or enter a category, new categories will be created automatically',
  'app.logrotate.form.category_required': 'Please enter category name',
  'app.logrotate.form.category_help':
    'Used to group rules. You can create a new category when creating; category is immutable in edit mode',
  'app.logrotate.form.category_create_failed': 'Failed to create category',
  'app.logrotate.form.path': 'Log Path',
  'app.logrotate.form.path_placeholder':
    'Please enter log file path, e.g.: /var/log/nginx/*.log',
  'app.logrotate.form.path_required': 'Please enter log file path',
  'app.logrotate.form.path_help':
    'Use absolute paths when possible. Wildcards are supported, e.g. /var/log/nginx/*.log',
  'app.logrotate.form.frequency': 'Frequency',
  'app.logrotate.form.frequency_placeholder': 'Please select frequency',
  'app.logrotate.form.frequency_required': 'Please select frequency',
  'app.logrotate.form.frequency_help':
    'Maps to logrotate time directives and controls how often rotation runs',
  'app.logrotate.form.count': 'Rotate Count',
  'app.logrotate.form.count_placeholder': 'Please enter rotate count, e.g.: 7',
  'app.logrotate.form.count_required': 'Please enter rotate count',
  'app.logrotate.form.count_min': 'Rotate count must be greater than 0',
  'app.logrotate.form.count_help':
    'Example: 7 means keeping up to 7 rotated history files',
  'app.logrotate.form.create': 'File Permission',
  'app.logrotate.form.create_help':
    'Permission, user, and group used when creating a new log file after rotation',
  'app.logrotate.form.rotate_options': 'Rotation Options',
  'app.logrotate.form.create_placeholder':
    'Please enter file permission, e.g.: create 0644 root root',
  'app.logrotate.form.section.basic': 'Basic Parameters',
  'app.logrotate.form.section.basic_desc':
    'Define the rule name and target log file path first',
  'app.logrotate.form.section.strategy': 'Rotation Strategy',
  'app.logrotate.form.section.strategy_desc':
    'Configure trigger frequency, retention count, and behavior flags',
  'app.logrotate.form.section.permission': 'File and Permission',
  'app.logrotate.form.section.permission_desc':
    'Control how new log files are created after rotation',
  'app.logrotate.form.section.script': 'Execution Scripts',
  'app.logrotate.form.section.script_desc':
    'Run custom shell commands before and after rotation',
  'app.logrotate.form.summary.title': 'Live Structured Summary',
  'app.logrotate.form.summary.desc':
    'Shows a live structured view from current form values for quick validation',
  'app.logrotate.form.advanced.permission': 'Advanced: File and Permission',
  'app.logrotate.form.advanced.script': 'Advanced: Execution Scripts',
  'app.logrotate.form.preview.title': 'Generated Config Preview',
  'app.logrotate.form.preview.desc':
    'This content will be written to the config file after saving',
  'app.logrotate.form.preview.empty':
    'Please fill in log path to generate preview',
  'app.logrotate.form.clear': 'Clear',

  // Permission settings
  'app.logrotate.permission.owner': 'Owner Permission',
  'app.logrotate.permission.group': 'Group Permission',
  'app.logrotate.permission.other': 'Other Permission',
  'app.logrotate.permission.read': 'Read',
  'app.logrotate.permission.write': 'Write',
  'app.logrotate.permission.execute': 'Execute',
  'app.logrotate.permission.mode': 'Permission Mode',
  'app.logrotate.permission.mode_placeholder': 'e.g.: 0644',
  'app.logrotate.permission.user': 'User',
  'app.logrotate.permission.user_placeholder': 'e.g.: root',
  'app.logrotate.permission.group_name': 'Group',
  'app.logrotate.permission.group_placeholder': 'e.g.: root',
  'app.logrotate.permission.preview': 'Preview',
  'app.logrotate.permission.advanced_show': 'Show advanced permissions',
  'app.logrotate.permission.advanced_hide': 'Hide advanced permissions',
  'app.logrotate.form.compress': 'Compress Old Files',
  'app.logrotate.form.compress_help':
    'Compress rotated history log files when enabled',
  'app.logrotate.form.delay_compress': 'Delay Compress',
  'app.logrotate.form.delay_compress_help':
    'Use with compression to skip the newest rotated file and compress it in the next cycle',
  'app.logrotate.form.missing_ok': 'Ignore Missing Files',
  'app.logrotate.form.missing_ok_help':
    'Do not fail when the target log file does not exist',
  'app.logrotate.form.not_if_empty': 'Ignore Empty Files',
  'app.logrotate.form.not_if_empty_help':
    'Skip rotation for empty logs to avoid meaningless history files',
  'app.logrotate.form.pre_rotate': 'Pre-rotate Command',
  'app.logrotate.form.pre_rotate_placeholder':
    'Please enter command to execute before rotation',
  'app.logrotate.form.pre_rotate_help':
    'Run preparation steps before rotation. Multi-line shell commands are supported',
  'app.logrotate.form.post_rotate': 'Post-rotate Command',
  'app.logrotate.form.post_rotate_placeholder':
    'Please enter command to execute after rotation',
  'app.logrotate.form.post_rotate_help':
    'Commonly used to reload services so new log files take effect. Multi-line shell commands are supported',
  'app.logrotate.form.script_tpl.sharedscripts': 'Insert sharedscripts',
  'app.logrotate.form.script_tpl.reload_nginx': 'Insert reload nginx',
  'app.logrotate.form.script_tpl.reload_rsyslog': 'Insert reload rsyslog',
  'app.logrotate.form.script_tpl.select_placeholder': 'Select script template',
  'app.logrotate.form.script_tpl.insert': 'Insert template',
  'app.logrotate.form.raw_placeholder':
    'Please enter complete logrotate configuration content',
  'app.logrotate.form.raw_create_not_supported':
    'Raw mode creation is not supported, please use form mode',
  'app.logrotate.form.parse_raw_failed':
    'Failed to parse raw configuration, please check the format',
  'app.logrotate.form.create_success': 'Config created successfully',
  'app.logrotate.form.create_failed': 'Failed to create config',
  'app.logrotate.form.update_success': 'Config updated successfully',
  'app.logrotate.form.update_failed': 'Failed to update config',
  'app.logrotate.form.load_content_failed': 'Failed to load config content',
  'app.logrotate.form.system_readonly':
    'System configuration is read-only and cannot be saved',
  'app.logrotate.form.open_config_dir': 'Open Config Directory',
  'app.logrotate.form.readonly_tag': 'Read-only',
  'app.logrotate.overview.tip_view':
    'View mode: review the structured configuration first.',
  'app.logrotate.overview.tip_edit':
    'Edit mode: modify structured fields with live preview.',
  'app.logrotate.overview.basic': 'Basic Information',
  'app.logrotate.overview.strategy': 'Rotation Strategy',
  'app.logrotate.overview.script': 'Scripts',
  'app.logrotate.overview.script_empty': 'No script configured',
  'app.logrotate.overview.raw_preview': 'Configuration Preview',
  'app.logrotate.overview.edit_button': 'Switch to Edit',
  'app.logrotate.overview.view_button': 'Back to View',
  'app.logrotate.overview.edit_title': 'Structured Edit',

  // History
  'app.logrotate.history.title': 'Config History',
  'app.logrotate.history.current': 'Current',
  'app.logrotate.history.column.commit': 'Commit ID',
  'app.logrotate.history.column.message': 'Commit Message',
  'app.logrotate.history.column.author': 'Author',
  'app.logrotate.history.column.date': 'Commit Date',
  'app.logrotate.history.operation.restore': 'Restore',
  'app.logrotate.history.operation.diff': 'Diff',
  'app.logrotate.history.restore.title': 'Confirm Restore',
  'app.logrotate.history.restore.content':
    'Are you sure you want to restore to commit {commit}?',
  'app.logrotate.history.restore.button': 'Restore to this version',
  'app.logrotate.history.diff.title': 'File Diff',
  'app.logrotate.history.diff.current': 'Current Version',
  'app.logrotate.history.diff.version': 'Historical Version {commit}',
  'app.logrotate.history.diff.historical': 'Historical Version',
  'app.logrotate.history.diff.description':
    'Green shows content that existed in the historical version but was removed in the current version; Red shows content that did not exist in the historical version but was added in the current version',
  'app.logrotate.history.message.load_failed': 'Failed to load history',
  'app.logrotate.history.message.restore_success':
    'Config restored successfully',
  'app.logrotate.history.message.restore_failed': 'Failed to restore config',
  'app.logrotate.history.message.diff_failed': 'Failed to get file diff',

  // Category Management
  'app.logrotate.category.manage.title': 'Category Management',
  'app.logrotate.category.manage.create': 'Create Category',
  'app.logrotate.category.manage.create_title': 'Create Category',
  'app.logrotate.category.manage.edit_title': 'Edit Category',
  'app.logrotate.category.manage.column.name': 'Category Name',
  'app.logrotate.category.manage.column.count': 'Config Count',
  'app.logrotate.category.manage.form.name': 'Category Name',
  'app.logrotate.category.manage.form.name_placeholder':
    'Please enter category name',
  'app.logrotate.category.manage.form.name_required':
    'Please enter category name',
  'app.logrotate.category.manage.form.name_pattern':
    'Category name can only contain letters, numbers, underscores and hyphens',
  'app.logrotate.category.manage.delete.title': 'Confirm Delete',
  'app.logrotate.category.manage.delete.content':
    'Are you sure you want to delete category "{name}"? This action cannot be undone.',
  'app.logrotate.category.manage.message.load_failed':
    'Failed to load category list',
  'app.logrotate.category.manage.message.create_success':
    'Category created successfully',
  'app.logrotate.category.manage.message.create_failed':
    'Failed to create category',
  'app.logrotate.category.manage.message.update_success':
    'Category updated successfully',
  'app.logrotate.category.manage.message.update_failed':
    'Failed to update category',
  'app.logrotate.category.manage.message.delete_success':
    'Category deleted successfully',
  'app.logrotate.category.manage.message.delete_failed':
    'Failed to delete category',

  'app.logrotate.permission.invalid_mode': 'Invalid permission mode',
  'app.logrotate.permission.no_permission': 'No permission',
  'app.logrotate.permission.friendly_format':
    'User {user} can {ownerPerms}, group {group} can {groupPerms}, others can {otherPerms}',
};
