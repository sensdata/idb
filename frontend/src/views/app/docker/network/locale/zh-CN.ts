export default {
  'app.docker.network.list.action.create': '新建网络',
  'app.docker.network.list.column.name': '网络名称',
  'app.docker.network.list.column.driver': '驱动类型',
  'app.docker.network.list.column.subnet': '子网',
  'app.docker.network.list.column.gateway': '网关',
  'app.docker.network.list.operation.inspect': '详情',
  'app.docker.network.list.operation.delete': '删除',
  'app.docker.network.list.operation.delete.confirm':
    '确定要删除该网络吗？若仍有容器连接该网络，删除会失败。',
  'app.docker.network.list.operation.delete.success': '删除成功',
  'app.docker.network.list.operation.delete.failed': '删除失败',
  'app.docker.network.create.title': '新建网络',
  'app.docker.network.create.guide.title': '创建网络说明',
  'app.docker.network.create.guide.desc':
    'bridge 适合容器互联，macvlan 适合接入物理网络，host 直接使用宿主机网络。',
  'app.docker.network.create.guide.host':
    'Host 驱动将直接复用宿主机网络，通常无需配置子网/网关/IP范围。',
  'app.docker.network.create.form.name': '网络名称',
  'app.docker.network.create.form.name.placeholder': '请输入网络名称',
  'app.docker.network.create.form.name.required': '请输入网络名称',
  'app.docker.network.create.form.driver': '驱动类型',
  'app.docker.network.create.form.driver.placeholder': '请选择驱动类型',
  'app.docker.network.create.form.driver.required': '请选择驱动类型',
  'app.docker.network.create.form.subnet': '子网',
  'app.docker.network.create.form.subnet.placeholder': '如 172.18.0.0/16',
  'app.docker.network.create.form.subnet.invalid':
    '请输入有效的 IPv4 CIDR 子网',
  'app.docker.network.create.form.gateway': '网关',
  'app.docker.network.create.form.gateway.placeholder': '如 172.18.0.1',
  'app.docker.network.create.form.gateway.invalid':
    '请输入有效的 IPv4 网关地址',
  'app.docker.network.create.form.ip_range': 'IP范围',
  'app.docker.network.create.form.ip_range.placeholder':
    '如 172.18.0.10-172.18.0.100',
  'app.docker.network.create.form.ip_range.invalid': '请输入有效的 IPv4 IP范围',
  'app.docker.network.create.form.exclude_ip': '排除IP(IPv4)',
  'app.docker.network.create.form.exclude_ip_v6': '排除IP(IPv6)',
  'app.docker.network.create.form.exclude_ip.add': '添加排除IP',
  'app.docker.network.create.form.exclude_ip.label': '标签',
  'app.docker.network.create.form.exclude_ip.ip': 'IP地址',
  'app.docker.network.create.form.subnet_v6': '子网(IPv6)',
  'app.docker.network.create.form.subnet_v6.placeholder': '如 fd00::/64',
  'app.docker.network.create.form.subnet_v6.invalid':
    '请输入有效的 IPv6 CIDR 子网',
  'app.docker.network.create.form.gateway_v6': '网关(IPv6)',
  'app.docker.network.create.form.gateway_v6.placeholder': '如 fd00::1',
  'app.docker.network.create.form.gateway_v6.invalid':
    '请输入有效的 IPv6 网关地址',
  'app.docker.network.create.form.ip_range_v6': 'IP范围(IPv6)',
  'app.docker.network.create.form.ip_range_v6.placeholder':
    '如 fd00::10-fd00::100',
  'app.docker.network.create.form.ip_range_v6.invalid':
    '请输入有效的 IPv6 IP范围',
  'app.docker.network.create.form.options': '参数',
  'app.docker.network.create.form.options.placeholder':
    '每行一个参数，如 key=value',
  'app.docker.network.create.form.labels': '标签',
  'app.docker.network.create.form.labels.placeholder':
    '每行一个标签，如 key=value',
  'app.docker.network.inspect.title': '网络详情',
  'app.docker.network.inspect.tab.friendly': '结构化视图',
  'app.docker.network.inspect.tab.raw': '原文(JSON)',
  'app.docker.network.inspect.section.basic': '基础信息',
  'app.docker.network.inspect.section.ipam': 'IPAM配置',
  'app.docker.network.inspect.section.options': '网络参数',
  'app.docker.network.inspect.section.labels': '标签',
  'app.docker.network.inspect.section.containers': '已连接容器',
  'app.docker.network.inspect.field.scope': '作用域',
  'app.docker.network.inspect.field.ipv6': 'IPv6启用',
  'app.docker.network.create.success': '网络创建成功',
  'app.docker.network.create.failed': '网络创建失败',
};
