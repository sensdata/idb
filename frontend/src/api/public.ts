import request from '@/helper/api-helper';

export function getPublicVersionApi() {
  return request.get<{
    version: string;
  }>('/public/version');
}
