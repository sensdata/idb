import request from '@/helper/api-helper';

// Certificate related interfaces
export interface PrivateKeyInfo {
  alias: string;
  key_algorithm: string;
  key_size: number;
  pem: string;
}

export interface CSRInfo {
  common_name: string;
  country: string;
  organization: string;
  email_addresses: string[];
  pem: string;
}

export interface CertificateSimpleInfo {
  domain: string;
  alt_names: string[];
  not_before: string;
  not_after: string;
  issuer_organization: string;
  status: string;
  source: string;
}

export interface CertificateGroup {
  alias: string;
  requester: string;
  certificates: CertificateSimpleInfo[];
}

export interface CertificateInfo {
  // 证书信息
  domain: string;
  alt_names: string[];
  not_before: string;
  not_after: string;
  country: string;
  organization: string;
  key_algorithm: string;
  key_size: number;
  is_ca: boolean;

  // 签发机构信息
  issuer_cn: string;
  issuer_country: string;
  issuer_organization: string;

  // 证书代码（PEM格式）
  pem: string;
  source: string;
}

// Request interfaces
export interface CreateGroupRequest {
  alias: string;
  domain_name: string;
  email: string;
  organization: string;
  organization_unit: string;
  country: string;
  province: string;
  city: string;
  key_algorithm: string;
}

export interface SelfSignedRequest {
  alias: string;
  expire_unit: 'day' | 'year';
  expire_value: number;
  alt_domains: string;
  alt_ips: string;
  is_ca: boolean;
}

export interface ImportCertificateRequest {
  alias: string;
  key_type: number;
  key_content?: string;
  key_path?: string;
  ca_type: number;
  ca_content?: string;
  ca_path?: string;
  csr_type: number;
  csr_content?: string;
  csr_path?: string;
}

export interface GroupPkRequest {
  alias: string;
}

export interface CertificateInfoRequest {
  source: string;
}

export interface DeleteGroupRequest {
  alias: string;
}

export interface DeleteCertificateRequest {
  source: string;
}

// API functions
/**
 * 获取证书组列表
 * @param hostId - 主机ID
 */
export function getCertificateGroups(hostId: number) {
  return request.get<{ items: CertificateGroup[] }>(
    `/certificates/${hostId}/group`
  );
}

/**
 * 创建证书组
 * @param hostId - 主机ID
 * @param data - 证书组创建参数
 */
export function createCertificateGroup(
  hostId: number,
  data: CreateGroupRequest
) {
  return request.post(`/certificates/${hostId}/group`, data);
}

/**
 * 删除证书组
 * @param hostId - 主机ID
 * @param data - 删除参数
 */
export function deleteCertificateGroup(
  hostId: number,
  data: DeleteGroupRequest
) {
  return request.delete(`/certificates/${hostId}/group`, data);
}

/**
 * 获取私钥信息
 * @param hostId - 主机ID
 * @param alias - 证书组别名
 */
export function getPrivateKeyInfo(hostId: number, alias: string) {
  return request.get<PrivateKeyInfo>(`/certificates/${hostId}/group/key`, {
    alias,
  });
}

/**
 * 获取CSR信息
 * @param hostId - 主机ID
 * @param alias - 证书组别名
 */
export function getCSRInfo(hostId: number, alias: string) {
  return request.get<CSRInfo>(`/certificates/${hostId}/group/csr`, { alias });
}

/**
 * 获取证书详细信息
 * @param hostId - 主机ID
 * @param source - 证书文件路径
 */
export function getCertificateInfo(hostId: number, source: string) {
  return request.get<CertificateInfo>(`/certificates/${hostId}`, { source });
}

/**
 * 删除证书
 * @param hostId - 主机ID
 * @param data - 删除参数
 */
export function deleteCertificate(
  hostId: number,
  data: DeleteCertificateRequest
) {
  return request.delete(`/certificates/${hostId}`, data);
}

/**
 * 生成自签名证书
 * @param hostId - 主机ID
 * @param data - 自签名证书参数
 */
export function generateSelfSignedCertificate(
  hostId: number,
  data: SelfSignedRequest
) {
  return request.post(`/certificates/${hostId}/sign/self`, data);
}

/**
 * 补齐证书链
 * @param hostId - 主机ID
 * @param data - 证书信息参数
 */
export function completeCertificateChain(
  hostId: number,
  data: CertificateInfoRequest
) {
  return request.post(`/certificates/${hostId}/complete`, data);
}

/**
 * 导入证书
 * @param hostId - 主机ID
 * @param formData - 表单数据
 */
export function importCertificate(hostId: number, formData: FormData) {
  return request.post(`/certificates/${hostId}/import`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}
