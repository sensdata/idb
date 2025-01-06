import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getProcessListApi(params: ApiListParams) {
  // return request.get('process/{host}', params);
  return new Promise<ApiListResult<any>>((resolve) => {
    setTimeout(() => {
      resolve({
        page: 1,
        page_size: 20,
        items: [
          {
            name: 'java',
            pid: 123,
            ppid: 1,
            status: 'running',
            cpu: 0.1,
            memory: 0.1,
            user: 'root',
            threads: 1,
            startTime: '2021-01-01 00:00:00',
          },
          {
            name: 'node',
            pid: 124,
            ppid: 1,
            status: 'running',
            cpu: 0.1,
            memory: 0.1,
            user: 'root',
            threads: 1,
            startTime: '2021-01-01 00:00:00',
          },
        ],
        total: 2,
      });
    }, 1000);
  });
}

export function getProcessDetailApi(params: { pid: string }) {
  // return request.get('process/{host}/detail', params);
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        name: 'redis',
        status: 'idle',
        pid: '10000',
        pppid: '30000',
        threads: '2',
        connections: '4',
        disk_read: '300kb',
        disk_write: '200kb',
        user: 'root',
        start_time: '2021-09-01 12:00:00',
        start_command: 'node index.js',
        memory: {
          rss: '100mb',
          swap: '200mb',
          vms: '300mb',
          hwm: '400mb',
          data: '500mb',
          stack: '600mb',
          locked: '700mb',
        },
        fs: [
          {
            file: '/usr/local/bin/redis',
            fd: '1',
          },
          {
            file: '/usr/local/bin/redis',
            fd: '2',
          },
        ],
        env: 'NODE_ENV=production\nPORT=3000',
        network: [
          {
            localAddres: '127.0.0.1',
            localPort: '6379',
            remoteAddress: '23.31.31.31',
            remotePort: '80',
            state: 'ESTABLISHED',
          },
        ],
      });
    }, 1000);
  });
}

export function killProcessApi(params: { pid: string }) {
  return request.delete('process/{host}/kill', params);
}
