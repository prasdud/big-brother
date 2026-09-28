import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.BB_DEV_API ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: fileURLToPath(new URL("../internal/web/dist", import.meta.url)),
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": apiTarget,
      "/healthz": apiTarget,
      "/metrics": apiTarget,
    },
  },
});
