<script>
	import { Button } from '$lib/components/ui/button/index.js';
</script>

<div
	class="textured-bg flex h-full w-full flex-col items-center justify-center gap-4 px-12 text-center"
>
	<div class="fixed top-0 flex items-start justify-between py-4">
		<div>
			<span class="text-main text-stats font-bold text-primary">24</span>
			<br />
			<span class="text-lg font-medium text-muted-foreground">Успешных записей</span>
		</div>
	</div>

	<div class="text-logo font-bold">
		LAB<br />GRAB
	</div>

	<span class="text-headline text-content text-accent-foreground">
		Автоматическая запись на лабораторные работы
	</span>
	<div>
		<Button class="text-md px-16 py-8 font-bold uppercase" href="/auth">Подключить</Button>
	</div>
</div>

<style>
	.text-stats {
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
