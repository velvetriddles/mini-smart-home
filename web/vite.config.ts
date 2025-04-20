import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Проксирование API-запросов на API Gateway
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      // Проксирование WebSocket-запросов
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
}); 