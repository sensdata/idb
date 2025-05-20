export default {
  'app.ssh.tabs.config': 'SSH Config',
  'app.ssh.tabs.password': 'SSH Password',
  'app.ssh.tabs.authkey': 'Auth Keys',
  'app.ssh.port.label': 'Port',
  'app.ssh.port.description':
    'Specify the port number for SSH service, default is 22.',
  'app.ssh.listen.label': 'Listen Address',
  'app.ssh.listen.description':
    'Specify the IP address for SSH service to listen on.',
  'app.ssh.root.label': 'Root User',
  'app.ssh.root.description':
    'SSH login method for root user, allows SSH login by default.',
  'app.ssh.password.label': 'Password Auth',
  'app.ssh.password.description':
    'Whether to enable password authentication, enabled by default.',
  'app.ssh.key.label': 'Key Auth',
  'app.ssh.key.description':
    'Whether to enable key authentication, enabled by default.',
  'app.ssh.passwordInfo.label': 'Password Info',
  'app.ssh.reverse.label': 'Reverse Lookup',
  'app.ssh.reverse.description':
    'Specify whether the SSH service should perform DNS resolution for clients, to speed up connection establishment time.',
  'app.ssh.autostart.label': 'Auto Start',
  'app.ssh.btn.setting': 'Settings',

  // Port setting modal
  'app.ssh.portModal.title': 'Port Settings',
  'app.ssh.portModal.saveSuccess': 'Port settings saved successfully',

  // Listen address setting modal
  'app.ssh.listenModal.title': 'Listen Address Settings',
  'app.ssh.listenModal.saveSuccess':
    'Listen address settings saved successfully',

  // Root user setting modal
  'app.ssh.rootModal.title': 'Root User Settings',
  'app.ssh.rootModal.allow': 'Allow SSH Login',
  'app.ssh.rootModal.deny': 'Deny SSH Login',
  'app.ssh.rootModal.saveSuccess': 'Root user settings saved successfully',

  // Password tab
  'app.ssh.password.generateKey': 'Generate Key',
  'app.ssh.password.hasPassword': 'Has Password',
  'app.ssh.password.noPassword': 'No Password',
  'app.ssh.password.download': 'Download',
  'app.ssh.password.set': 'Set Password',
  'app.ssh.password.update': 'Update Password',
  'app.ssh.password.clear': 'Clear Password',
  'app.ssh.password.delete': 'Delete',

  'app.ssh.password.columns.keyName': 'Key Name',
  'app.ssh.password.columns.encryptionMode': 'Encryption',
  'app.ssh.password.columns.keyBits': 'Key Bits',
  'app.ssh.password.columns.password': 'Password',
  'app.ssh.password.columns.createTime': 'Create Time',
  'app.ssh.password.columns.enabled': 'Enabled',

  'app.ssh.password.generateSuccess': 'SSH key generated successfully',
  'app.ssh.password.enableSuccess': 'SSH key enabled successfully',
  'app.ssh.password.disableSuccess': 'SSH key disabled successfully',
  'app.ssh.password.downloadSuccess': 'SSH key downloaded successfully',
  'app.ssh.password.setSuccess': 'SSH key password set successfully',
  'app.ssh.password.updateSuccess': 'SSH key password updated successfully',
  'app.ssh.password.clearSuccess': 'SSH key password cleared successfully',
  'app.ssh.password.deleteSuccess': 'SSH key deleted successfully',
  'app.ssh.password.operationFailed': 'Operation failed',

  'app.ssh.password.clearConfirm':
    'Are you sure you want to clear the password for key "{keyName}"?',
  'app.ssh.password.deleteConfirm':
    'Are you sure you want to delete key "{keyName}"?',

  // Generate key modal
  'app.ssh.password.generateModal.title': 'Generate SSH Key',
  'app.ssh.password.generateModal.keyName': 'Key Name',
  'app.ssh.password.generateModal.encryptionMode': 'Encryption Mode',
  'app.ssh.password.generateModal.keyBits': 'Key Bits',
  'app.ssh.password.generateModal.password': 'Password',
  'app.ssh.password.generateModal.enable': 'Enable After Generation',
  'app.ssh.password.generateModal.keyNameRequired': 'Please enter a key name',
  'app.ssh.password.generateModal.encryptionModeRequired':
    'Please select an encryption mode',
  'app.ssh.password.generateModal.keyBitsRequired': 'Please select key bits',

  // Set password modal
  'app.ssh.password.setModal.title': 'Set Key Password',
  'app.ssh.password.setModal.password': 'Password',
  'app.ssh.password.setModal.passwordRequired': 'Please enter a password',

  // Update password modal
  'app.ssh.password.updateModal.title': 'Update Key Password',
  'app.ssh.password.updateModal.oldPassword': 'Old Password',
  'app.ssh.password.updateModal.newPassword': 'New Password',
  'app.ssh.password.updateModal.oldPasswordRequired':
    'Please enter the old password',
  'app.ssh.password.updateModal.newPasswordRequired':
    'Please enter a new password',

  // Auth Key tab
  'app.ssh.authKey.add': 'Add Key',
  'app.ssh.authKey.remove': 'Remove',
  'app.ssh.authKey.addSuccess': 'SSH key added successfully',
  'app.ssh.authKey.removeSuccess': 'SSH key removed successfully',
  'app.ssh.authKey.columns.algorithm': 'Algorithm',
  'app.ssh.authKey.columns.key': 'Key',
  'app.ssh.authKey.columns.comment': 'Comment',
  'app.ssh.authKey.columns.operations': 'Operations',
  'app.ssh.authKey.modal.title': 'Add SSH Key',
  'app.ssh.authKey.modal.content': 'Key Content',
  'app.ssh.authKey.modal.placeholder': 'Paste your SSH public key here',
  'app.ssh.authKey.modal.description': 'Format: ssh-rsa AAAAB3NzaC1... comment',
  'app.ssh.authKey.modal.emptyError': 'Key content cannot be empty',
  'app.ssh.authKey.modal.formatError': 'Invalid key format',
  'app.ssh.authKey.removeModal.title': 'Remove SSH Key',
  'app.ssh.authKey.removeModal.content':
    'Are you sure you want to remove this SSH key?',

  // Status component
  'app.ssh.status.running': 'Running',
  'app.ssh.status.stopped': 'Stopped',
  'app.ssh.status.starting': 'Starting...',
  'app.ssh.status.stopping': 'Stopping...',
  'app.ssh.status.unknown': 'Unknown',
  'app.ssh.status.error': 'Error',
  'app.ssh.status.unhealthy': 'Unhealthy',
  'app.ssh.status.stop': 'Stop',
  'app.ssh.status.reload': 'Reload',
  'app.ssh.status.restart': 'Restart',
  'app.ssh.status.stopSuccess': 'SSH service stopped successfully',
  'app.ssh.status.stopFailed': 'Failed to stop SSH service',
  'app.ssh.status.reloadSuccess': 'SSH service reloaded successfully',
  'app.ssh.status.reloadFailed': 'Failed to reload SSH service',
  'app.ssh.status.restartSuccess': 'SSH service restarted successfully',
  'app.ssh.status.restartFailed': 'Failed to restart SSH service',
  'app.ssh.status.autoStartEnabled': 'Auto-start Enabled',
  'app.ssh.status.autoStartDisabled': 'Auto-start Disabled',
};
