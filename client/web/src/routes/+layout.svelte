<script lang="ts">
	import '../app.css';
	import { Navigation } from '$lib/components/navigation/index.js';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from 'svelte-sonner';

	let { children, data } = $props();
</script>

<div class="h-screen w-screen overflow-y-auto pb-28">
	<ModeWatcher />
	<Toaster position="top-center" richColors theme="system" />
	{@render children?.()}
</div>
<Navigation />
