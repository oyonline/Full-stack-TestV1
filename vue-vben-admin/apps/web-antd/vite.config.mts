import { defineConfig } from '@vben/vite-config';
import { loadEnv } from 'vite';

export default defineConfig(async ({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  if (mode === 'production' && !env.VITE_GLOB_API_URL?.trim()) {
    throw new Error('VITE_GLOB_API_URL must be set for production builds.');
  }

  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            target: 'http://127.0.0.1:10082/',
            ws: true,
          },
          '/static': {
            changeOrigin: true,
            target: 'http://127.0.0.1:10082/',
          },
          '/form-generator': {
            changeOrigin: true,
            target: 'http://127.0.0.1:10082/',
          },
        },
      },
    },
  };
});
