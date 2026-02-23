<script lang="ts">
	import { ChevronLeft } from '@lucide/svelte';
	import { goto } from '$app/navigation';

	type Props = {
		title: string;
		back?: string;
	};

	let { title, back }: Props = $props();

	function titleTransition(_node: Element, { duration = 280 }: { duration?: number } = {}) {
		return {
			duration,
			css: (t: number) => {
				const ease = 1 - Math.pow(1 - t, 3);
				const scale = 0.72 + 0.28 * ease;
				const blur = (1 - ease) * 10;
				const opacity = ease;
				return `
					transform: scale(${scale});
					filter: blur(${blur}px);
					opacity: ${opacity};
					will-change: transform, filter, opacity;
				`;
			}
		};
	}
</script>

<header
	class="sticky top-0 z-10 flex items-center justify-center border-b border-border bg-background px-4 py-4"
>
	{#if back}
		<button
			type="button"
			class="absolute left-4 flex items-center justify-center text-muted-foreground transition-colors hover:text-foreground"
			onclick={() => goto(back!)}
		>
			<ChevronLeft class="h-5 w-5" />
		</button>
	{/if}

	{#key title}
		<h1
			class="text-base font-semibold"
			transition:titleTransition
		>
			{title}
		</h1>
	{/key}
</header>