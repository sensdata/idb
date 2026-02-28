export default {
  'app.certificate.title': 'Certificate Management',
  'app.certificate.createGroup': 'Create Certificate Group',
  'app.certificate.import': 'Import Certificate',
  'app.certificate.updateCertificate': 'Update Certificate',
  'app.certificate.generateSelfSigned': 'Generate Self-Signed Certificate',
  'app.certificate.viewCertificate': 'View Certificate',
  'app.certificate.viewCertificates': 'View Certificate List',
  'app.certificate.completeChain': 'Complete Certificate Chain',
  'app.certificate.deleteGroup': 'Delete Certificate Group',
  'app.certificate.deleteCertificate': 'Delete Certificate',
  'app.certificate.viewPrivateKey': 'View Private Key',
  'app.certificate.viewCSR': 'View CSR',
  'app.certificate.certificateActions': 'Certificate Actions',
  'app.certificate.groupActions': 'Group Actions',
  'app.certificate.noCertificateActions': 'No certificate actions available',
  'app.certificate.certificateDetail': 'Certificate Details',
  'app.certificate.privateKeyInfo': 'Private Key Information',
  'app.certificate.privateKeyPath': 'Private Key Path',
  'app.certificate.csrInfo': 'CSR Information',

  // Table headers
  'app.certificate.alias': 'Alias',
  'app.certificate.requester': 'Requester',
  'app.certificate.certificates': 'Certificates',
  'app.certificate.domain': 'Domain',
  'app.certificate.status': 'Status',
  'app.certificate.chainStatus': 'Chain',
  'app.certificate.expiresOn': 'Expires On',
  'app.certificate.noCertificates': 'No certificates',
  'app.certificate.certificatesCount': 'certificates',

  // Status
  'app.certificate.status.valid': 'Valid',
  'app.certificate.status.expired': 'Expired',
  'app.certificate.status.expiringSoon': 'Expiring Soon ({days} days)',
  'app.certificate.status.notYetValid': 'Not Yet Valid',
  'app.certificate.status.unknown': 'Unknown',
  'app.certificate.chainStatus.completed': 'Completed',
  'app.certificate.chainStatus.incomplete': 'Incomplete',
  'app.certificate.chainStatus.unknown': 'Unknown',

  // Form fields
  'app.certificate.form.alias': 'Alias',
  'app.certificate.form.aliasPlaceholder': 'Enter certificate group alias',
  'app.certificate.form.aliasRequired': 'Please enter alias',
  'app.certificate.form.aliasFormat':
    'Alias can only contain letters, numbers, underscores and hyphens',

  'app.certificate.form.domainName': 'Domain Name',
  'app.certificate.form.domainNamePlaceholder': 'Enter domain name',
  'app.certificate.form.domainNameRequired': 'Please enter domain name',

  'app.certificate.form.email': 'Email',
  'app.certificate.form.emailPlaceholder': 'Enter email address',
  'app.certificate.form.emailRequired': 'Please enter email address',
  'app.certificate.form.emailFormat': 'Please enter a valid email address',

  'app.certificate.form.organization': 'Organization',
  'app.certificate.form.organizationPlaceholder': 'Enter organization name',
  'app.certificate.form.organizationRequired': 'Please enter organization name',

  'app.certificate.form.organizationUnit': 'Organization Unit',
  'app.certificate.form.organizationUnitPlaceholder': 'Enter organization unit',

  'app.certificate.form.country': 'Country',
  'app.certificate.form.countryPlaceholder': 'Enter country code (e.g., US)',
  'app.certificate.form.countryRequired': 'Please enter country code',
  'app.certificate.form.countryFormat': 'Country code must be 2 characters',

  'app.certificate.form.province': 'Province',
  'app.certificate.form.provincePlaceholder': 'Enter province',

  'app.certificate.form.city': 'City',
  'app.certificate.form.cityPlaceholder': 'Enter city',

  'app.certificate.form.keyAlgorithm': 'Key Algorithm',
  'app.certificate.form.keyAlgorithmPlaceholder': 'Select key algorithm',
  'app.certificate.form.keyAlgorithmRequired': 'Please select key algorithm',

  'app.certificate.form.expireValue': 'Validity Period',
  'app.certificate.form.expireValuePlaceholder': 'Enter validity period',
  'app.certificate.form.expireValueRequired': 'Please enter validity period',
  'app.certificate.form.expireValueRange':
    'Validity period must be between 1-9999',

  'app.certificate.form.expireUnit': 'Validity Unit',
  'app.certificate.form.expireUnitPlaceholder': 'Select validity unit',
  'app.certificate.form.expireUnitRequired': 'Please select validity unit',
  'app.certificate.form.days': 'Days',
  'app.certificate.form.years': 'Years',

  'app.certificate.form.altDomains': 'Alternative Domains',
  'app.certificate.form.altDomainsPlaceholder':
    'One domain per line, wildcards supported (e.g., *.example.com)',
  'app.certificate.form.altDomainsHelp':
    'Enter one domain per line, wildcards supported',
  'app.certificate.form.altDomainsInvalid': 'Invalid domain: {domain}',

  'app.certificate.form.altIPs': 'Alternative IP Addresses',
  'app.certificate.form.altIPsPlaceholder': 'One IP address per line',
  'app.certificate.form.altIPsHelp':
    'Enter one IP address per line, IPv4 and IPv6 supported',
  'app.certificate.form.altIPsInvalid': 'Invalid IP address: {ip}',

  'app.certificate.form.isCA': 'Allow Certificate Signing',
  'app.certificate.form.isCAHelp':
    'Enable this certificate to sign other certificates',

  // Import form
  'app.certificate.import.privateKey': 'Private Key',
  'app.certificate.import.certificate': 'Certificate',
  'app.certificate.import.csr': 'CSR',
  'app.certificate.import.options': 'Options',
  'app.certificate.import.completeChain': 'Auto Complete Certificate Chain',
  'app.certificate.import.keyType': 'Private Key Import Method',
  'app.certificate.import.certificateType': 'Certificate Import Method',
  'app.certificate.import.csrType': 'CSR Import Method',
  'app.certificate.import.fileUpload': 'File Upload',
  'app.certificate.import.textInput': 'Text Input',
  'app.certificate.import.localPath': 'Local Path',
  'app.certificate.import.selectFile': 'Select File',

  'app.certificate.import.keyFile': 'Private Key File',
  'app.certificate.import.keyContent': 'Private Key Content',
  'app.certificate.import.keyPath': 'Private Key Path',
  'app.certificate.import.keyContentPlaceholder':
    'Paste private key content (PEM format)',
  'app.certificate.import.keyPathPlaceholder': 'Enter private key file path',
  'app.certificate.import.keyFileRequired': 'Please select private key file',
  'app.certificate.import.keyContentRequired':
    'Please enter private key content',
  'app.certificate.import.keyPathRequired': 'Please enter private key path',

  'app.certificate.import.certificateFile': 'Certificate File',
  'app.certificate.import.certificateContent': 'Certificate Content',
  'app.certificate.import.certificatePath': 'Certificate Path',
  'app.certificate.import.certificateContentPlaceholder':
    'Paste certificate content (PEM format)',
  'app.certificate.import.certificatePathPlaceholder':
    'Enter certificate file path',
  'app.certificate.import.certificateFileRequired':
    'Please select certificate file',
  'app.certificate.import.certificateContentRequired':
    'Please enter certificate content',
  'app.certificate.import.certificatePathRequired':
    'Please enter certificate path',

  'app.certificate.import.csrFile': 'CSR File',
  'app.certificate.import.csrContent': 'CSR Content',
  'app.certificate.import.csrPath': 'CSR Path',
  'app.certificate.import.csrContentPlaceholder':
    'Paste CSR content (PEM format)',
  'app.certificate.import.csrPathPlaceholder': 'Enter CSR file path',

  // Certificate detail
  'app.certificate.basicInfo': 'Basic Information',
  'app.certificate.certificateConfig': 'Certificate Configuration',
  'app.certificate.organizationInfo': 'Organization Information',
  'app.certificate.locationInfo': 'Location Information',
  'app.certificate.subjectInfo': 'Subject Information',
  'app.certificate.country': 'Country',
  'app.certificate.organization': 'Organization',
  'app.certificate.issuerInfo': 'Issuer Information',
  'app.certificate.notBefore': 'Valid From',
  'app.certificate.notAfter': 'Valid Until',
  'app.certificate.keyAlgorithm': 'Key Algorithm',
  'app.certificate.keySize': 'Key Size',
  'app.certificate.isCA': 'Is CA Certificate',
  'app.certificate.source': 'File Path',
  'app.certificate.viewFile': 'View File',
  'app.certificate.copySourcePath': 'Copy Certificate Path',
  'app.certificate.issuerCN': 'Issuer CN',
  'app.certificate.issuerCountry': 'Issuer Country',
  'app.certificate.issuerOrganization': 'Issuer Organization',
  'app.certificate.altNames': 'Alternative Names',
  'app.certificate.pemContent': 'Certificate Content (PEM Format)',
  'app.certificate.copyPEM': 'Copy PEM',
  'app.certificate.privateKeyContent': 'Private Key Content (PEM Format)',
  'app.certificate.copyPrivateKey': 'Copy Private Key',
  'app.certificate.copyPrivateKeyPath': 'Copy Private Key Path',
  'app.certificate.csrContent': 'CSR Content (PEM Format)',
  'app.certificate.copyCSR': 'Copy CSR',
  'app.certificate.csrUnavailable': 'No CSR content available',
  'app.certificate.commonName': 'Common Name',
  'app.certificate.emailAddresses': 'Email Addresses',

  // Messages
  'app.certificate.createSuccess': 'Certificate group created successfully',
  'app.certificate.createError': 'Failed to create certificate group',
  'app.certificate.importSuccess': 'Certificate imported successfully',
  'app.certificate.importError': 'Failed to import certificate',
  'app.certificate.updateSuccess': 'Certificate updated successfully',
  'app.certificate.updateError': 'Failed to update certificate',
  'app.certificate.generateSuccess':
    'Self-signed certificate generated successfully',
  'app.certificate.generateError': 'Failed to generate self-signed certificate',
  'app.certificate.completeSuccess': 'Certificate chain completed successfully',
  'app.certificate.completeError': 'Failed to complete certificate chain',
  'app.certificate.deleteSuccess': 'Deleted successfully',
  'app.certificate.deleteError': 'Failed to delete',
  'app.certificate.copySuccess': 'Copied successfully',
  'app.certificate.copyError': 'Failed to copy',
  'app.certificate.pathUnavailable': 'No path available to copy',
  'app.certificate.loadError': 'Failed to load certificate list',
  'app.certificate.loadDetailError': 'Failed to load certificate details',
  'app.certificate.loadPrivateKeyError':
    'Failed to load private key information',
  'app.certificate.loadCSRError': 'Failed to load CSR information',

  // Confirm dialogs
  'app.certificate.deleteGroupConfirm.title':
    'Confirm Delete Certificate Group',
  'app.certificate.deleteGroupConfirm.content':
    'Are you sure you want to delete certificate group "{alias}"? This will delete all certificates and private keys in this group and cannot be undone.',
  'app.certificate.deleteCertificateConfirm.title':
    'Confirm Delete Certificate',
  'app.certificate.deleteCertificateConfirm.content':
    'Are you sure you want to delete this certificate? This action cannot be undone.',

  // Error messages
  'app.certificate.error.noHost': 'Please select a host first',
};
