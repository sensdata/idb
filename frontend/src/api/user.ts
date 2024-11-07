import request from '@/helper/api-helper';
import { UserState } from '@/store/modules/user/types';

export interface LoginData {
  name: string;
  password: string;
}

export interface LoginRes {
  token: string;
  name: string;
}
export function login(data: LoginData) {
  return {
    token:
      'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiTmFtZSI6ImFkbWluIiwiQnVmZmVyVGltZSI6MzYwMCwiaXNzIjoiaWRiIiwiZXhwIjoxNzMwOTAxODU0fQ.s5r9tuJi3KIjM-rklhuG5fW-oX8fLxx0Ipex6rVXKwc',
    name: 'admin',
    role: 'admin',
  };
  // return request.post<LoginRes>('auth/sessions', data);
}

export function logout() {
  return request.post<LoginRes>('user/logout');
}

export function getUserInfo() {
  return {
    name: 'admin',
    role: 'admin',
  };
  // return request.post<UserState>('user/info');
}
