export default {
  'app.docker.network.list.action.create': 'Create Network',
  'app.docker.network.list.column.name': 'Network Name',
  'app.docker.network.list.column.driver': 'Driver',
  'app.docker.network.list.column.subnet': 'Subnet',
  'app.docker.network.list.column.gateway': 'Gateway',
  'app.docker.network.list.operation.inspect': 'Inspect',
  'app.docker.network.list.operation.delete': 'Delete',
  'app.docker.network.list.operation.delete.confirm':
    'Delete this network? The operation will fail if containers are still attached.',
  'app.docker.network.list.operation.delete.success':
    'Network deleted successfully',
  'app.docker.network.list.operation.delete.failed': 'Delete failed',
  'app.docker.network.create.title': 'Create Network',
  'app.docker.network.create.guide.title': 'Network creation tips',
  'app.docker.network.create.guide.desc':
    'bridge is for container interconnection, macvlan for physical network access, and host reuses the host network stack.',
  'app.docker.network.create.guide.host':
    'With host driver, subnet/gateway/IP range are usually unnecessary.',
  'app.docker.network.create.form.name': 'Network Name',
  'app.docker.network.create.form.name.placeholder':
    'Please enter network name',
  'app.docker.network.create.form.name.required': 'Please enter network name',
  'app.docker.network.create.form.driver': 'Driver',
  'app.docker.network.create.form.driver.placeholder': 'Please select driver',
  'app.docker.network.create.form.driver.required': 'Please select driver',
  'app.docker.network.create.form.subnet': 'Subnet',
  'app.docker.network.create.form.subnet.placeholder': 'e.g. 172.18.0.0/16',
  'app.docker.network.create.form.subnet.invalid':
    'Please enter a valid IPv4 CIDR subnet',
  'app.docker.network.create.form.gateway': 'Gateway',
  'app.docker.network.create.form.gateway.placeholder': 'e.g. 172.18.0.1',
  'app.docker.network.create.form.gateway.invalid':
    'Please enter a valid IPv4 gateway',
  'app.docker.network.create.form.ip_range': 'IP Range',
  'app.docker.network.create.form.ip_range.placeholder':
    'e.g. 172.18.0.10-172.18.0.100',
  'app.docker.network.create.form.ip_range.invalid':
    'Please enter a valid IPv4 IP range',
  'app.docker.network.create.form.exclude_ip': 'Exclude IP (IPv4)',
  'app.docker.network.create.form.exclude_ip_v6': 'Exclude IP (IPv6)',
  'app.docker.network.create.form.exclude_ip.add': 'Add Exclude IP',
  'app.docker.network.create.form.exclude_ip.label': 'Label',
  'app.docker.network.create.form.exclude_ip.ip': 'IP Address',
  'app.docker.network.create.form.subnet_v6': 'Subnet (IPv6)',
  'app.docker.network.create.form.subnet_v6.placeholder': 'e.g. fd00::/64',
  'app.docker.network.create.form.subnet_v6.invalid':
    'Please enter a valid IPv6 CIDR subnet',
  'app.docker.network.create.form.gateway_v6': 'Gateway (IPv6)',
  'app.docker.network.create.form.gateway_v6.placeholder': 'e.g. fd00::1',
  'app.docker.network.create.form.gateway_v6.invalid':
    'Please enter a valid IPv6 gateway',
  'app.docker.network.create.form.ip_range_v6': 'IP Range (IPv6)',
  'app.docker.network.create.form.ip_range_v6.placeholder':
    'e.g. fd00::10-fd00::100',
  'app.docker.network.create.form.ip_range_v6.invalid':
    'Please enter a valid IPv6 IP range',
  'app.docker.network.create.form.options': 'Options',
  'app.docker.network.create.form.options.placeholder':
    'One per line, e.g. key=value',
  'app.docker.network.create.form.labels': 'Labels',
  'app.docker.network.create.form.labels.placeholder':
    'One per line, e.g. key=value',
  'app.docker.network.inspect.title': 'Network Details',
  'app.docker.network.inspect.tab.friendly': 'Structured View',
  'app.docker.network.inspect.tab.raw': 'Raw (JSON)',
  'app.docker.network.inspect.section.basic': 'Basic Information',
  'app.docker.network.inspect.section.ipam': 'IPAM Configuration',
  'app.docker.network.inspect.section.options': 'Network Options',
  'app.docker.network.inspect.section.labels': 'Labels',
  'app.docker.network.inspect.section.containers': 'Connected Containers',
  'app.docker.network.inspect.field.scope': 'Scope',
  'app.docker.network.inspect.field.ipv6': 'IPv6 Enabled',
  'app.docker.network.create.success': 'Network created successfully',
  'app.docker.network.create.failed': 'Network creation failed',
};
