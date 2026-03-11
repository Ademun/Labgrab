import devtoolsJson from 'vite-plugin-devtools-json';
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
	server: { port: 443, host: 'localhost' },
	plugins: [tailwindcss(), sveltekit(), devtoolsJson()],
	resolve: { alias: { $lib: path.resolve('./src/lib') } }
});
