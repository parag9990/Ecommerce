import react from "@vitejs/plugin-react";
import { loadEnv } from "vite";
import { defineConfig } from "vitest/config";
export default defineConfig(function (_a) {
    var mode = _a.mode;
    var env = loadEnv(mode, process.cwd(), "");
    var proxyTarget = env.VITE_API_PROXY_TARGET || "http://localhost:8080";
    return {
        plugins: [react()],
        server: {
            port: 5174,
            strictPort: false,
            proxy: {
                "/api": {
                    target: proxyTarget,
                    changeOrigin: true,
                    secure: false
                }
            }
        },
        preview: {
            port: 4174,
            strictPort: false
        },
        test: {
            css: true,
            environment: "jsdom",
            globals: true,
            setupFiles: "./src/test/setup.ts"
        }
    };
});
