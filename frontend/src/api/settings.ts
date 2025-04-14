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
  return request.post<{
    redirect_url: string;
  }>('/settings', data);
}

export function getAvailableIpsApi() {
  return request.get<{
    ips: Array<{
      ip: string;
      name: string;
    }>;
  }>('/settings/ips');
}

export interface TimezoneOption {
  id: number;
  created_at: string;
  updated_at: string;
  value: string;
  abbr: string;
  offset: number;
  isdst: boolean;
  text: string;
  utc: string;
}

export function getTimezonesApi(params: { page: number; page_size: number }) {
  return request.get<{
    items: TimezoneOption[];
    total: number;
  }>('/settings/timezones', params);
}
