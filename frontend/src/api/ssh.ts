import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

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

/**
 * 获取SSH密钥列表
 * @param hostId - 主机ID
 * @param params - 分页参数
 */
export function getSSHKeys(hostId: number, params: ApiListParams) {
  return request.get<ApiListResult<any>>(`/ssh/${hostId}/keys`, { params });
}

/**
 * 生成SSH密钥
 * @param hostId - 主机ID
 * @param data - 密钥生成参数
 */
export function generateSSHKey(
  hostId: number,
  data: {
    key_name: string;
    encryption_mode: 'rsa' | 'ed25519' | 'ecdsa' | 'dsa';
    key_bits: 1024 | 2048;
    password?: string;
    enable?: boolean;
  }
) {
  return request.post(`/ssh/${hostId}/keys`, data);
}

/**
 * 下载SSH私钥
 * @param hostId - 主机ID
 * @param source - 密钥路径
 */
export function downloadSSHKey(hostId: number, source: string) {
  return request.get(
    `/ssh/${hostId}/keys/download`,
    { source },
    { responseType: 'blob' }
  );
}

/**
 * 启用或禁用SSH密钥
 * @param hostId - 主机ID
 * @param data - 启用/禁用参数
 */
export function toggleSSHKeyEnabled(
  hostId: number,
  data: {
    key_name: string;
    enable: boolean;
  }
) {
  return request.post(`/ssh/${hostId}/keys/enable`, data);
}

/**
 * 设置SSH密钥密码
 * @param hostId - 主机ID
 * @param data - 密码设置参数
 */
export function setSSHKeyPassword(
  hostId: number,
  data: {
    key_name: string;
    password: string;
  }
) {
  return request.post(`/ssh/${hostId}/keys/password/set`, data);
}

/**
 * 更新SSH密钥密码
 * @param hostId - 主机ID
 * @param data - 密码更新参数
 */
export function updateSSHKeyPassword(
  hostId: number,
  data: {
    key_name: string;
    old_password: string;
    new_password: string;
  }
) {
  return request.post(`/ssh/${hostId}/keys/password/update`, data);
}

/**
 * 清除SSH密钥密码
 * @param hostId - 主机ID
 * @param data - 清除密码参数
 */
export function clearSSHKeyPassword(
  hostId: number,
  data: {
    key_name: string;
  }
) {
  return request.post(`/ssh/${hostId}/keys/password/clear`, data);
}

/**
 * 删除SSH密钥
 * @param hostId - 主机ID
 * @param keyName - 密钥名称
 * @param onlyPrivateKey - 是否只删除私钥
 */
export function deleteSSHKey(
  hostId: number,
  keyName: string,
  onlyPrivateKey = false
) {
  return request.delete(`/ssh/${hostId}/keys`, {
    key_name: keyName,
    only_private_key: onlyPrivateKey,
  });
}
