export default {
  'app.ssh.pageTitle': 'SSH Management',
  'app.ssh.tabs.config': 'SSH Config',
  'app.ssh.tabs.keyPairs': 'Key Pairs Management',
  'app.ssh.tabs.publicKeys': 'Public Keys Management',
  'app.ssh.config.title': 'SSH Configuration',
  'app.ssh.keyPairs.title': 'Key Pairs Management',
  'app.ssh.publicKeys.title': 'Public Keys Management',

  // Mode switch
  'app.ssh.mode.visual': 'Form Mode',
  'app.ssh.mode.source': 'Source Mode',
  'app.ssh.mode.switchConfirmTitle': 'Confirm Mode Switch',
  'app.ssh.mode.switchConfirmContent':
    'Switching to form mode will parse the current source configuration and update the form. Unsaved changes may be lost. Continue?',

  // Source mode
  'app.ssh.source.save': 'Save',
  'app.ssh.source.reset': 'Reset',
  'app.ssh.source.placeholder': 'Edit SSH configuration file content here',
  'app.ssh.source.info':
    'This text editor allows direct editing of the SSH configuration file. After making changes, click the "Save" button to apply them.',
  'app.ssh.source.saveSuccess': 'Source configuration saved successfully',
  'app.ssh.source.saveError': 'Failed to save source configuration',
  'app.ssh.source.resetSuccess': 'Source configuration reset',
  'app.ssh.source.parseSuccess': 'Source configuration parsed successfully',
  'app.ssh.source.parseError': 'Failed to parse source configuration',
  'app.ssh.source.emptyConfig': 'Configuration content is empty',
  'app.ssh.source.noChanges': 'No valid configuration changes detected',
  'app.ssh.savingConfig': 'Saving SSH configuration...',

  // Unsaved changes modal
  'app.ssh.unsavedChanges.title': 'Unsaved Changes',
  'app.ssh.unsavedChanges.content':
    'You have unsaved changes in the source editor. Switching to form mode will discard these changes. Do you want to continue?',
  'app.ssh.unsavedChanges.discard': 'Discard Changes',
  'app.ssh.unsavedChanges.cancel': 'Cancel',

  // Config updates
  'app.ssh.config.updateSuccess': 'Configuration updated successfully',
  'app.ssh.config.updateError': 'Failed to update configuration',

  // Error messages
  'app.ssh.error.fetchFailed': 'Failed to fetch SSH configuration',
  'app.ssh.error.noHost': 'No host selected',
  'app.ssh.error.emptyConfig': 'Empty SSH configuration received',
  'app.ssh.error.fetchError': 'Failed to fetch SSH configuration content',

  // Loading message
  'app.ssh.loading': 'Loading SSH configuration...',
  'app.ssh.savingChanges': 'Saving configuration changes...',

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
  'app.ssh.key.label': 'Public Key Auth',
  'app.ssh.key.description':
    'Whether to enable public key authentication, enabled by default.',
  'app.ssh.passwordInfo.label': 'Password Info',
  'app.ssh.reverse.label': 'Reverse Lookup',
  'app.ssh.reverse.description':
    'Specify whether the SSH service should perform DNS resolution for clients, to speed up connection establishment time.',
  'app.ssh.sftp.label': 'SFTP Subsystem',
  'app.ssh.sftp.description':
    'Enable or disable the SFTP (SSH File Transfer Protocol) subsystem for secure file transfers.',
  'app.ssh.autostart.label': 'Auto Start',
  'app.ssh.btn.setting': 'Settings',

  // Port setting modal
  'app.ssh.portModal.title': 'Port Settings',
  'app.ssh.portModal.port': 'Port Number',
  'app.ssh.portModal.description':
    'Specify the port number for SSH service, default is 22.',
  'app.ssh.portModal.save': 'Save',
  'app.ssh.portModal.cancel': 'Cancel',
  'app.ssh.portModal.saveSuccess': 'Port settings saved successfully',
  'app.ssh.portModal.saveError': 'Failed to save port settings',

  // Listen address setting modal
  'app.ssh.listenModal.title': 'Listen Address Settings',
  'app.ssh.listenModal.address': 'Listen Address',
  'app.ssh.listenModal.description':
    'Specify the IP address for SSH service to listen on, default is 0.0.0.0.',
  'app.ssh.listenModal.save': 'Save',
  'app.ssh.listenModal.cancel': 'Cancel',
  'app.ssh.listenModal.saveSuccess':
    'Listen address settings saved successfully',
  'app.ssh.listenModal.saveError': 'Failed to save listen address settings',

  // Root user setting modal
  'app.ssh.rootModal.title': 'Root User Settings',
  'app.ssh.rootModal.label': 'Allow root user login',
  'app.ssh.rootModal.description':
    'Whether to allow root user to login via SSH, enabled by default.',
  'app.ssh.rootModal.save': 'Save',
  'app.ssh.rootModal.cancel': 'Cancel',
  'app.ssh.rootModal.allow': 'Allow SSH Login',
  'app.ssh.rootModal.deny': 'Deny SSH Login',
  'app.ssh.rootModal.saveSuccess': 'Root user settings saved successfully',
  'app.ssh.rootModal.saveError': 'Failed to save root user settings',

  // Password tab
  'app.ssh.keyPairs.generateKey': 'Generate Key',
  'app.ssh.keyPairs.hasPassword': 'Has Password',
  'app.ssh.keyPairs.noPassword': 'No Password',
  'app.ssh.keyPairs.download': 'Download',
  'app.ssh.keyPairs.set': 'Set Password',
  'app.ssh.keyPairs.update': 'Update Password',
  'app.ssh.keyPairs.clear': 'Clear Password',
  'app.ssh.keyPairs.delete': 'Delete',

  'app.ssh.keyPairs.columns.keyName': 'Key Name',
  'app.ssh.keyPairs.columns.encryptionMode': 'Encryption',
  'app.ssh.keyPairs.columns.keyBits': 'Key Bits',
  'app.ssh.keyPairs.columns.password': 'Password',
  'app.ssh.keyPairs.columns.createTime': 'Create Time',
  'app.ssh.keyPairs.columns.enabled': 'Enabled',
  'app.ssh.keyPairs.columns.user': 'User',
  'app.ssh.keyPairs.columns.keyPath': 'Key Path',
  'app.ssh.keyPairs.columns.fingerprint': 'Fingerprint',
  'app.ssh.keyPairs.columns.status': 'Status',
  'app.ssh.keyPairs.enabled': 'Enabled',
  'app.ssh.keyPairs.disabled': 'Disabled',
  'app.ssh.keyPairs.enable': 'Enable',
  'app.ssh.keyPairs.disable': 'Disable',

  'app.ssh.keyPairs.generateSuccess': 'SSH key generated successfully',
  'app.ssh.keyPairs.enableSuccess': 'SSH key enabled successfully',
  'app.ssh.keyPairs.disableSuccess': 'SSH key disabled successfully',
  'app.ssh.keyPairs.downloadSuccess': 'SSH key downloaded successfully',
  'app.ssh.keyPairs.setSuccess': 'SSH key password set successfully',
  'app.ssh.keyPairs.updateSuccess': 'SSH key password updated successfully',
  'app.ssh.keyPairs.clearSuccess': 'SSH key password cleared successfully',
  'app.ssh.keyPairs.deleteSuccess': 'SSH key deleted successfully',
  'app.ssh.keyPairs.operationFailed': 'Operation failed',

  'app.ssh.keyPairs.clearConfirm':
    'Are you sure you want to clear the password for key "{keyName}"?',
  'app.ssh.keyPairs.deleteConfirm':
    'Are you sure you want to delete key "{keyName}"?',

  // Generate key modal
  'app.ssh.keyPairs.generateModal.title': 'Generate SSH Key',
  'app.ssh.keyPairs.generateModal.keyName': 'Key Name',
  'app.ssh.keyPairs.generateModal.encryptionMode': 'Encryption Mode',
  'app.ssh.keyPairs.generateModal.keyBits': 'Key Bits',
  'app.ssh.keyPairs.generateModal.password': 'Password',
  'app.ssh.keyPairs.generateModal.enable': 'Enable After Generation',
  'app.ssh.keyPairs.generateModal.keyNameRequired': 'Please enter a key name',
  'app.ssh.keyPairs.generateModal.encryptionModeRequired':
    'Please select an encryption mode',
  'app.ssh.keyPairs.generateModal.keyBitsRequired': 'Please select key bits',

  // Set password modal
  'app.ssh.keyPairs.setModal.title': 'Set Key Password',
  'app.ssh.keyPairs.setModal.password': 'Password',
  'app.ssh.keyPairs.setModal.passwordRequired': 'Please enter a password',

  // Update password modal
  'app.ssh.keyPairs.updateModal.title': 'Update Key Password',
  'app.ssh.keyPairs.updateModal.oldPassword': 'Old Password',
  'app.ssh.keyPairs.updateModal.newPassword': 'New Password',
  'app.ssh.keyPairs.updateModal.oldPasswordRequired':
    'Please enter the old password',
  'app.ssh.keyPairs.updateModal.newPasswordRequired':
    'Please enter a new password',

  // Auth Key tab
  'app.ssh.publicKeys.add': 'Add Key',
  'app.ssh.publicKeys.addPublicKey': 'Add Public Key',
  'app.ssh.publicKeys.remove': 'Remove',
  'app.ssh.publicKeys.addSuccess': 'SSH key added successfully',
  'app.ssh.publicKeys.removeSuccess': 'SSH key removed successfully',
  'app.ssh.publicKeys.addError': 'Failed to add SSH key',
  'app.ssh.publicKeys.removeError': 'Failed to remove SSH key',
  'app.ssh.publicKeys.loadError': 'Failed to load SSH keys',
  'app.ssh.publicKeys.columns.algorithm': 'Algorithm',
  'app.ssh.publicKeys.columns.key': 'Key',
  'app.ssh.publicKeys.columns.comment': 'Comment',
  'app.ssh.publicKeys.columns.operations': 'Operations',
  'app.ssh.publicKeys.modal.title': 'Add SSH Key',
  'app.ssh.publicKeys.modal.addPublicKey': 'Add Public Key',
  'app.ssh.publicKeys.modal.content': 'Key Content',
  'app.ssh.publicKeys.modal.placeholder': 'Paste your SSH public key here',
  'app.ssh.publicKeys.modal.description':
    'Format: ssh-rsa AAAAB3NzaC1... comment',
  'app.ssh.publicKeys.modal.emptyError': 'Key content cannot be empty',
  'app.ssh.publicKeys.modal.formatError': 'Invalid key format',
  'app.ssh.publicKeys.modal.invalidAlgorithm': 'Invalid SSH key algorithm',
  'app.ssh.publicKeys.modal.invalidKeyFormat':
    'Invalid key format (must be Base64 encoded)',
  'app.ssh.publicKeys.modal.keyTooShort': 'SSH key is too short',
  'app.ssh.publicKeys.removeModal.title': 'Remove SSH Key',
  'app.ssh.publicKeys.removeModal.content':
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
