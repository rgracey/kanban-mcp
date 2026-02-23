import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

// https://vite.dev/config/
export default defineConfig({
	plugins: [tailwindcss(), svelte()],
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				// Disable response buffering so SSE streams are forwarded immediately
				configure: (proxy) => {
					proxy.on('proxyRes', (proxyRes) => {
						const ct = proxyRes.headers['content-type'] ?? '';
						if (ct.includes('text/event-stream')) {
							proxyRes.headers['cache-control'] = 'no-cache';
						}
					});
				}
			}
		}
	}
});
