import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const serverProxyTarget = `http://localhost:${env.PB_SERVER_PORT}`;

  return {
    root: "src/web",
    build: {
      outDir: "../../dist",
      emptyOutDir: true,
    },
    server: {
      proxy: {
        "/ws": {
          target: serverProxyTarget,
          ws: true,
        },
        "/app.v1.AppService": {
          target: serverProxyTarget,
        },
      },
    },
  };
});
