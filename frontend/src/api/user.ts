import request from '@/helper/api-helper';
import { UserState } from '@/store/modules/user/types';

export interface LoginDataDo {
  name: string;
  password: string;
}

export interface LoginRes {
  token: string;
  name: string;
  id: number;
}
export function loginApi(data: LoginDataDo) {
  return request.post<LoginRes>('auth/sessions', data);
}

export function logoutApi() {
  return request.delete<LoginRes>('auth/sessions');
}

export function getUserInfoApi() {
  return request.get<UserState>('users/profile');
}

export function changePasswordApi(data: {
  id: number;
  old_password: string;
  password: string;
}) {
  return request.put('users/password', data);
}
