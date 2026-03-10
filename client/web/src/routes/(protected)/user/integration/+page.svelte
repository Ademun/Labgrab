<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { applyAction } from '$app/forms';
	import { superForm } from 'sveltekit-superforms';
	import { toast } from 'svelte-sonner';
	import { Header } from '$lib/components/navigation/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { TriangleAlert, CheckCircle, Circle, Clock, RefreshCw } from '@lucide/svelte';
	import { fly } from 'svelte/transition';
	import { api } from '$lib/api/client.js';

	let { data } = $props();

	const apiReady = $derived(data.user?.api_ready ?? false);
	const userInfo = $derived(data.userInfo ?? null);

	function formatLastAuth(iso: string | null | undefined): string {
		if (!iso) return 'Нет данных';
		return new Intl.DateTimeFormat('ru', {
			day: 'numeric',
			month: 'long',
			hour: '2-digit',
			minute: '2-digit'
		}).format(new Date(iso));
	}

	// ─── Auth overlay states ──────────────────────────────────────────────────
	let isAuthenticating = $state(false);
	let isAuthSuccess = $state(false);
	let dikidiAuthFailed = $state(false);

	async function runDikidiAuth() {
		isAuthenticating = true;
		isAuthSuccess = false;
		dikidiAuthFailed = false;
		try {
			await api.dikidiAuth();
			isAuthenticating = false;
			isAuthSuccess = true;
			await new Promise((r) => setTimeout(r, 1800));
			await goto('/user');
		} catch {
			dikidiAuthFailed = true;
		} finally {
			isAuthenticating = false;
		}
	}

	// ─── Connect dialog (first-time) ─────────────────────────────────────────
	let showConnectDialog = $state(false);
	let connectCountdown = $state(5);
	let connectCountdownInterval: ReturnType<typeof setInterval> | null = null;

	function openConnectDialog() {
		if (!$connectFormData.dikidi_phone_number || !$connectFormData.dikidi_password) {
			toast.error('Заполните все поля');
			return;
		}
		showConnectDialog = true;
		connectCountdown = 5;
		connectCountdownInterval = setInterval(() => {
			connectCountdown -= 1;
			if (connectCountdown <= 0) {
				clearInterval(connectCountdownInterval!);
				connectCountdownInterval = null;
			}
		}, 1000);
	}

	function handleConnectCancel() {
		if (connectCountdownInterval) {
			clearInterval(connectCountdownInterval);
			connectCountdownInterval = null;
		}
		showConnectDialog = false;
	}

	function handleConnectConfirm() {
		showConnectDialog = false;
		connectSubmit();
	}

	// ─── Manual auth dialog ───────────────────────────────────────────────────
	let showAuthDialog = $state(false);
	let authCountdown = $state(5);
	let authCountdownInterval: ReturnType<typeof setInterval> | null = null;

	function openAuthDialog() {
		showAuthDialog = true;
		authCountdown = 5;
		authCountdownInterval = setInterval(() => {
			authCountdown -= 1;
			if (authCountdown <= 0) {
				clearInterval(authCountdownInterval!);
				authCountdownInterval = null;
			}
		}, 1000);
	}

	function handleAuthCancel() {
		if (authCountdownInterval) {
			clearInterval(authCountdownInterval);
			authCountdownInterval = null;
		}
		showAuthDialog = false;
	}

	async function handleAuthConfirm() {
		showAuthDialog = false;
		await runDikidiAuth();
	}

	// ─── Connect form (first-time) ────────────────────────────────────────────
	let isConnectSubmitting = $state(false);

	const connectForm = superForm(data.connectForm!, {
		applyAction: false,
		onSubmit: () => { isConnectSubmitting = true; },
		onResult: async ({ result }) => {
			isConnectSubmitting = false;
			if (result.type === 'failure') {
				await applyAction(result);
				if (result.data?.error) toast.error(result.data.error as string);
				return;
			}
			if (result.type === 'redirect') {
				await runDikidiAuth();
			}
		}
	});

	const { form: connectFormData, enhance: connectEnhance, submit: connectSubmit } = connectForm;

	// ─── Update form ──────────────────────────────────────────────────────────
	let isUpdateSubmitting = $state(false);

	const updateForm = superForm(data.updateForm!, {
		onSubmit: () => { isUpdateSubmitting = true; },
		onResult: async ({ result }) => {
			isUpdateSubmitting = false;
			if (result.type === 'failure' && result.data?.error) {
				toast.error(result.data.error as string);
				return;
			}
			if (result.type === 'success') {
				toast.success('Данные обновлены');
				await invalidateAll();
			}
		}
	});

	const { form: updateFormData, enhance: updateEnhance } = updateForm;
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Подключение аккаунта" back="/user" />

	<div class="flex flex-1 flex-col items-center justify-center px-6 py-8">
		<div class="w-full max-w-sm" in:fly={{ y: 16, duration: 300, delay: 60 }}>
			{#if isAuthenticating}
				<div
					class="flex flex-col items-center gap-5 rounded-2xl border border-border/40 bg-card px-6 py-10 text-center shadow-xl"
				>
					<div class="h-10 w-10 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
					<div class="flex flex-col gap-1">
						<p class="text-base font-semibold">Подключение к Dikidi...</p>
						<p class="text-sm leading-relaxed text-muted-foreground">
							Это может занять до 30 секунд. Не закрывайте страницу.
						</p>
					</div>
				</div>
			{:else if isAuthSuccess}
				<div
					class="flex flex-col items-center gap-5 rounded-2xl border border-border/40 bg-card px-6 py-10 text-center shadow-xl"
					in:fly={{ y: 8, duration: 250 }}
				>
					<CheckCircle class="h-10 w-10 text-[#34C759]" strokeWidth={1.5} />
					<div class="flex flex-col gap-1">
						<p class="text-base font-semibold">Аккаунт подключён</p>
						<p class="text-sm leading-relaxed text-muted-foreground">Переходим в профиль...</p>
					</div>
				</div>
			{:else if dikidiAuthFailed}
				<div
					class="flex flex-col items-center gap-5 rounded-2xl border border-destructive/20 bg-card px-6 py-10 text-center shadow-xl"
				>
					<TriangleAlert class="h-10 w-10 text-destructive" strokeWidth={1.5} />
					<div class="flex flex-col gap-1">
						<p class="text-base font-semibold">Не удалось подключиться</p>
						<p class="text-sm leading-relaxed text-muted-foreground">
							Система автоматически повторит попытку позже.<br />
							Если ошибка повторяется — проверьте введённые данные.
						</p>
					</div>
					<Button variant="outline" class="w-full py-5" onclick={() => { dikidiAuthFailed = false; }}>
						Попробовать снова
					</Button>
				</div>
			{:else if userInfo}
				<div class="flex flex-col gap-4">
					<!-- Статус -->
					<div class="flex flex-col gap-4 rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
						<div class="flex items-center justify-between">
							<span class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
								Статус подключения
							</span>
							{#if userInfo.api_authed}
								<span class="flex items-center gap-1.5 text-sm font-semibold text-[#34C759]">
									<Circle class="h-2.5 w-2.5 fill-[#34C759]" strokeWidth={0} />
									Активно
								</span>
							{:else}
								<span class="flex items-center gap-1.5 text-sm font-semibold text-destructive">
									<Circle class="h-2.5 w-2.5 fill-destructive" strokeWidth={0} />
									Не авторизовано
								</span>
							{/if}
						</div>

						<div class="h-px bg-border/40"></div>

						<div class="flex flex-col gap-3 text-sm">
							<div class="flex items-center justify-between">
								<span class="text-muted-foreground">Номер телефона</span>
								<span class="font-medium">{userInfo.phone_number}</span>
							</div>
							<div class="flex items-center justify-between">
								<span class="flex items-center gap-1.5 text-muted-foreground">
									<Clock class="h-3.5 w-3.5" />
									Последняя авторизация
								</span>
								<span class="font-medium">{formatLastAuth(userInfo.last_auth)}</span>
							</div>
						</div>

						<div class="h-px bg-border/40"></div>

						<p class="text-xs leading-relaxed text-muted-foreground">
							Система автоматически повторяет попытку авторизации каждые 10 минут.
						</p>

						<Button
							variant="outline"
							class="w-full py-5"
							onclick={openAuthDialog}
						>
							<RefreshCw class="mr-2 h-4 w-4" />
							Авторизоваться сейчас
						</Button>
					</div>

					<!-- Форма обновления данных -->
					<div class="rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
						<p class="mb-5 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
							Обновить данные
						</p>
						<form method="POST" action="?/update" use:updateEnhance class="space-y-5">
							<Form.Field form={updateForm} name="dikidi_phone_number">
								<Form.Control>
									{#snippet children({ props })}
										<Form.Label class="mb-4">Номер телефона</Form.Label>
										<Input
											{...props}
											type="tel"
											autocomplete="tel"
											disabled={isUpdateSubmitting}
											placeholder="+7 900 000 00 00"
											bind:value={$updateFormData.dikidi_phone_number}
										/>
									{/snippet}
								</Form.Control>
								<Form.FieldErrors />
							</Form.Field>

							<Form.Field form={updateForm} name="dikidi_password">
								<Form.Control>
									{#snippet children({ props })}
										<Form.Label class="mb-4">Новый пароль</Form.Label>
										<Input
											{...props}
											type="password"
											autocomplete="new-password"
											disabled={isUpdateSubmitting}
											placeholder="Введите новый пароль"
											bind:value={$updateFormData.dikidi_password}
										/>
									{/snippet}
								</Form.Control>
								<Form.FieldErrors />
							</Form.Field>

							<Button
								type="submit"
								class="w-full py-5 text-md uppercase tracking-widest font-semibold"
								disabled={isUpdateSubmitting}
							>
								{isUpdateSubmitting ? 'Сохранение...' : 'Сохранить'}
							</Button>
						</form>
					</div>
				</div>
			{:else if apiReady}
				<div class="mb-5 flex flex-col gap-1 px-1">
					<p class="text-md leading-relaxed text-foreground text-center">
						Укажите данные от Dikidi — они хранятся в зашифрованном виде и используются только для
						автоматической записи на лабораторные.
					</p>
				</div>

				<div class="rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
					<form method="POST" action="?/connect" use:connectEnhance class="space-y-5">
						<Form.Field form={connectForm} name="dikidi_phone_number">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label class="mb-4">Номер телефона</Form.Label>
									<Input
										{...props}
										type="tel"
										autocomplete="tel"
										disabled={isConnectSubmitting}
										placeholder="+7 900 000 00 00"
										bind:value={$connectFormData.dikidi_phone_number}
									/>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>

						<Form.Field form={connectForm} name="dikidi_password">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label class="mb-4">Пароль</Form.Label>
									<Input
										{...props}
										type="password"
										autocomplete="current-password"
										disabled={isConnectSubmitting}
										placeholder="Пароль от Dikidi"
										bind:value={$connectFormData.dikidi_password}
									/>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>

						<button type="submit" class="sr-only" aria-hidden="true" tabindex="-1"></button>
					</form>

					<Button
						class="mt-2 w-full py-5 text-md uppercase tracking-widest font-semibold"
						disabled={isConnectSubmitting}
						onclick={openConnectDialog}
					>
						{isConnectSubmitting ? 'Подключение...' : 'Подключить'}
					</Button>
				</div>
			{:else}
				<div class="flex flex-col items-center gap-6 px-1 text-center">
					<div class="flex flex-col gap-2">
						<p class="text-base font-semibold">Сначала заполните профиль</p>
						<p class="text-sm leading-relaxed text-muted-foreground">
							Перед подключением аккаунта необходимо заполнить персональную информацию.
						</p>
					</div>
					<a href="/user/details">
						<Button class="px-8 py-5">Заполнить профиль</Button>
					</a>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Connect dialog -->
<AlertDialog.Root bind:open={showConnectDialog}>
	<AlertDialog.Content class="max-w-sm rounded-3xl px-6 py-8">
		<AlertDialog.Header class="flex flex-col items-center gap-3 text-center">
			<TriangleAlert class="h-14 w-14 text-[#FF9F0A]" strokeWidth={1.5} />
			<AlertDialog.Title class="text-lg font-bold leading-snug">Вы уверены?</AlertDialog.Title>
			<AlertDialog.Description class="text-center text-sm leading-relaxed text-muted-foreground">
				Продолжая, вы подтверждаете, что осознаёте риски: возможную блокировку учётной записи и
				ответственность за нарушение пользовательского соглашения платформы.
				<br /><br />
				<span class="font-medium text-foreground">Сервис не несёт ответственности за последствия.</span>
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer class="mt-6 flex flex-col gap-3">
			<AlertDialog.Action
				class="w-full py-5"
				disabled={connectCountdown > 0}
				onclick={handleConnectConfirm}
			>
				{connectCountdown > 0 ? `Продолжить (${connectCountdown})` : 'Продолжить'}
			</AlertDialog.Action>
			<AlertDialog.Cancel class="w-full py-5" onclick={handleConnectCancel}>
				Отменить
			</AlertDialog.Cancel>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<!-- Manual auth dialog -->
<AlertDialog.Root bind:open={showAuthDialog}>
	<AlertDialog.Content class="max-w-sm rounded-3xl px-6 py-8">
		<AlertDialog.Header class="flex flex-col items-center gap-3 text-center">
			<RefreshCw class="h-14 w-14 text-primary" strokeWidth={1.5} />
			<AlertDialog.Title class="text-lg font-bold leading-snug">Авторизоваться сейчас?</AlertDialog.Title>
			<AlertDialog.Description class="text-center text-sm leading-relaxed text-muted-foreground">
				Будет выполнена попытка авторизации в Dikidi. Это может занять до 30 секунд.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer class="mt-6 flex flex-col gap-3">
			<AlertDialog.Action
				class="w-full py-5"
				disabled={authCountdown > 0}
				onclick={handleAuthConfirm}
			>
				{authCountdown > 0 ? `Продолжить (${authCountdown})` : 'Продолжить'}
			</AlertDialog.Action>
			<AlertDialog.Cancel class="w-full py-5" onclick={handleAuthCancel}>
				Отменить
			</AlertDialog.Cancel>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>