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

export interface SettingsStatusChecks {
  database_connected: boolean;
  default_host_found: boolean;
  default_agent_online: boolean;
  default_agent_fresh: boolean;
}

export interface CenterRuntimeStatus {
  pid: number;
  version: string;
  bind_ip: string;
  bind_port: number;
  bind_domain: string;
  https_enabled: boolean;
  access_url: string;
  started_at: number;
  uptime: number;
  cpu_percent: number;
  cpu_seconds: number;
  memory_rss: number;
  heap_alloc: number;
  heap_sys: number;
  stack_inuse: number;
  goroutines: number;
  open_fds: number;
}

export interface AgentRuntimeStatus {
  host_id: number;
  host_name: string;
  host_addr: string;
  agent_addr: string;
  agent_port: number;
  agent_version: string;
  installed: string;
  connected: string;
  last_heartbeat: number;
  cpu: number;
  memory: number;
  mem_total: string;
  mem_used: string;
  disk: number;
  boot_time: string;
  run_time: number;
}

export interface SettingsStatus {
  collected_at: number;
  center: CenterRuntimeStatus;
  agent?: AgentRuntimeStatus;
  checks: SettingsStatusChecks;
}

export function getSettingsStatusApi() {
  return request.get<SettingsStatus>('/settings/status');
}

export function updateSettingsApi(data: SettingsForm) {
  return request.post<{
    redirect_url: string;
  }>('/settings', data);
}

export function getSettingsAboutApi() {
  return request.get<{
    version: string;
    new_version: string;
  }>('/settings/about');
}

export function upgradeApi() {
  return request.post('/settings/upgrade');
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
