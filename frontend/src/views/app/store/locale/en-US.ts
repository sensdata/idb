export default {
  'app.store.app.tabs.all': 'All',
  'app.store.app.tabs.installed': 'Installed',
  'app.store.app.list.install': 'Install',
  'app.store.app.list.installed': 'Installed',
  'app.store.app.list.version': 'Version',
  'app.store.app.list.upgrade': 'Upgrade',
  'app.store.app.list.uninstall': 'Uninstall',
  'app.store.app.list.hasUpgrade': 'Update Available',
  'app.store.app.list.install_at': 'Installed at',
  'app.store.app.syncAppList': 'Sync App List',
  'app.store.app.installed.manage': 'Manage',

  'app.store.app.message.syncSuccess': 'App list updated successfully',

  'app.store.app.install.title': 'Install App',
  'app.store.app.upgrade.title': 'Upgrade App',
  'app.store.app.install.mode.form': 'Edit by form',
  'app.store.app.install.mode.yaml': 'Edit by compose.yaml',
  'app.store.app.install.name': 'App Name',
  'app.store.app.install.version': 'Version',
  'app.store.app.install.version.required': 'Please select a version',
  'app.store.app.install.version.placeholder': 'Please select a version',
  'app.store.app.install.required': 'Please enter {label}',
  'app.store.app.install.minLength': 'Please enter at least {min} characters',
  'app.store.app.install.maxLength':
    'Please enter no more than {max} characters',
  'app.store.app.install.pattern': 'Invalid format',
  'app.store.app.upgrade.confirm':
    'The docker-compose.yml of version {version} will be used to overwrite, please backup the data before upgrading',
  'app.store.app.install.success': 'App installed successfully',
  'app.store.app.install.success.confirm.title': 'Installation Successful',
  'app.store.app.install.success.confirm.content':
    'The app has been successfully installed! Would you like to go to the container management page to view it?',
  'app.store.app.install.success.confirm.ok': 'Go to Container Management',
  'app.store.app.install.success.confirm.cancel': 'Stay on Current Page',
  'app.store.app.upgrade.success': 'App upgraded successfully',
  'app.store.app.uninstall.confirm':
    'Are you sure you want to uninstall the app?',

  'app.store.upgradeLog.title': 'Upgrade App',
  'app.store.upgradeLog.waitingForLogs': 'Waiting for logs...',
  'app.store.upgradeLog.close': 'Close',
  'app.store.upgradeLog.success': 'Upgrade success',
  'app.store.upgradeLog.progress': 'Upgrading...',
  'app.store.upgradeLog.failed': 'Upgrade failed',
  'app.store.upgradeLog.timeout': 'Upgrade timeout',
  'app.store.upgradeLog.logConnected': 'Connection success',
  'app.store.upgradeLog.logConnectionFailed': 'Connection failed',

  'app.store.uninstallLog.title': 'Uninstall App',
  'app.store.uninstallLog.waitingForLogs': 'Waiting for logs...',
  'app.store.uninstallLog.close': 'Close',
  'app.store.uninstallLog.success': 'Uninstall success',
  'app.store.uninstallLog.progress': 'Uninstalling...',
  'app.store.uninstallLog.failed': 'Uninstall failed',
  'app.store.uninstallLog.timeout': 'Uninstall timeout',
  'app.store.uninstallLog.logConnected': 'Connection success',
  'app.store.uninstallLog.logConnectionFailed': 'Connection failed',

  // Database Manager
  'app.store.database.manager.title': '{type} Management - {name}',
  'app.store.database.tab.info': 'Basic Info',
  'app.store.database.tab.config': 'Configuration',
  'app.store.database.tab.password': 'Password',
  'app.store.database.tab.remote': 'Remote Access',
  'app.store.database.tab.port': 'Port Settings',
  'app.store.database.tab.connection': 'Connection',

  'app.store.database.connection.internal': 'Container connection',
  'app.store.database.connection.external': 'Public connection',

  'app.store.database.info.name': 'Name',
  'app.store.database.info.version': 'Version',
  'app.store.database.info.host': 'Address',
  'app.store.database.info.port': 'Port',
  'app.store.database.info.status': 'Status',

  'app.store.database.button.start': 'Start',
  'app.store.database.button.stop': 'Stop',
  'app.store.database.button.restart': 'Restart',
  'app.store.database.button.save': 'Save Config',
  'app.store.database.button.refresh': 'Refresh',
  'app.store.database.button.refreshStatus': 'Refresh Status',
  'app.store.database.button.changePassword': 'Change Password',
  'app.store.database.button.changePort': 'Change Port',

  'app.store.database.config.placeholder': 'Configuration file content',

  'app.store.database.password.current': 'Current Password',
  'app.store.database.password.new': 'New Password',
  'app.store.database.password.newPlaceholder': 'Please enter new password',

  'app.store.database.remote.status': 'Remote Access Status',
  'app.store.database.remote.enabled': 'Enabled',
  'app.store.database.remote.disabled': 'Disabled',
  'app.store.database.remote.warning':
    'After enabling remote access, please ensure a strong password is set for security',

  'app.store.database.port.label': 'Port Number',
  'app.store.database.port.placeholder': 'Please enter port number',

  'app.store.database.message.loadConfigFailed': 'Failed to load configuration',
  'app.store.database.message.loadPasswordFailed': 'Failed to load password',
  'app.store.database.message.loadRemoteAccessFailed':
    'Failed to load remote access status',
  'app.store.database.message.loadDataFailed': 'Failed to load data',
  'app.store.database.message.operationSuccess':
    '{operation} operation successful',
  'app.store.database.message.operationFailed': '{operation} operation failed',
  'app.store.database.message.configSaveSuccess':
    'Configuration saved successfully',
  'app.store.database.message.configSaveFailed': 'Failed to save configuration',
  'app.store.database.message.passwordRequired': 'Please enter new password',
  'app.store.database.message.passwordChangeSuccess':
    'Password changed successfully',
  'app.store.database.message.passwordChangeFailed':
    'Failed to change password',
  'app.store.database.message.remoteAccessSuccess': 'Remote access {status}',
  'app.store.database.message.remoteAccessFailed':
    'Failed to set remote access',
  'app.store.database.message.portChangeSuccess': 'Port changed successfully',
  'app.store.database.message.portChangeFailed': 'Failed to change port',
  'app.store.database.message.refreshSuccess': 'Status refreshed successfully',
  'app.store.database.message.refreshFailed': 'Failed to refresh status',

  'app.store.database.remoteStatus.enabled': 'enabled',
  'app.store.database.remoteStatus.disabled': 'disabled',
};
