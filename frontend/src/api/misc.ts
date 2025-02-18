import request from '@/helper/api-helper';

export interface IpOption {
  label: string;
  value: string;
}

export function getAvailableIpsApi() {
  return request.get<IpOption[]>('/misc/ip');
}
