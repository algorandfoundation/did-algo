import { svelte } from '@sveltejs/vite-plugin-svelte';
import * as path from 'path';
import { nodePolyfills } from 'vite-plugin-node-polyfills';
import { defineConfig } from 'vitest/config';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte(),
    nodePolyfills({
      protocolImports: true
    })
  ],
  test: {
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,ts}']
  },
  resolve: {
    alias: {
      // To support absolute import paths for components:
      //   `import Counter from '~/lib/Counter.svelte';`
      '~': path.resolve('./src')
    }
  }
});
