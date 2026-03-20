import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            target: 'http://172.16.96.174:10086/',
            ws: true,
          },
        },
      },
    },
  };
});
