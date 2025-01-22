import request from '@/helper/api-helper';
import { ApiListResult } from '@/types/global';

export function getTerminalSessionsApi(host: number) {
  return request.get<
    ApiListResult<{
      name: string;
      session: string;
      status: string;
      time: string;
    }>
  >(`/terminals/${host}/sessions`);
}

export function renameTerminalSessionApi(
  host: number,
  params: {
    session: string;
    data: string;
  }
) {
  return request.post(`/terminals/${host}/sessions/rename`, params);
}

export function detachTerminalSessionApi(
  host: number,
  params: {
    session: string;
  }
) {
  return request.post(`/terminals/${host}/sessions/detach`, params);
}

export function quitTerminalSessionApi(
  host: number,
  params: {
    session: string;
  }
) {
  return request.post(`/terminals/${host}/sessions/quit`, params);
}

export function pruneTerminalSessionApi(host: number) {
  return request.post(`/terminals/${host}/sessions/prune`);
}
