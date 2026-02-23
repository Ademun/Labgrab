<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { onMount } from 'svelte';
	import { fly, fade, scale } from 'svelte/transition';
	import { cubicOut, backOut } from 'svelte/easing';

	const TARGET_COUNT = 24;

	let displayCount = $state(0);
	let mounted = $state(false);

	function animateCount(target: number, duration = 2000) {
		const start = performance.now();
		function tick(now: number) {
			const elapsed = now - start;
			const progress = Math.min(elapsed / duration, 1);
			const eased = 1 - Math.pow(1 - progress, 3);
			displayCount = Math.round(eased * target);
			if (progress < 1) requestAnimationFrame(tick);
		}
		requestAnimationFrame(tick);
	}

	onMount(() => {
		const t = setTimeout(() => {
			mounted = true;
			animateCount(TARGET_COUNT);
		}, 80);
		return () => clearTimeout(t);
	});
</script>

{#if mounted}
	<div
		class="textured-bg flex h-full w-full flex-col items-center justify-center gap-4 px-12 text-center"
		in:fade={{ duration: 400, easing: cubicOut }}
	>
		<div
			class="fixed top-0 flex items-start justify-between py-4"
			in:fly={{ y: -20, duration: 420, delay: 100, easing: cubicOut }}
		>
			<div>
				<span class="text-main text-stats counter-value font-bold text-primary">
					{displayCount}
				</span>
				<br />
				<span class="text-lg font-medium text-muted-foreground">Успешных записей</span>
			</div>
		</div>

		<div class="logo-spring" in:scale={{ start: 0.82, duration: 600, delay: 80, easing: backOut }}>
			<div class="text-logo font-bold">
				LAB<br />GRAB
			</div>
		</div>

		<span
			class="text-headline text-content text-accent-foreground"
			in:fly={{ y: 18, duration: 380, delay: 260, easing: cubicOut }}
		>
			Автоматическая запись на лабораторные работы
		</span>

		<div in:fly={{ y: 16, duration: 380, delay: 360, easing: backOut }}>
			<Button class="text-md btn-pulse px-16 py-8 font-bold uppercase" href="/auth">
				Подключить
			</Button>
		</div>
	</div>
{/if}

<style>
	.text-stats {
		font-size: 2rem;
	}

	.counter-value {
		display: inline-block;
		min-width: 2.5ch;
		text-align: right;
		font-variant-numeric: tabular-nums;
	}

	.text-logo {
		font-family: 'Archivo Black', sans-serif;
		font-size: 5rem;
		line-height: 1;
		color: transparent;
		background: linear-gradient(
			135deg,
			var(--foreground),
			var(--foreground) 68%,
			var(--primary) 68%,
			var(--primary) 85%,
			var(--foreground) 85%
		);
		background-clip: text;
		-webkit-background-clip: text;
	}

	.text-headline {
		font-size: 1.255rem;
	}
	.logo-spring {
		will-change: transform, opacity;
	}

	.btn-pulse {
		animation: button-breathe 3.2s ease-in-out 1.2s infinite;
	}

	@keyframes button-breathe {
		0%,
		100% {
			box-shadow: 0 0 0 0 color-mix(in oklch, var(--primary), transparent 80%);
		}
		50% {
			box-shadow: 0 0 0 8px color-mix(in oklch, var(--primary), transparent 100%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.btn-pulse {
			animation: none;
		}
		.logo-spring {
			animation: none;
		}
	}
</style>
