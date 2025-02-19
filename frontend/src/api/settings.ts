import request from '@/helper/api-helper';

export interface SettingsForm {
  bind_domain: string;
  bind_ip: string;
  bind_port: number;
  https: 'yes' | 'no';
  https_cert_path: string;
  https_cert_type: 'default' | 'custom';
  https_key_path: string;
}

export function getSettingsApi() {
  return request.get<SettingsForm>('/settings');
}

export function updateSettingsApi(data: SettingsForm) {
  return request.post<SettingsForm>('/settings', data);
}

export function getAvailableIpsApi() {
  return request.get<{
    ips: Array<{
      ip: string;
      name: string;
    }>;
  }>('/settings/ips');
}
