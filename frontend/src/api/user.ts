import request from '@/helper/api-helper';
import { UserState } from '@/store/modules/user/types';

export interface LoginDataDo {
  name: string;
  password: string;
}

export interface LoginRes {
  token: string;
  name: string;
}
export function loginApi(data: LoginDataDo) {
  return request.post<LoginRes>('auth/sessions', data);
}

export function logoutApi() {
  return request.post<LoginRes>('user/logout');
}

export function getUserInfoApi() {
  return request.get<UserState>('users/profile');
}

export function changePasswordApi(data: {
  old_password: string;
  new_password: string;
}) {
  return request.put('users/password', data);
}
