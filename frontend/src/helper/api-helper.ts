import axios from 'axios';
import type { AxiosRequestConfig, AxiosResponse } from 'axios';
// import { Message } from '@arco-design/web-vue';
import { clearToken, getToken } from '@/helper/auth';
import { t } from '@/utils/i18n';

export interface ApiResponse<T = unknown> {
  code: number;
  message?: string;
  data?: T;
}

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';
axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL;

let apiHostId: number | undefined;
export function setApiHostId(hostId?: number) {
  apiHostId = hostId;
}

export function resolveApiUrl(url: string, params?: Record<string, any>) {
  const urlParams = new URLSearchParams();
  Object.entries(params || {}).forEach(([key, value]) => {
    urlParams.set(key, String(value));
  });
  if (url.indexOf('{host}') !== -1) {
    const paramsHost: any = params && (params as any).host;
    let hostId = apiHostId;
    if (typeof paramsHost === 'number') {
      hostId = paramsHost;
    } else if (typeof paramsHost === 'string' && /^\d+$/.test(paramsHost)) {
      hostId = Number(paramsHost);
    }
    if (hostId) {
      url = url.replace('{host}', String(hostId));
    }
  }
  if (url.startsWith('/') && (API_BASE_URL || '').endsWith('/')) {
    url = url.slice(1);
  }
  return API_BASE_URL + url + '?' + urlParams.toString();
}

axios.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // let each request carry token
    // this example using the JWT token
    // Authorization is a custom headers key
    // please modify it according to the actual situation
    const token = getToken();
    // hostId 优先使用 query params 里的 host（仅当是数字），否则回退到全局当前主机 apiHostId
    let hostId = apiHostId;
    const paramsHost: any = config.params && (config.params as any).host;
    if (typeof paramsHost === 'number') {
      hostId = paramsHost;
    } else if (typeof paramsHost === 'string' && /^\d+$/.test(paramsHost)) {
      hostId = Number(paramsHost);
    }

    if (config.url && config.url.indexOf('{host}') !== -1 && hostId) {
      config.url = config.url.replace('{host}', String(hostId));
    }
    if (token) {
      if (!config.headers) {
        config.headers = {};
      }
      config.headers.Authorization = token;
    }
    return config;
  },
  (error) => {
    // do something
    return Promise.reject(error);
  }
);
// add response interceptors
axios.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    if (response.config?.responseType === 'blob') {
      // 如果是文件流，直接过
      return response;
    }

    if (response.data?.code === 200) {
      return response.data.data;
    }

    // 构造错误并保留原始响应，便于组件读取 data
    const msg =
      response.data && response.data.message
        ? String(response.data.message)
        : response.statusText;
    const err: any = new Error(msg);
    err.response = response; // 附带完整响应
    return Promise.reject(err);
  },
  (error) => {
    if (error.response?.status === 401) {
      clearToken();
      if (window.location.pathname.startsWith('/login')) {
        window.location.reload();
      } else {
        // 保存当前URL，以便登录后返回
        const currentPath = window.location.pathname + window.location.search;
        const redirect = encodeURIComponent(currentPath.substring(1)); // 去除前导/并编码
        window.location.href = `/login?redirect=${redirect}`;
      }
      return Promise.reject(new Error(t('common.request.unauthorized')));
    }
    if (error?.response?.data?.message) {
      // 保留原始的 response 对象，以便组件可以访问完整的错误信息（包括 data 字段）
      // 创建一个新的 Error 对象，但保留原始的 response
      const err = new Error(error.response.data.message);
      (err as any).response = error.response;
      return Promise.reject(err);
    }
    if (error?.message && error.message.includes('timeout')) {
      // Message.error(t('common.request.timeout'));
      return Promise.reject(new Error(t('common.request.timeout') as string));
    }

    // eslint-disable-next-line no-console
    console.log(error);
    // Message.error(error.message);
    return Promise.reject(error);
  }
);

const axiosRequest = (config: any) => {
  const { url, method, params, data, headersType, responseType } = config;
  return axios({
    ...config,
    url,
    method,
    params,
    data,
    responseType,
    headers: {
      'Content-Type': headersType || 'application/json',
    },
  });
};

class ApiHelper {
  request: any;

  constructor(req: (config: any) => any) {
    this.request = req;
  }

  get<T = any>(
    url: string,
    params?: Record<string, any>,
    config?: AxiosRequestConfig<any>
  ) {
    return this.request({
      method: 'get',
      url,
      params,
      ...config,
    }) as unknown as Promise<T>;
  }

  post<T = any>(
    url: string,
    data?: Record<string, any>,
    config?: AxiosRequestConfig<any>
  ) {
    return this.request({
      method: 'post',
      url,
      data,
      ...config,
    }) as unknown as Promise<T>;
  }

  delete<T = any>(
    url: string,
    params?: Record<string, any>,
    config?: AxiosRequestConfig<any>
  ) {
    return this.request({
      method: 'delete',
      url,
      params,
      ...config,
    }) as unknown as Promise<T>;
  }

  put<T = any>(
    url: string,
    data?: Record<string, any>,
    config?: AxiosRequestConfig<any>
  ) {
    return this.request({
      method: 'put',
      url,
      data,
      ...config,
    }) as unknown as Promise<T>;
  }
}

export default new ApiHelper(axiosRequest);
