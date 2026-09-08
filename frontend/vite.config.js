/// <reference types="vitest/config" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        port: 3000,
    },
    build: {
        // The Go server serves ./build (see main.go), so emit there instead of
        // Vite's default ./dist to keep the Dockerfile and server unchanged.
        outDir: 'build',
    },
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: './src/setupTests.js',
    },
});
