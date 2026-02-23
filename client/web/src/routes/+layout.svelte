<script lang="ts">
	import '../app.css';
	import { Navigation } from '$lib/components/navigation/index.js';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from 'svelte-sonner';
	import { onNavigate } from '$app/navigation';

	let { children, data } = $props();

	const TAB_ROUTES = ['/', '/subscriptions', '/user'];

	function isTab(path: string): boolean {
		return TAB_ROUTES.includes(path);
	}

	function getTransitionType(from: string, to: string): 'tab' | 'push' | 'pop' {
		if (isTab(from) && isTab(to)) return 'tab';
		const fromDepth = from.split('/').filter(Boolean).length;
		const toDepth = to.split('/').filter(Boolean).length;
		return toDepth >= fromDepth ? 'push' : 'pop';
	}

	onNavigate((navigation) => {
		if (!document.startViewTransition) return;

		const from = navigation.from?.url.pathname ?? '/';
		const to = navigation.to?.url.pathname ?? '/';
		const type = getTransitionType(from, to);

		document.documentElement.setAttribute('data-transition', type);

		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});
</script>

<div class="flex h-dvh w-full flex-col" style="view-transition-name: page;">
	<ModeWatcher />
	<Toaster position="top-center" richColors theme="system" />
	<div class="flex-1 overflow-y-auto">
		{@render children?.()}
	</div>
	<Navigation />
</div>
