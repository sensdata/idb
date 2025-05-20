import request from '@/helper/api-helper';

/**
 * 获取SSH配置
 * @param hostId - 主机ID
 */
export function getSSHConfig(hostId: number) {
  return request.get(`/ssh/${hostId}/config`);
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
