import { defineConfig, mergeConfig } from 'vite';
import https from 'https';
import http from 'http';
import eslint from 'vite-plugin-eslint';
import baseConfig from './vite.config.base';

// 代理目标服务器配置
const PROXY_HOST = '39.99.155.139';
const PROXY_PORT = 9918;

// 尝试发起一次 HTTP/HTTPS 请求，判断能否连通
function tryRequest(protocol: 'https' | 'http'): Promise<boolean> {
  return new Promise((resolve) => {
    const client = protocol === 'https' ? https : http;
    const req = client.request(
      {
        hostname: PROXY_HOST,
        port: PROXY_PORT,
        path: '/api/v1/public/version',
        method: 'GET',
        timeout: 2000,
        ...(protocol === 'https' ? { rejectUnauthorized: false } : {}),
      },
      (res) => {
        // 收到响应就认为该协议可用
        res.resume();
        resolve(true);
      }
    );

    req.on('error', () => resolve(false));
    req.on('timeout', () => {
      req.destroy();
      resolve(false);
    });

    req.end();
  });
}

// 启动时检测后端是 HTTP 还是 HTTPS
async function detectProtocol(): Promise<'https' | 'http'> {
  if (await tryRequest('https')) {
    console.log('[dev-proxy] 检测到后端使用 HTTPS');
    return 'https';
  }

  if (await tryRequest('http')) {
    console.log('[dev-proxy] 检测到后端使用 HTTP');
    return 'http';
  }

  console.log('[dev-proxy] 无法检测到后端协议，默认使用 HTTPS');
  return 'https';
}

export default defineConfig(async () => {
  const protocol = await detectProtocol();
  const httpTarget = `${protocol}://${PROXY_HOST}:${PROXY_PORT}`;
  const wsTarget = `${
    protocol === 'https' ? 'wss' : 'ws'
  }://${PROXY_HOST}:${PROXY_PORT}`;

  console.log(`[dev-proxy] 使用 ${protocol.toUpperCase()} 作为后端协议`);

  return mergeConfig(
    {
      mode: 'development',
      server: {
        open: true,
        fs: {
          strict: true,
        },
        proxy: {
          '^/api/v1/terminals/.*?/start': {
            target: wsTarget,
            ws: true,
            changeOrigin: true,
            // HTTPS 下关闭证书校验，HTTP 下该配置无效
            secure: protocol === 'https' ? false : undefined,
          },
          '^/api/v1/docker/.*?/containers/terminal(\\?|$)': {
            target: wsTarget,
            ws: true,
            changeOrigin: true,
            secure: protocol === 'https' ? false : undefined,
          },
          '/api/v1': {
            target: httpTarget,
            changeOrigin: true,
            secure: protocol === 'https' ? false : undefined,
          },
        },
        port: 5300,
      },
      plugins: [
        eslint({
          cache: false,
          include: ['src/**/*.ts', 'src/**/*.tsx', 'src/**/*.vue'],
          exclude: ['node_modules'],
        }),
      ],
    },
    baseConfig
  );
});
