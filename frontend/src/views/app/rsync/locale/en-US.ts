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
};
