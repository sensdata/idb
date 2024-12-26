import { mergeConfig } from 'vite';
import eslint from 'vite-plugin-eslint';
import baseConfig from './vite.config.base';

export default mergeConfig(
  {
    mode: 'development',
    server: {
      open: true,
      fs: {
        strict: true,
      },
      proxy: {
        '/api/ws': {
          target: 'ws://39.99.155.139:9918',
          ws: true,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api\/ws/, '/api/v1/ws'),
        },
        '/api': {
          target: 'http://39.99.155.139:9918',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '/api/v1'),
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
