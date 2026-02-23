<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import type { AuthRequest } from '$lib/api/schema/auth.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { api } from '$lib/api/client.js';
	import { toast } from 'svelte-sonner';
	import { getErrorMessage } from '$lib/utils/toast-errors.js';
	import { Header } from '$lib/components/navigation/index.js';

	let isAuthenticating = $state(false);

	async function onTelegramAuth(data: AuthRequest) {
		isAuthenticating = true;
		try {
			await api.auth(data);
			await goto('/user');
		} catch (error) {
			toast.error(getErrorMessage(error));
			isAuthenticating = false;
		}
	}

	onMount(() => {
		(window as any).onTelegramAuth = onTelegramAuth;

		return () => {
			delete (window as any).onTelegramAuth;
		};
	});
</script>

<div class="textured-bg flex h-full w-full flex-col">
	<Header title="Вход" />

	<div class="flex flex-1 items-center justify-center">
		<div
			class="card-border flex max-w-8/10 flex-col items-center justify-center gap-4 rounded-lg border border-border/40 bg-card px-6 py-6 shadow-xl"
		>
			<h1 class="text-main text-center text-xl leading-none font-bold">
				LAB
				<br />
				<span class="text-primary">GRAB</span>
			</h1>
			<p class="wrap text-center">Войдите через Telegram чтобы продолжить</p>
			<div class="flex justify-center">
				{#if isAuthenticating}
					<div class="flex-between flex items-center gap-4">
						<Spinner />
						<span>Авторизация...</span>
					</div>
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
</div>
