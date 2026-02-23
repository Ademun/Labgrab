<script lang="ts">
	import { Header } from '$lib/components/navigation/index.js';
	import { Heart } from '@lucide/svelte';
	import { fly } from 'svelte/transition';
	import qr from './payment.png';
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Поддержать разработчика" back="/user" />

	<div class="flex flex-1 flex-col items-center justify-center px-6 py-8">
		<div
			class="flex w-full max-w-sm flex-col items-center gap-6"
			in:fly={{ y: 16, duration: 300, delay: 60 }}
		>
			<div class="flex flex-col items-center gap-2 text-center">
				<Heart class="h-8 w-8 fill-[#FF3B30] text-[#FF3B30]" />
				<h2 class="text-lg font-semibold">Спасибо за вашу поддержку!</h2>
				<p class="text-sm text-muted-foreground">
					Это помогает развивать проект и держать его бесплатным.
				</p>
			</div>

			<div
				class="flex aspect-square items-center justify-center rounded-2xl border-2 border-dashed border-border bg-white p-2 text-muted-foreground"
			>
				<img src={qr} alt="QR - код" />
			</div>

			<p class="text-xs text-muted-foreground">Отсканируйте QR-код для перевода</p>
		</div>
	</div>
</div>
