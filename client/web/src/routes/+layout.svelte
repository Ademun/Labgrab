<script lang="ts">
	import '../app.css';
	import { Navigation } from '$lib/components/navigation/index.js';
	import { onMount } from 'svelte';
	import { app } from '$lib/stores/app.svelte.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { ModeWatcher } from 'mode-watcher';

	let { children } = $props();
	let isLoading = $state(false);

	onMount(async () => {
		isLoading = true;
		try {
			await app.load();
			isLoading = false;
		} catch (error) {
			console.error(error);
		}
	});
</script>

<div class="h-screen w-screen overflow-y-auto pb-28">
	<ModeWatcher />
	{#if isLoading}
		<Spinner />
	{:else}
		{@render children?.()}
	{/if}
</div>
<Navigation />
