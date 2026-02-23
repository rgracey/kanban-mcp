import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

// https://vite.dev/config/
export default defineConfig({
	plugins: [tailwindcss(), svelte()],
	server: {
		proxy: {
			'/api': 'http://localhost:8080'
		}
	}
});
