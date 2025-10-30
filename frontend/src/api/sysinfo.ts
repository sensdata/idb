import request from '@/helper/api-helper';

// todo: 缺少空闲时间后面的空闲比例
export interface SysInfoOverviewRes {
  boot_time: string;
  cpu_usage: string;
  current_load: {
    process_count1: string;
    process_count5: string;
    process_count15: string;
  };
  memory_usage: {
    buffered: string;
    cached: string;
    free: string;
    free_rate: number;
    kernel: string;
    physical: string;
    real_used: string;
    total: string;
    used: string;
    used_rate: number;
  };
  idle_time: number;
  idle_rate: number;
  run_time: number;
  server_time: string;
  server_time_zone: string;
  storage: Array<{
    free: string;
    name: string;
    total: string;
    used: string;
    used_rate: number;
  }>;
  swap_usage: {
    free: string;
    free_rate: number;
    total: string;
    used: string;
    used_rate: number;
  };
}
export function getSysInfoOverviewtApi() {
  return request.get<SysInfoOverviewRes>('sysinfo/{host}/overview');
}

export interface SysInfoNetworkRes {
  dns: {
    retry: number;
    servers: string[];
    timeout: number;
  };
  networks: Array<{
    address: Array<{
      gate: string | null;
      ip: string;
      mask: string | null;
      type: string; // e.g., IPv4/IPv6
    }>;
    mac: string;
    name: string;
    proto: string; // dhcp | static
    status: string; // up | down
    traffic: {
      rx: string;
      rx_bytes: number;
      rx_speed: string;
      tx: string;
      tx_bytes: number;
      tx_speed: string;
    };
  }>;
}
export function getSysInfoNetworkApi() {
  return request.get<SysInfoNetworkRes>('sysinfo/{host}/network');
}

export interface SysInfoSystemRes {
  host_name: string;
  kernel: string;
  platform: string;
  version: string;
  vertual: string;
}
export function getSysInfoSystemApi() {
  return request.get<SysInfoSystemRes>('sysinfo/{host}/system');
}

export interface SysInfoHardwareRes {
  cpu_count: number;
  cpu_cores: number;
  processor: number;
  module_names: string[];
  memory: string;
}
export function getSysInfoHardwareApi() {
  return request.get<SysInfoHardwareRes>('sysinfo/{host}/hardware');
}

export function syncTimeApi() {
  return request.post('sysinfo/{host}/action/sync/time');
}

export interface UpdateTimeParams {
  timestamp: number;
}

export function updateTimeApi(data: UpdateTimeParams) {
  return request.post('sysinfo/{host}/action/upd/time', data);
}

export function updateTimeZoneApi(data: { timezone: string }) {
  return request.post('sysinfo/{host}/action/upd/timezone', data);
}

export interface UpdateHostNameParams {
  host_name: string;
}

export function updateHostNameApi(data: UpdateHostNameParams) {
  return request.post('sysinfo/{host}/action/upd/hostname', data);
}

export interface UpdateDNSParams {
  retry: number;
  servers: string[];
  timeout: number;
}
export function updateDNSApi(data: UpdateDNSParams) {
  return request.post('sysinfo/{host}/action/upd/dns', data);
}

export function clearMemoryCacheApi() {
  return request.post('sysinfo/{host}/action/memcache/clear');
}

export function getAutoClearMemoryCacheApi() {
  return request.get('sysinfo/{host}/action/memcache/auto/set');
}

export function setAutoClearMemoryCacheApi(data: { interval: number }) {
  return request.post('sysinfo/{host}/action/memcache/auto/set', data);
}

export interface CreateSwapParams {
  size: number;
}
export function createSwapApi(data: CreateSwapParams) {
  return request.post('sysinfo/{host}/action/swap/create', data);
}

export function deleteSwapApi() {
  return request.post('sysinfo/{host}/action/swap/delete');
}

export interface SysInfoSettingsRes {
  max_open_files: number;
  max_watch_files: number;
}
export function getSysInfoSettingsApi() {
  return request.get<SysInfoSettingsRes>('sysinfo/{host}/settings');
}

export interface UpdateSettingsParams {
  max_open_files: number;
  max_watch_files: number;
}

export function updateSysInfoSettingsApi(data: UpdateSettingsParams) {
  return request.post('sysinfo/{host}/action/upd/settings', data);
}
