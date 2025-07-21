export default {
  'app.certificate.title': '证书管理',
  'app.certificate.createGroup': '创建证书组',
  'app.certificate.import': '导入证书',
  'app.certificate.updateCertificate': '更新证书',
  'app.certificate.generateSelfSigned': '生成自签名证书',
  'app.certificate.viewDetail': '查看详情',
  'app.certificate.viewDetails': '查看详情',
  'app.certificate.completeChain': '补齐证书链',
  'app.certificate.deleteGroup': '删除证书组',
  'app.certificate.viewPrivateKey': '查看私钥',
  'app.certificate.viewCSR': '查看CSR',
  'app.certificate.certificateDetail': '证书详情',
  'app.certificate.privateKeyInfo': '私钥信息',
  'app.certificate.csrInfo': 'CSR信息',

  // Table headers
  'app.certificate.alias': '别名',
  'app.certificate.requester': '申请者',
  'app.certificate.certificates': '证书',
  'app.certificate.domain': '域名',
  'app.certificate.status': '状态',
  'app.certificate.expiresOn': '过期时间',
  'app.certificate.noCertificates': '暂无证书',
  'app.certificate.certificatesCount': '个证书',

  // Status
  'app.certificate.status.valid': '有效',
  'app.certificate.status.expired': '已过期',
  'app.certificate.status.expiringSoon': '即将过期 ({days}天)',
  'app.certificate.status.notYetValid': '尚未生效',
  'app.certificate.status.unknown': '未知',

  // Form fields
  'app.certificate.form.alias': '别名',
  'app.certificate.form.aliasPlaceholder': '请输入证书组别名',
  'app.certificate.form.aliasRequired': '请输入别名',
  'app.certificate.form.aliasFormat': '别名只能包含字母、数字、下划线和连字符',

  'app.certificate.form.domainName': '域名',
  'app.certificate.form.domainNamePlaceholder': '请输入域名',
  'app.certificate.form.domainNameRequired': '请输入域名',

  'app.certificate.form.email': '邮箱',
  'app.certificate.form.emailPlaceholder': '请输入邮箱地址',
  'app.certificate.form.emailRequired': '请输入邮箱地址',
  'app.certificate.form.emailFormat': '请输入有效的邮箱地址',

  'app.certificate.form.organization': '组织',
  'app.certificate.form.organizationPlaceholder': '请输入组织名称',
  'app.certificate.form.organizationRequired': '请输入组织名称',

  'app.certificate.form.organizationUnit': '部门',
  'app.certificate.form.organizationUnitPlaceholder': '请输入部门名称',

  'app.certificate.form.country': '国家',
  'app.certificate.form.countryPlaceholder': '请输入国家代码（如：CN）',
  'app.certificate.form.countryRequired': '请输入国家代码',
  'app.certificate.form.countryFormat': '国家代码必须是2位字符',

  'app.certificate.form.province': '省份',
  'app.certificate.form.provincePlaceholder': '请输入省份',

  'app.certificate.form.city': '城市',
  'app.certificate.form.cityPlaceholder': '请输入城市',

  'app.certificate.form.keyAlgorithm': '密钥算法',
  'app.certificate.form.keyAlgorithmPlaceholder': '请选择密钥算法',
  'app.certificate.form.keyAlgorithmRequired': '请选择密钥算法',

  'app.certificate.form.expireValue': '有效期',
  'app.certificate.form.expireValuePlaceholder': '请输入有效期',
  'app.certificate.form.expireValueRequired': '请输入有效期',
  'app.certificate.form.expireValueRange': '有效期必须在1-9999之间',

  'app.certificate.form.expireUnit': '有效期单位',
  'app.certificate.form.expireUnitPlaceholder': '请选择有效期单位',
  'app.certificate.form.expireUnitRequired': '请选择有效期单位',
  'app.certificate.form.days': '天',
  'app.certificate.form.years': '年',

  'app.certificate.form.altDomains': '备用域名',
  'app.certificate.form.altDomainsPlaceholder':
    '每行一个域名，支持通配符（如：*.example.com）',
  'app.certificate.form.altDomainsHelp': '每行输入一个域名，支持通配符域名',
  'app.certificate.form.altDomainsInvalid': '无效的域名：{domain}',

  'app.certificate.form.altIPs': '备用IP地址',
  'app.certificate.form.altIPsPlaceholder': '每行一个IP地址',
  'app.certificate.form.altIPsHelp': '每行输入一个IP地址，支持IPv4和IPv6',
  'app.certificate.form.altIPsInvalid': '无效的IP地址：{ip}',

  'app.certificate.form.isCA': '允许签发下级证书',
  'app.certificate.form.isCAHelp': '启用后该证书可用于签发其他证书',

  // Import form
  'app.certificate.import.privateKey': '私钥',
  'app.certificate.import.certificate': '证书',
  'app.certificate.import.csr': 'CSR',
  'app.certificate.import.options': '选项',
  'app.certificate.import.completeChain': '自动补齐证书链',
  'app.certificate.import.keyType': '私钥导入方式',
  'app.certificate.import.certificateType': '证书导入方式',
  'app.certificate.import.csrType': 'CSR导入方式',
  'app.certificate.import.fileUpload': '文件上传',
  'app.certificate.import.textInput': '文本输入',
  'app.certificate.import.localPath': '本地路径',
  'app.certificate.import.selectFile': '选择文件',

  'app.certificate.import.keyFile': '私钥文件',
  'app.certificate.import.keyContent': '私钥内容',
  'app.certificate.import.keyPath': '私钥路径',
  'app.certificate.import.keyContentPlaceholder': '请粘贴私钥内容（PEM格式）',
  'app.certificate.import.keyPathPlaceholder': '请输入私钥文件路径',
  'app.certificate.import.keyFileRequired': '请选择私钥文件',
  'app.certificate.import.keyContentRequired': '请输入私钥内容',
  'app.certificate.import.keyPathRequired': '请输入私钥路径',

  'app.certificate.import.certificateFile': '证书文件',
  'app.certificate.import.certificateContent': '证书内容',
  'app.certificate.import.certificatePath': '证书路径',
  'app.certificate.import.certificateContentPlaceholder':
    '请粘贴证书内容（PEM格式）',
  'app.certificate.import.certificatePathPlaceholder': '请输入证书文件路径',
  'app.certificate.import.certificateFileRequired': '请选择证书文件',
  'app.certificate.import.certificateContentRequired': '请输入证书内容',
  'app.certificate.import.certificatePathRequired': '请输入证书路径',

  'app.certificate.import.csrFile': 'CSR文件',
  'app.certificate.import.csrContent': 'CSR内容',
  'app.certificate.import.csrPath': 'CSR路径',
  'app.certificate.import.csrContentPlaceholder': '请粘贴CSR内容（PEM格式）',
  'app.certificate.import.csrPathPlaceholder': '请输入CSR文件路径',

  // Certificate detail
  'app.certificate.basicInfo': '基本信息',
  'app.certificate.certificateConfig': '证书配置',
  'app.certificate.organizationInfo': '组织信息',
  'app.certificate.locationInfo': '地理信息',
  'app.certificate.subjectInfo': '主体信息',
  'app.certificate.issuerInfo': '签发机构信息',
  'app.certificate.notBefore': '生效时间',
  'app.certificate.notAfter': '过期时间',
  'app.certificate.keySize': '密钥长度',
  'app.certificate.isCA': '是否为CA证书',
  'app.certificate.source': '文件路径',
  'app.certificate.issuerCN': '签发者CN',
  'app.certificate.issuerCountry': '签发者国家',
  'app.certificate.issuerOrganization': '签发者组织',
  'app.certificate.altNames': '备用域名',
  'app.certificate.pemContent': '证书内容（PEM格式）',
  'app.certificate.copyPEM': '复制PEM',
  'app.certificate.privateKeyContent': '私钥内容（PEM格式）',
  'app.certificate.copyPrivateKey': '复制私钥',
  'app.certificate.csrContent': 'CSR内容（PEM格式）',
  'app.certificate.copyCSR': '复制CSR',
  'app.certificate.commonName': '通用名称',
  'app.certificate.emailAddresses': '邮箱地址',

  // Messages
  'app.certificate.createSuccess': '证书组创建成功',
  'app.certificate.createError': '证书组创建失败',
  'app.certificate.importSuccess': '证书导入成功',
  'app.certificate.importError': '证书导入失败',
  'app.certificate.updateSuccess': '证书更新成功',
  'app.certificate.updateError': '证书更新失败',
  'app.certificate.generateSuccess': '自签名证书生成成功',
  'app.certificate.generateError': '自签名证书生成失败',
  'app.certificate.completeSuccess': '证书链补齐成功',
  'app.certificate.completeError': '证书链补齐失败',
  'app.certificate.deleteSuccess': '删除成功',
  'app.certificate.deleteError': '删除失败',
  'app.certificate.copySuccess': '复制成功',
  'app.certificate.copyError': '复制失败',
  'app.certificate.loadError': '加载证书列表失败',
  'app.certificate.loadDetailError': '加载证书详情失败',
  'app.certificate.loadPrivateKeyError': '加载私钥信息失败',
  'app.certificate.loadCSRError': '加载CSR信息失败',

  // Confirm dialogs
  'app.certificate.deleteGroupConfirm.title': '确认删除证书组',
  'app.certificate.deleteGroupConfirm.content':
    '确定要删除证书组 "{alias}" 吗？此操作将删除该组下的所有证书和私钥，且无法恢复。',
  'app.certificate.deleteCertificateConfirm.title': '确认删除证书',
  'app.certificate.deleteCertificateConfirm.content':
    '确定要删除此证书吗？此操作无法恢复。',

  // Error messages
  'app.certificate.error.noHost': '请先选择主机',
};
