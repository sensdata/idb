import request from '@/helper/api-helper';

/**
 * 获取SSH配置
 * @param hostId - 主机ID
 */
export function getSSHConfig(hostId: number) {
  return request.get(`/ssh/${hostId}/config`);
}

/**
 * 获取SSH配置文件内容
 * @param hostId - 主机ID
 */
export function getSSHConfigContent(hostId: number) {
  return request.get(`/ssh/${hostId}/config/content`);
}

/**
 * 更新SSH配置
 * @param hostId - 主机ID
 * @param values - 要更新的配置键值对列表
 */
export function updateSSHConfig(
  hostId: number,
  values: Array<{ key: string; old_value: string; new_value: string }>
) {
  return request.put(`/ssh/${hostId}/config`, { values });
}

/**
 * 更新SSH配置文件内容
 * @param hostId - 主机ID
 * @param content - 配置文件内容
 */
export function updateSSHConfigContent(hostId: number, content: string) {
  return request.put(`/ssh/${hostId}/config/content`, { content });
}

/**
 * 操作SSH服务
 * @param hostId - 主机ID
 * @param operation - 操作类型: 'enable' | 'disable' | 'stop' | 'reload' | 'restart'
 */
export function operateSSH(
  hostId: number,
  operation: 'enable' | 'disable' | 'stop' | 'reload' | 'restart'
) {
  return request.post(`/ssh/${hostId}/operate`, { operation });
}
