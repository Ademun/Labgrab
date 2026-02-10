<script>
	import { Button } from '$lib/components/ui/button/index.ts';
	import { fade, fly, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
</script>

<div class="w-full lines h-full flex flex-col items-center justify-center text-center gap-12">
	<div
		class="fixed top-0 flex justify-between items-start py-4"
		in:fly={{
			y: -20,
			duration: 500,
			opacity: 0
		}}
		out:fly={{
			y: -20,
			duration: 250,
			opacity: 0
		}}
	>
		<div>
			<span
				class="text-stats font-bold text-primary"
				in:fade={{
					delay: 200,
					duration: 400
				}}
				out:fade={{
					duration: 200,
					delay: 0
				}}>24</span
			>
			<br />
			<span
				class="text-lg font-medium text-muted-foreground"
				in:fade={{
					delay: 400,
					duration: 400
				}}
				out:fade={{
					duration: 200,
					delay: 100
				}}>Успешных записей</span
			>
		</div>
	</div>

	<!-- Main logo - center stage with playful entrance -->
	<div
		class="text-logo font-bold"
		in:fly={{
			y: 30,
			duration: 700,
			easing: backOut,
			opacity: 0
		}}
		out:fly={{
			y: 30,
			duration: 300,
			easing: backOut,
			opacity: 0
		}}
	>
		LAB<br />GRAB
	</div>

	<span
		class="text-headline text-accent-foreground"
		in:fade={{
			delay: 300,
			duration: 400
		}}
		out:fade={{
			duration: 250,
			delay: 50
		}}
	>
		Автоматическая запись на лабораторные работы. Подключи и отдыхай.
	</span>

	<div
		in:scale={{
			duration: 600,
			start: 0.85,
			delay: 500
		}}
		out:scale={{
			duration: 250,
			start: 0.85,
			delay: 100
		}}
	>
		<Button class="px-16 py-8 text-md font-bold" href="/auth">ПОДКЛЮЧИТЬ</Button>
	</div>
</div>

<style>
	.lines {
		background: url('../lib/assets/grid.svg') no-repeat center center;
	}

	.text-stats {
		font-family: 'Archivo Black', sans-serif;
		font-size: 2rem;
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
	}

	.text-headline {
		font-size: 1.255rem;
	}
</style>
