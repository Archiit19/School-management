import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api/auth": {
        target: "http://localhost:8081",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/auth/, ""),
      },
      "/api/users": {
        target: "http://localhost:8082",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/users/, ""),
      },
      "/api/academic": {
        target: "http://localhost:8083",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/academic/, ""),
      },
      "/api/students": {
        target: "http://localhost:8084",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/students/, ""),
      },
      "/api/attendance": {
        target: "http://localhost:8085",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/attendance/, ""),
      },
      "/api/exams": {
        target: "http://localhost:8086",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/exams/, ""),
      },
      "/api/finance": {
        target: "http://localhost:8087",
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api\/finance/, ""),
      },
    },
  },
});
