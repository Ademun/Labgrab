<script lang="ts">
	import type { TelegramAuthData } from '$lib/api/types';
	import { onMount } from 'svelte';
	import { fade, fly, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import { goto } from '$app/navigation';

	import { api } from '$lib/api/client';
	import { handleApiError, toast } from '$lib/utils/toast';
	import { Spinner } from '$lib/components/ui/spinner';

	let isAuthenticating = $state<boolean>(false);

	async function onTelegramAuth(user: TelegramAuthData) {
		isAuthenticating = true;

		try {
			await api.authenticateWithTelegram(user);

			await goto('/auth/details');
		} catch (error) {
			console.error('Authentication failed:', error);

			handleApiError(error, 'Не удалось авторизоваться через Telegram');
		}

		isAuthenticating = false;
	}

	onMount(() => {
		(window as any).onTelegramAuth = onTelegramAuth;

		return () => {
			delete (window as any).onTelegramAuth;
		};
	});
</script>

<div class="textured-bg h-full w-full flex flex-col items-center p-8">
	<h1
		class="text-xl font-medium"
		in:fade={{
			delay: 100,
			duration: 300
		}}
		out:fade={{
			duration: 250,
			delay: 0
		}}
	>
		Вход
	</h1>

	<div
		class="bg-card rounded-2xl shadow-[0_4px_16px_rgba(0,0,0,0.06),0_2px_4px_rgba(0,0,0,0.08)]
               px-10 py-12 max-w-[400px] w-full text-center my-auto"
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
		<div
			class="font-black text-[2.5rem] leading-[0.9] mb-3 uppercase tracking-tight"
			in:fade={{
				delay: 100,
				duration: 400
			}}
			out:fade={{
				duration: 250,
				delay: 0
			}}
		>
			<div>LAB</div>
			<div>GR<span class="text-primary">AB</span></div>
		</div>

		<p
			class="text-md text-muted-foreground mb-10"
			in:fade={{
				delay: 400,
				duration: 400
			}}
			out:fade={{
				duration: 250,
				delay: 50
			}}
		>
			Автоматическая запись
		</p>

		<h1
			class="text-xl font-semibold mb-2"
			in:scale={{
				duration: 400,
				start: 0.9
			}}
			out:scale={{
				duration: 250,
				start: 0.9,
				delay: 0
			}}
		>
			Вход в систему
		</h1>

		<p
			class="text-md text-muted-foreground mb-8"
			in:fade={{
				delay: 100,
				duration: 400
			}}
			out:fade={{
				duration: 250,
				delay: 0
			}}
		>
			Войдите через Telegram чтобы продолжить
		</p>

		<div
			class="flex justify-center"
			in:scale={{
				duration: 700,
				start: 0.8
			}}
			out:scale={{
				duration: 250,
				start: 0.8,
				delay: 100
			}}
		>
			{#if isAuthenticating}
				<Spinner />
			{:else}
				<script
					async
					src="https://telegram.org/js/telegram-widget.js?22"
					data-telegram-login="ademun_timetable_bot"
					data-userpic="false"
					data-size="large"
					data-onauth="onTelegramAuth(user)"
					data-request-access="write"
					data-radius="8"
				>
				</script>
			{/if}
		</div>
	</div>
</div>
