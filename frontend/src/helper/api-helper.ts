import axios from 'axios';
import type { AxiosRequestConfig, AxiosResponse } from 'axios';
import { Message } from '@arco-design/web-vue';
import { getToken } from '@/utils/auth';
import { t } from '@/utils/i18n';

export interface ApiResponse<T = unknown> {
  code: number;
  message?: string;
  data?: T;
}

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';
export const API_PROXY_PREFIX = 'proxy/';
export const API_PROXY_BASE_URL = API_BASE_URL + API_PROXY_PREFIX;
axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL;

axios.interceptors.request.use(
  (config: AxiosRequestConfig) => {
    // let each request carry token
    // this example using the JWT token
    // Authorization is a custom headers key
    // please modify it according to the actual situation
    const token = getToken();
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
    if (response.config.responseType === 'blob') {
      // 如果是文件流，直接过
      return response;
    }

    if (response.data.code === 200) {
      return response.data.data;
    }

    Message.error(response.data.message as string);

    return Promise.reject(response.data.message);
  },
  (error) => {
    if (
      error.response?.status === 401 &&
      !window.location.pathname.startsWith('/login')
    ) {
      window.location.href = '/login';
      return Promise.reject(error);
    }
    if (error?.response?.data?.message) {
      Message.error(String(error.response.data.message));
      return Promise.reject(error.response.data.message);
    }
    if (error?.message && error.message.includes('timeout')) {
      Message.error(t('common.request.timeout'));
      return Promise.reject(new Error(t('common.request.timeout') as string));
    }

    // eslint-disable-next-line no-console
    console.log(error); // for debug
    Message.error(error.message);
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

  put<T = any>(
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
}

export default new ApiHelper(axiosRequest);
