import request from '@/helper/api-helper';

export function getTerminalSessionsApi(host: number) {
  return request.get(`/terminals/${host}/sessions`);
}

export function renameTerminalSessionApi(
  host: number,
  params: {
    session: string;
    data: string;
  }
) {
  return request.post(`/terminals/${host}/session/rename`, params);
}

export function detachTerminalSessionApi(
  host: number,
  params: {
    session: string;
  }
) {
  return request.post(`/terminals/${host}/session/detach`, params);
}

export function quitTerminalSessionApi(
  host: number,
  params: {
    session: string;
  }
) {
  return request.post(`/terminals/${host}/session/quit`, params);
}
