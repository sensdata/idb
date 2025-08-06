export default {
  'app.nftables.title': 'NFTables Firewall Management',
  'app.nftables.description':
    'Manage server firewall rule configurations with global and local modes',

  // Menu items
  'app.nftables.menu.config': 'Application Configuration',
  'app.nftables.menu.ports': 'Port Configuration',
  'app.nftables.menu.ipBlacklist': 'IP Blacklist',
  'app.nftables.menu.ping': 'Ping Configuration',

  // Page titles
  'app.nftables.ports.pageTitle': 'Port Configuration',
  'app.nftables.ipBlacklist.pageTitle': 'IP Blacklist Management',
  'app.nftables.ping.pageTitle': 'Ping Configuration',

  // Firewall type selection
  'app.nftables.firewall.title': 'Firewall Type',
  'app.nftables.firewall.nftables.desc':
    'Modern firewall management tool with support for complex rule configurations and advanced features',
  'app.nftables.firewall.iptables.desc':
    'Traditional firewall management tool, widely supported but with relatively simple functionality',
  'app.nftables.firewall.iptables.notSupported':
    'IPTables Configuration Not Supported',
  'app.nftables.firewall.iptables.notSupportedDesc':
    'This system only supports visual configuration management for NFTables. For IPTables usage, please configure through terminal or other tools.',

  // Service management
  'app.nftables.service.title': 'NFTables Service Status',
  'app.nftables.service.control': 'Service Control',

  // Buttons
  'app.nftables.button.refresh': 'Refresh Status',
  'app.nftables.button.install': 'Install Now',
  'app.nftables.button.start': 'Start Service',
  'app.nftables.button.stop': 'Stop Service',
  'app.nftables.button.restart': 'Restart Service',
  'app.nftables.button.test': 'Test Config',
  'app.nftables.button.load': 'Load Config',
  'app.nftables.button.save': 'Save Config',
  'app.nftables.button.apply': 'Apply Config',
  'app.nftables.button.format': 'Format',
  'app.nftables.button.history': 'History',
  'app.nftables.button.switchToNftables': 'Switch to NFTables',
  'app.nftables.button.switchToIptables': 'Switch to IPTables',
  'app.nftables.button.activateConfig': 'Activate This Config',
  'app.nftables.button.addRule': 'New Rule',
  'app.nftables.button.configPorts': 'Configure Ports',
  'app.nftables.button.scan': 'Scan',
  'app.nftables.button.addIP': 'Add IP',
  'app.nftables.button.configIPBlacklist': 'Configure IP Blacklist',
  'app.nftables.button.configPort': 'Configure Port',

  // Status
  'app.nftables.status.installed': 'Installed',
  'app.nftables.status.notInstalled': 'Not Installed',
  'app.nftables.status.running': 'Running Status',
  'app.nftables.status.configured': 'Configuration Status',
  'app.nftables.status.ruleActive': 'Rules Active',
  'app.nftables.status.needConfig': 'Need Configuration',
  'app.nftables.status.stopped': 'Stopped',
  'app.nftables.status.unknown': 'Unknown',
  'app.nftables.status.nftablesActive': 'NFTables Active',
  'app.nftables.status.iptablesLegacyActive': 'IPTables (Legacy) Active',
  'app.nftables.status.iptablesNftActive':
    'IPTables-NFT (Compatibility Layer) Active',
  'app.nftables.status.noFirewall': 'No Firewall System',
  'app.nftables.status.uncertain': 'Uncertain State',
  'app.nftables.status.currentActive': 'Currently Active',
  'app.nftables.status.inactive': 'Inactive',
  'app.nftables.status.stop': 'Stop',
  'app.nftables.status.reload': 'Reload',
  'app.nftables.status.restart': 'Restart',
  'app.nftables.status.starting': 'Starting',
  'app.nftables.status.stopping': 'Stopping',
  'app.nftables.status.error': 'Error',
  'app.nftables.status.unhealthy': 'Unhealthy',
  'app.nftables.status.autoStartEnabled': 'Auto-start Enabled',
  'app.nftables.status.autoStartDisabled': 'Auto-start Disabled',

  // Installation guide
  'app.nftables.installation.title': 'Install NFTables',
  'app.nftables.installation.step1.title': 'Check System Dependencies',
  'app.nftables.installation.step1.desc':
    'Check if the system has the necessary conditions to install NFTables',
  'app.nftables.installation.step2.title': 'Download and Install',
  'app.nftables.installation.step2.desc':
    'Download and install NFTables package from software repository',
  'app.nftables.installation.step3.title': 'Configure Service',
  'app.nftables.installation.step3.desc':
    'Enable NFTables service and perform basic configuration',
  'app.nftables.installation.log.title': 'Installation Log',
  'app.nftables.installation.log.clear': 'Clear',

  // Configuration management
  'app.nftables.config.title': 'Application Configuration',
  'app.nftables.config.firewallStatus': 'Firewall Status',
  'app.nftables.config.installStatus': 'Installation Status',
  'app.nftables.config.activeSystem': 'Current Active System',
  'app.nftables.config.modeChanged': 'Switched to {mode} mode',
  'app.nftables.config.mode': 'Configuration Mode',
  'app.nftables.config.configScope': 'Configuration Scope',
  'app.nftables.config.editMode': 'Edit Mode',

  // Visual configuration
  'app.nftables.config.visual.title': 'Visual Configuration',
  'app.nftables.config.visual.desc':
    'Configure firewall rules through form interface, simple and intuitive',
  'app.nftables.config.visual.developing': 'Feature Under Development',
  'app.nftables.config.visual.developingDesc':
    'The visual configuration interface is under development. Please use file mode for configuration temporarily.',

  // File configuration
  'app.nftables.config.file.title': 'File Configuration',
  'app.nftables.config.file.desc':
    'Directly edit NFTables configuration file, suitable for advanced users',
  'app.nftables.config.file.exists': 'Exists',
  'app.nftables.config.file.new': 'New',

  // Form configuration
  'app.nftables.config.form.title': 'Form Configuration',
  'app.nftables.config.form.desc':
    'Configure port rules through form interface, simple and intuitive',
  'app.nftables.form.portHelp': 'Single port: 80, Port list: 80, 443, 9918',
  'app.nftables.form.updateRule': 'Update Rule',
  'app.nftables.form.cancel': 'Cancel',
  'app.nftables.form.sourceOptional': 'Source (Optional)',
  'app.nftables.form.sourcePlaceholder': 'e.g. 192.168.1.0/24 or 0.0.0.0/0',

  // Advanced rules
  'app.nftables.form.advancedRules': 'Advanced Rules',
  'app.nftables.form.rulesList': 'Rules List',
  'app.nftables.form.addRule': 'Add Rule',
  'app.nftables.form.noRules': 'No rules',
  'app.nftables.form.rule': 'Rule',
  'app.nftables.form.ruleType': 'Rule Type',
  'app.nftables.form.basicRule': 'Basic Rule',
  'app.nftables.form.rateLimit': 'Rate Limit',
  'app.nftables.form.concurrentLimit': 'Concurrent Limit',
  'app.nftables.form.rateValue': 'Rate Value',
  'app.nftables.form.ratePlaceholder': 'e.g. 100/second',
  'app.nftables.form.concurrentCount': 'Concurrent Count',
  'app.nftables.form.concurrentPlaceholder': 'e.g. 10',

  // Configuration types
  'app.nftables.config.type.local': 'Local',
  'app.nftables.config.type.global': 'Global',
  'app.nftables.config.globalDescription':
    'Global configuration will be applied to all hosts',
  'app.nftables.config.localDescription':
    'Local configuration only applies to the current host',
  'app.nftables.config.activateHint':
    'Currently viewing {viewing} configuration, but {active} configuration is active',

  // Control bar
  'nftables.controlBar.title': 'Configuration Management',
  'nftables.controlBar.currentStatus':
    'Current Config: {configType}, Current Mode: {configMode}',

  // Config types
  'nftables.configType.local': 'Local',
  'nftables.configType.global': 'Global',

  // Config modes
  'nftables.configMode.form': 'Form',
  'nftables.configMode.file': 'File',

  // Process status
  'app.nftables.config.processStatusList': 'Process Status List',
  'app.nftables.config.columns.process': 'Process',
  'app.nftables.config.columns.port': 'Port',
  'app.nftables.config.columns.addresses': 'Listening Addresses',
  'app.nftables.config.accessible': 'Accessible',
  'app.nftables.config.fullyAccessible': 'Fully Accessible',
  'app.nftables.config.rejected': 'Rejected',
  'app.nftables.config.restricted': 'Restricted',
  'app.nftables.config.unknown': 'Unknown',
  'app.nftables.config.notAccessible': 'Restricted',
  'app.nftables.config.localOnly': 'Local Only',

  // Configuration modes
  'app.nftables.config.formMode': 'Form Mode',
  'app.nftables.config.visualMode': 'Visual Mode',
  'app.nftables.config.fileMode': 'File Mode',

  // Basic policy
  'app.nftables.config.policy.title': 'Basic Policy',
  'app.nftables.config.policy.input': 'Input Policy',
  'app.nftables.config.policy.output': 'Output Policy',
  'app.nftables.config.policy.accept': 'ACCEPT (Allow)',
  'app.nftables.config.policy.drop': 'DROP (Deny)',

  // Port rules
  'app.nftables.config.rules.title': 'Port Rules',
  'app.nftables.config.rules.protocol': 'Protocol',
  'app.nftables.config.rules.port': 'Port',
  'app.nftables.config.rules.action': 'Action',
  'app.nftables.config.rules.source': 'Source',
  'app.nftables.config.rules.description': 'Description',
  'app.nftables.config.rules.allow': 'Allow',
  'app.nftables.config.rules.deny': 'Deny',
  'app.nftables.config.rules.unknown': 'Unknown',
  'app.nftables.config.rules.portPlaceholder': '80',
  'app.nftables.config.rules.descPlaceholder': 'Rule description',
  'app.nftables.config.rules.deleteConfirm':
    'Are you sure to delete the rule for port {port}?',

  // Ping configuration
  'app.nftables.ping.title': 'Ping Configuration',
  'app.nftables.ping.description':
    'Configure whether the server responds to ICMP ping requests',
  'app.nftables.ping.currentStatus': 'Current Status',
  'app.nftables.ping.allowed': 'Ping Allowed',
  'app.nftables.ping.blocked': 'Ping Blocked',
  'app.nftables.ping.allowPing': 'Allow Ping',
  'app.nftables.ping.blockPing': 'Block Ping',
  'app.nftables.ping.statusDescription.allowed':
    'The server currently responds to ICMP ping requests from external sources.',
  'app.nftables.ping.statusDescription.blocked':
    'The server currently blocks ICMP ping requests from external sources.',
  'app.nftables.ping.configureTitle': 'Ping Settings',
  'app.nftables.ping.enableLabel': 'Allow external ping requests',
  'app.nftables.ping.enableHelp':
    'When enabled, the server will respond to ICMP ping requests. When disabled, ping requests will be blocked by the firewall.',
  'app.nftables.ping.applySettings': 'Apply Settings',
  'app.nftables.ping.loading': 'Loading ping status...',
  'app.nftables.ping.saving': 'Saving settings...',
  'app.nftables.ping.saveSuccess': 'Ping settings updated successfully',
  'app.nftables.ping.saveFailed': 'Failed to update ping settings',
  'app.nftables.ping.loadFailed': 'Failed to load ping status',
  'app.nftables.config.rules.empty': 'No port rules',

  // Editor
  'app.nftables.config.editor.placeholder':
    'Please enter NFTables configuration...',
  'app.nftables.config.editor.lineCount': '{count} lines',
  'app.nftables.config.editor.reloadTip':
    'Note: After saving the configuration, you need to reload the firewall rules for them to take effect',

  // Mode switch
  'app.nftables.config.modeSwitch.form': 'Switched to form mode',
  'app.nftables.config.modeSwitch.file': 'Switched to file mode',

  // Editor (legacy)
  'app.nftables.editor.placeholder': 'Please enter nftables configuration...',

  // Common status
  'app.nftables.common.installed': 'Installed',
  'app.nftables.common.notInstalled': 'Not Installed',
  'app.nftables.common.running': 'Running',
  'app.nftables.common.stopped': 'Stopped',
  'app.nftables.common.configured': 'Configured',
  'app.nftables.common.notConfigured': 'Not Configured',

  // Tab labels
  'app.nftables.tabs.overview': 'Overview',
  'app.nftables.tabs.service': 'Service Management',
  'app.nftables.tabs.config': 'Configuration Management',
  'app.nftables.tabs.monitoring': 'Monitoring & Logs',

  // Overview page
  'app.nftables.overview.status': 'System Status',
  'app.nftables.overview.quickActions': 'Quick Actions',
  'app.nftables.overview.startConfig': 'Start Configuration',
  'app.nftables.overview.editConfig': 'Edit Configuration',

  // Monitoring & logs
  'app.nftables.monitoring.title': 'Monitoring Center',
  'app.nftables.monitoring.developing': 'Feature Under Development',
  'app.nftables.monitoring.developingDesc':
    'Monitoring and logging features are under development. Stay tuned.',

  // Port configuration
  'app.nftables.ports.title': 'Port Configuration',

  // IP Blacklist configuration
  'app.nftables.ipBlacklist.title': 'IP Blacklist',
  'app.nftables.ipBlacklist.description':
    'Manage IP blacklist rules to block malicious IP access',
  'app.nftables.ipBlacklist.form.title': 'Visual Configuration',
  'app.nftables.ipBlacklist.form.desc':
    'Configure IP blacklist rules through form interface, simple and intuitive',
  'app.nftables.ipBlacklist.file.title': 'File Configuration',
  'app.nftables.ipBlacklist.file.desc':
    'Directly edit NFTables IP blacklist configuration file',
  'app.nftables.ipBlacklist.rules.title': 'IP Blacklist Rules',
  'app.nftables.ipBlacklist.rules.empty': 'No IP blacklist rules yet',
  'app.nftables.ipBlacklist.rules.ip': 'IP Address',
  'app.nftables.ipBlacklist.rules.type': 'Type',
  'app.nftables.ipBlacklist.rules.description': 'Description',
  'app.nftables.ipBlacklist.rules.action': 'Action',
  'app.nftables.ipBlacklist.rules.createdAt': 'Created At',
  'app.nftables.ipBlacklist.rules.ipPlaceholder':
    '192.168.1.100 or 192.168.1.0/24',
  'app.nftables.ipBlacklist.rules.descPlaceholder': 'Rule description',
  'app.nftables.ipBlacklist.rules.deleteConfirm':
    'Are you sure you want to delete the blacklist rule for IP {ip}?',
  'app.nftables.ipBlacklist.rules.dropHint':
    'All IP addresses added to the blacklist will be dropped, preventing access to the server',
  'app.nftables.ipBlacklist.rules.invalidIPFormat':
    'Please enter a valid IP address format',
  'app.nftables.ipBlacklist.rules.ipFormatHint':
    'Supports single IP (192.168.1.100), CIDR network (192.168.1.0/24), or IP range (192.168.1.1-192.168.1.10)',
  'app.nftables.ipBlacklist.type.single': 'Single IP',
  'app.nftables.ipBlacklist.type.cidr': 'CIDR Network',
  'app.nftables.ipBlacklist.type.range': 'IP Range',
  'app.nftables.ipBlacklist.action.drop': 'DROP (Drop)',
  'app.nftables.ipBlacklist.action.reject': 'REJECT (Reject)',

  // Policy and rules
  'app.nftables.policy.title': 'Basic Policy',
  'app.nftables.rules.title': 'Port Rules',
  'app.nftables.rules.empty': 'No rules yet',
  'app.nftables.rules.deleteConfirm':
    'Are you sure you want to delete the rule for port {port}?',
  'app.nftables.services.title': 'Application Services',
  'app.nftables.templates.title': 'Preset Templates',

  // Generic text
  'app.nftables.loading': 'Loading...',
  'app.nftables.savingChanges': 'Saving changes...',

  // Messages
  'app.nftables.message.statusRefreshed': 'Status refreshed successfully',
  'app.nftables.message.refreshFailed': 'Status refresh failed',
  'app.nftables.message.fetchSuccess':
    'Process data loaded successfully, {count} records in total',
  'app.nftables.message.fetchFailed': 'Failed to fetch data',
  'app.nftables.message.statusFetchSuccess':
    'Firewall status fetched successfully',
  'app.nftables.message.statusFetchFailed': 'Failed to fetch firewall status',
  'app.nftables.message.switchToNftablesSuccess':
    'Successfully switched to NFTables',
  'app.nftables.message.switchToIptablesSuccess':
    'Successfully switched to IPTables',
  'app.nftables.message.switchFailed': 'Failed to switch firewall system',
  'app.nftables.message.installingNftables': 'Installing NFTables...',
  'app.nftables.message.installSuccess': 'NFTables installed successfully',
  'app.nftables.message.installFailed': 'NFTables installation failed',
  'app.nftables.message.serviceStarted': 'Service started successfully',
  'app.nftables.message.startFailed': 'Service start failed',
  'app.nftables.message.serviceStopped': 'Service stopped successfully',
  'app.nftables.message.stopFailed': 'Service stop failed',
  'app.nftables.message.serviceRestarted': 'Service restarted successfully',
  'app.nftables.message.restartFailed': 'Service restart failed',
  'app.nftables.message.configReloaded': 'Configuration reloaded successfully',
  'app.nftables.message.reloadFailed': 'Configuration reload failed',
  'app.nftables.message.configLoaded': 'Configuration loaded successfully',
  'app.nftables.message.loadFailed': 'Configuration load failed',
  'app.nftables.message.configSaved': 'Configuration saved successfully',
  'app.nftables.message.saveFailed': 'Configuration save failed',
  'app.nftables.message.testPassed': 'Configuration test passed',
  'app.nftables.message.testFailed': 'Configuration test failed',
  'app.nftables.message.configFormatted':
    'Configuration formatted successfully',
  'app.nftables.message.formatFailed': 'Configuration format failed',
  'app.nftables.message.configApplied': 'Configuration applied successfully',
  'app.nftables.message.applyFailed': 'Configuration apply failed',
  'app.nftables.message.initializingDefaultConfig':
    'Initializing default configuration...',
  'app.nftables.message.defaultConfigInitialized':
    'Default configuration initialized',
  'app.nftables.message.defaultConfigInitFailed':
    'Default configuration initialization failed',

  // Process descriptions
  'app.nftables.process.unknown': 'Unknown Process',
  'app.nftables.process.nginx': 'Nginx High Performance Web Server',
  'app.nftables.process.mysql': 'MySQL Relational Database',
  'app.nftables.process.mysqld': 'MySQL Database Daemon',
  'app.nftables.process.postgresql': 'PostgreSQL Open Source Database',
  'app.nftables.process.postgres': 'PostgreSQL Database Service',
  'app.nftables.process.redis': 'Redis In-Memory Key-Value Database',
  'app.nftables.process.mongodb': 'MongoDB Document Database',
  'app.nftables.process.apache': 'Apache HTTP Web Server',
  'app.nftables.process.httpd': 'Apache HTTP Daemon',
  'app.nftables.process.docker': 'Docker Container Platform',
  'app.nftables.process.dockerd': 'Docker Daemon',
  'app.nftables.process.dockerProxy': 'Docker Network Proxy',
  'app.nftables.process.dockerContainerd': 'Docker containerd Runtime',
  'app.nftables.process.containerd': 'containerd Container Runtime',
  'app.nftables.process.containerdShim': 'containerd shim Process',
  'app.nftables.process.kubernetes': 'Kubernetes Container Orchestration',
  'app.nftables.process.kubectl': 'Kubernetes Command Line Tool',
  'app.nftables.process.ssh': 'SSH Secure Shell Connection',
  'app.nftables.process.sshd': 'SSH Server Daemon',
  'app.nftables.process.systemd': 'System and Service Manager',
  'app.nftables.process.nodejs': 'Node.js JavaScript Runtime',
  'app.nftables.process.node': 'Node.js Application',
  'app.nftables.process.java': 'Java Virtual Machine Application',
  'app.nftables.process.tomcat': 'Apache Tomcat Web Server',
  'app.nftables.process.php': 'PHP Script Interpreter',
  'app.nftables.process.phpFpm': 'PHP FastCGI Process Manager',
  'app.nftables.process.podman': 'Podman Daemonless Container Engine',
  'app.nftables.process.mariadb': 'MariaDB Database Server',
  'app.nftables.process.bind': 'BIND DNS Server',
  'app.nftables.process.named': 'BIND DNS Daemon',
  'app.nftables.process.fail2ban': 'Fail2ban Intrusion Prevention System',
  'app.nftables.process.iptables': 'iptables Firewall Rules',
  'app.nftables.process.firewall': 'Firewall Service',
  'app.nftables.process.selinux': 'SELinux Security Module',

  // Process category descriptions
  'app.nftables.process.category.database': 'Database Service',
  'app.nftables.process.category.webServer': 'Web Server',
  'app.nftables.process.category.container': 'Container Service',
  'app.nftables.process.category.network': 'Network Service',
  'app.nftables.process.category.security': 'Security Service',
  'app.nftables.process.category.development': 'Development Tool',
  'app.nftables.process.category.system': 'System Service',
  'app.nftables.process.category.application': 'Application',

  // Form related
  'app.nftables.form.source': 'Source Address',
  'app.nftables.form.basicConfig': 'Basic Configuration',
  'app.nftables.form.configMode': 'Configuration Mode',
  'app.nftables.form.simpleMode': 'Simple Mode',
  'app.nftables.form.simpleModeDesc':
    'Suitable for beginners, just select port and action to configure quickly',
  'app.nftables.form.advancedMode': 'Advanced Mode',
  'app.nftables.form.advancedModeDesc':
    'Suitable for advanced users, supports complex rules like rate limiting and concurrent limiting',
  'app.nftables.form.accessControl': 'Access Control',
  'app.nftables.form.allowDesc': 'Allow access to this port',
  'app.nftables.form.denyDesc':
    'Deny access to this port, no response returned',
  'app.nftables.form.rejectDesc':
    'Reject access to this port, return rejection response',
  'app.nftables.form.noRulesHint': 'No advanced rules configured yet',
  'app.nftables.form.addFirstRule': 'Add First Rule',
  'app.nftables.form.configPreview': 'Configuration Preview',
  'app.nftables.form.noRulesPreview': '# No rules configured',
  'app.nftables.form.rateHelpText':
    'Format: amount/time_unit, e.g. 100/second, 50/minute',
  'app.nftables.form.concurrentHelpText':
    'Maximum number of concurrent connections allowed',

  // Form validation
  'app.nftables.validation.portRequired': 'Please enter port number',
  'app.nftables.validation.portRange': 'Port number must be between 1-65535',
  'app.nftables.validation.protocolRequired': 'Please select protocol type',
  'app.nftables.validation.actionRequired': 'Please select action',
  'app.nftables.validation.sourceFormat':
    'Please enter a valid IP address or CIDR format',

  // Drawer titles
  'app.nftables.drawer.addRule': 'Add Port Rule',
  'app.nftables.drawer.editRule': 'Edit Port Rule',
  'app.nftables.drawer.addIPRule': 'Add IP Blacklist Rule',

  // Error messages
  'app.nftables.error.fetchConfigFailed': 'Failed to fetch configuration',
  'app.nftables.error.saveConfigFailed': 'Failed to save configuration',
  'app.nftables.error.refreshConfigFailed': 'Failed to refresh configuration',
  'app.nftables.error.initializeFailed': 'Initialization failed',
  'app.nftables.error.deleteRuleFailed': 'Failed to delete rule',

  // Base Rules Management
  'app.nftables.baseRules.title': 'Base Rules',
  'app.nftables.baseRules.tooltip':
    'Configure firewall base policies that determine default traffic handling',
  'app.nftables.baseRules.inputPolicy': 'Input Policy',
  'app.nftables.baseRules.inputPolicyDescription':
    'Set the default inbound traffic handling policy, affecting all unconfigured ports',

  // Policy Options
  'app.nftables.baseRules.accept': 'Accept',
  'app.nftables.baseRules.acceptDesc':
    'Allow all inbound traffic by default (lower security)',
  'app.nftables.baseRules.drop': 'Drop',
  'app.nftables.baseRules.dropDesc':
    'Drop all inbound traffic by default (high security)',
  'app.nftables.baseRules.reject': 'Reject',
  'app.nftables.baseRules.rejectDesc':
    'Reject all inbound traffic by default and return error message',

  // Security Warning
  'app.nftables.baseRules.warningTitle': 'Security Notice',
  'app.nftables.baseRules.warningDescription':
    'When using "Drop" policy, ensure necessary port rules are properly configured to avoid server access issues',

  // Base Rules Messages
  'app.nftables.message.baseRulesSaved': 'Base rules saved successfully',
  'app.nftables.message.saveBaseRulesFailed': 'Failed to save base rules',
  'app.nftables.message.fetchBaseRulesFailed': 'Failed to fetch base rules',

  // Additional messages
  'app.nftables.message.ruleAdded': 'Rule added successfully',
  'app.nftables.message.ruleUpdated': 'Rule updated successfully',
  'app.nftables.message.ruleDeleted': 'Rule deleted successfully',
  'app.nftables.message.configRefreshed':
    'Configuration refreshed successfully',
  'app.nftables.message.operationFailed': 'Operation failed',
  'app.nftables.message.actionRequired': 'Please select an action',
  'app.nftables.message.protocolRequired': 'Please select a protocol type',
  'app.nftables.message.noHost': 'No host available',

  // Editor specific
  'app.nftables.config.editor.modified': 'Modified',
  'app.nftables.config.editor.tips':
    'Press Ctrl+S to save, or use the Save button above',
  'app.nftables.config.editor.emptyContent':
    'Configuration content cannot be empty',
  'app.nftables.config.editor.unsavedChanges':
    'You have unsaved changes. Are you sure you want to leave?',
  'app.nftables.config.editor.confirmRefresh': 'Confirm Refresh',
  'app.nftables.config.editor.confirmRefreshContent':
    'You have unsaved changes. Refreshing will discard these changes. Continue?',
};
