export default {
  // Actions
  'app.rsync.action.createTask': 'Create Task',

  // Table columns
  'app.rsync.columns.id': 'Task ID',
  'app.rsync.columns.src': 'Source Path',
  'app.rsync.columns.dst': 'Destination Path',
  'app.rsync.columns.mode': 'Mode',
  'app.rsync.columns.status': 'Status',
  'app.rsync.columns.progress': 'Progress',

  // Modes
  'app.rsync.mode.copy': 'Full Copy',
  'app.rsync.mode.incremental': 'Incremental Sync',

  // Status
  'app.rsync.status.running': 'Running',
  'app.rsync.status.success': 'Success',
  'app.rsync.status.failed': 'Failed',
  'app.rsync.status.unknown': 'Unknown',

  // Create drawer
  'app.rsync.create.title': 'Create Sync Task',

  // Form
  'app.rsync.form.srcHost': 'Source Host',
  'app.rsync.form.srcPath': 'Source Path',
  'app.rsync.form.dstHost': 'Destination Host',
  'app.rsync.form.dstPath': 'Destination Path',
  'app.rsync.form.mode': 'Sync Mode',
  'app.rsync.form.placeholder.srcHost': 'Select source host',
  'app.rsync.form.placeholder.dstHost': 'Select destination host',
  'app.rsync.form.placeholder.srcPath': '/path/to/source',
  'app.rsync.form.placeholder.dstPath': '/path/to/destination',

  // Detail drawer
  'app.rsync.detail.title': 'Task Detail - {id}',
  'app.rsync.field.taskId': 'Task ID',
  'app.rsync.field.src': 'Source Path',
  'app.rsync.field.dst': 'Destination Path',
  'app.rsync.field.cacheDir': 'Cache Directory',
  'app.rsync.field.mode': 'Mode',
  'app.rsync.field.status': 'Status',
  'app.rsync.field.progress': 'Progress',
  'app.rsync.field.step': 'Current Step',
  'app.rsync.field.startTime': 'Start Time',
  'app.rsync.field.endTime': 'End Time',
  'app.rsync.field.error': 'Error',
  'app.rsync.field.lastLog': 'Latest Log',

  // Messages
  'app.rsync.message.fetchListFailed': 'Failed to fetch task list',
  'app.rsync.message.fetchHostFailed': 'Failed to fetch hosts',
  'app.rsync.message.formIncomplete': 'Please complete all fields',
  'app.rsync.message.createSuccess': 'Task created successfully',
  'app.rsync.message.createFailed': 'Failed to create task',
  'app.rsync.message.fetchDetailFailed': 'Failed to fetch task detail',
  'app.rsync.message.cancelSuccess': 'Task cancelled',
  'app.rsync.message.cancelFailed': 'Failed to cancel task',
  'app.rsync.message.retrySuccess': 'Task retried',
  'app.rsync.message.retryFailed': 'Failed to retry task',
  'app.rsync.message.deleteSuccess': 'Task deleted',
  'app.rsync.message.deleteFailed': 'Failed to delete task',

  // ==================== Remote Sync ====================
  'app.rsync.client.title': 'Remote Sync',
  'app.rsync.client.action.create': 'Create Config',
  'app.rsync.client.action.edit': 'Edit Config',
  'app.rsync.client.action.run': 'Run',
  'app.rsync.client.action.cancel': 'Cancel',

  // Table columns
  'app.rsync.client.columns.name': 'Config Name',
  'app.rsync.client.columns.direction': 'Sync Type',
  'app.rsync.client.columns.localPath': 'Local Path',
  'app.rsync.client.columns.remoteType': 'Remote Type',
  'app.rsync.client.columns.remoteHost': 'Remote Host',
  'app.rsync.client.columns.state': 'State',
  'app.rsync.client.columns.attempt': 'Attempts',
  'app.rsync.client.columns.createdAt': 'Created At',

  // Sync direction
  'app.rsync.client.direction.localToRemote': 'Local -> Remote',
  'app.rsync.client.direction.remoteToLocal': 'Remote -> Local',

  // Remote type
  'app.rsync.client.remoteType.rsync': 'Rsync Service',
  'app.rsync.client.remoteType.ssh': 'SSH Service',

  // Auth mode
  'app.rsync.client.authMode.password': 'Password',
  'app.rsync.client.authMode.anonymous': 'Anonymous',
  'app.rsync.client.authMode.privateKey': 'Private Key',

  // State
  'app.rsync.client.state.pending': 'Pending',
  'app.rsync.client.state.running': 'Running',
  'app.rsync.client.state.success': 'Success',
  'app.rsync.client.state.succeeded': 'Succeeded',
  'app.rsync.client.state.failed': 'Failed',
  'app.rsync.client.state.canceled': 'Canceled',

  // Form
  'app.rsync.client.form.name': 'Config Name',
  'app.rsync.client.form.direction': 'Sync Type',
  'app.rsync.client.form.localPath': 'Local File/Directory Path',
  'app.rsync.client.form.remoteType': 'Remote Type',
  'app.rsync.client.form.remoteHost': 'Host',
  'app.rsync.client.form.remotePort': 'Port',
  'app.rsync.client.form.authMode': 'Auth Mode',
  'app.rsync.client.form.username': 'Rsync Account',
  'app.rsync.client.form.password': 'Rsync Password',
  'app.rsync.client.form.sshPrivateKey': 'SSH Private Key',
  'app.rsync.client.form.remotePath': 'Remote Path',
  'app.rsync.client.form.remotePathInModule': 'Path in Module',
  'app.rsync.client.form.module': 'Module Name',
  'app.rsync.client.form.enqueue': 'Execute Immediately',

  // Form placeholders
  'app.rsync.client.form.placeholder.name': 'Enter config name',
  'app.rsync.client.form.placeholder.localPath': '/etc/nginx/ssl/',
  'app.rsync.client.form.placeholder.remoteHost':
    'Enter remote host IP or domain',
  'app.rsync.client.form.placeholder.remotePort': 'Enter port',
  'app.rsync.client.form.placeholder.username': 'Enter account',
  'app.rsync.client.form.placeholder.password': 'Enter password',
  'app.rsync.client.form.placeholder.sshPrivateKey':
    'Enter SSH private key content',
  'app.rsync.client.form.placeholder.remotePath':
    'Enter remote path, e.g. /data/backup',
  'app.rsync.client.form.placeholder.remotePathInModule':
    'Enter path in module, e.g. /subdir (optional)',
  'app.rsync.client.form.placeholder.module': 'Enter module name',

  // Messages
  'app.rsync.client.message.fetchListFailed': 'Failed to fetch config list',
  'app.rsync.client.message.createSuccess': 'Config created successfully',
  'app.rsync.client.message.createFailed': 'Failed to create config',
  'app.rsync.client.message.updateSuccess': 'Config updated successfully',
  'app.rsync.client.message.updateFailed': 'Failed to update config',
  'app.rsync.client.message.deleteSuccess': 'Config deleted',
  'app.rsync.client.message.deleteFailed': 'Failed to delete config',
  'app.rsync.client.message.testSuccess': 'Test successful',
  'app.rsync.client.message.testFailed': 'Test failed',
  'app.rsync.client.message.retrySuccess': 'Retried',
  'app.rsync.client.message.retryFailed': 'Retry failed',
  'app.rsync.client.message.runSuccess': 'Task started',
  'app.rsync.client.message.runFailed': 'Failed to start task',
  'app.rsync.client.message.cancelSuccess': 'Cancelled',
  'app.rsync.client.message.cancelFailed': 'Cancel failed',
  'app.rsync.client.message.confirmDelete':
    'Are you sure to delete this config?',
  'app.rsync.client.message.confirmCancel': 'Are you sure to cancel this task?',

  // Logs
  'app.rsync.client.logs.title': 'Task Logs',
  'app.rsync.client.logs.button': 'Logs',
  'app.rsync.client.logs.fetchFailed': 'Failed to fetch logs',
  'app.rsync.client.logs.loadFailed': 'Failed to load log content',
  'app.rsync.client.logs.loading': 'Loading...',
  'app.rsync.client.logs.empty': 'No logs',
  'app.rsync.client.logs.fileList': 'Log Files',
  'app.rsync.client.logs.selectFile': 'Please select a log file',
};
