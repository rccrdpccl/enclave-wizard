import react from "@vitejs/plugin-react-swc";
import { defineConfig } from "vite";

export default defineConfig((_env) => {
  return {
    plugins: [react()],
    server: {
      proxy: {
        "/api/v1": {
          target: process.env.API_PROXY_TARGET || "http://localhost:8080",
          changeOrigin: true,
        },
        "/openapi.json": {
          target: process.env.API_PROXY_TARGET || "http://localhost:8080",
          changeOrigin: true,
        },
      },
    },
  };
});
