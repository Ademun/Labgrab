<script lang="ts">
	import '../app.css';
	import { Navigation } from '$lib/components/navigation/index.js';
	import { ModeWatcher } from 'mode-watcher';

	let { children, data } = $props();
</script>

<div class="h-screen w-screen overflow-y-auto pb-28">
	<ModeWatcher />
	{@render children?.()}
</div>
<Navigation />
