import request from '@/helper/api-helper';
// todo: 缺少左上角内存+CPU+网络的轮询接口

// todo: 缺少虚拟内存
// todo: 缺少空闲时间后面的空闲比例
// todo: 1/5/15分钟进程数不是百分比
export interface SysInfoOverviewRes {
  boot_time: string;
  cpu_usage: string;
  current_load: {
    process_count1: number;
    process_count15: number;
    process_count5: number;
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
  // 需要返回秒，涉及多语言
  idle_time: number;
  // todo: 需要返回秒
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
    retryTimes: number;
    servers: [string];
    timeout: number;
  };
  networks: [
    {
      address: {
        gate: string;
        ip: string;
        mask: string;
        type: string;
      };
      mac: string;
      name: string;
      proto: string;
      status: string;
      traffic: {
        rx: string;
        rx_bytes: number;
        rx_speed: string;
        tx: string;
        tx_bytes: number;
        tx_speed: string;
      };
    }
  ];
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
  return request.get('sysinfo/{host}/system');
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
