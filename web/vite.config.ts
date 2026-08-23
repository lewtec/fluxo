import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import relay from "vite-plugin-relay";
import tailwind from "@tailwindcss/vite";

export default defineConfig({
  plugins: [tailwind(), react(), relay],
  define: {
    // Identifier, not JSON: same globalThis as optimizeDeps so parse-torrent
    // sees the Buffer installed in main.tsx ({} was a detached stub).
    global: 'globalThis',
  },
  resolve: {
    alias: {
      buffer: 'buffer',
      path: 'path-browserify',
    }
  },
  server: {
    port: 5173,
    proxy: {
      "/graphql": {
        target: "http://localhost:8080",
        changeOrigin: true,
        ws: true,
      },
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  optimizeDeps: {
    esbuildOptions: {
      define: {
        global: 'globalThis'
      }
    }
  }
});
