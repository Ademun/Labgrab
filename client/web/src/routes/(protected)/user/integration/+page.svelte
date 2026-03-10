<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { toast } from 'svelte-sonner';
	import { Header } from '$lib/components/navigation/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { TriangleAlert } from '@lucide/svelte';
	import { fly } from 'svelte/transition';

	let { data } = $props();

	let showDialog = $state(false);
	let countdown = $state(5);
	let countdownInterval: ReturnType<typeof setInterval> | null = null;
	let isSubmitting = $state(false);

	const form = superForm(data.form, {
		onSubmit: () => {
			isSubmitting = true;
		},
		onResult: ({ result }) => {
			isSubmitting = false;
			if (result.type === 'failure' && result.data?.error) {
				toast.error(result.data.error as string);
			}
		}
	});

	const { form: formData, enhance, submit } = form;

	function openDialog() {
		if (!$formData.dikidi_phone_number || !$formData.dikidi_password) {
			toast.error('Заполните все поля');
			return;
		}
		showDialog = true;
		countdown = 5;
		countdownInterval = setInterval(() => {
			countdown -= 1;
			if (countdown <= 0) {
				clearInterval(countdownInterval!);
				countdownInterval = null;
			}
		}, 1000);
	}

	function handleCancel() {
		if (countdownInterval) {
			clearInterval(countdownInterval);
			countdownInterval = null;
		}
		showDialog = false;
	}

	function handleConfirm() {
		showDialog = false;
		submit();
	}
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Подключение аккаунта" back="/user" />

	<div class="flex flex-1 flex-col items-center justify-center px-6 py-8">
		<div class="w-full max-w-sm" in:fly={{ y: 16, duration: 300, delay: 60 }}>
			<div class="mb-5 flex flex-col gap-1 px-1">
				<p class="text-md leading-relaxed text-foreground text-center">
					Укажите данные от Dikidi — они хранятся в зашифрованном виде и используются только для
					автоматической записи на лабораторные.
				</p>
			</div>

			<div class="rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
				<form method="POST" action="?/connect" use:enhance class="space-y-5">
					<Form.Field {form} name="dikidi_phone_number">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label class="mb-4">Номер телефона</Form.Label>
								<Input
									{...props}
									type="tel"
									autocomplete="tel"
									disabled={isSubmitting}
									placeholder="+7 900 000 00 00"
									bind:value={$formData.dikidi_phone_number}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field {form} name="dikidi_password">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label class="mb-4">Пароль</Form.Label>
								<Input
									{...props}
									type="password"
									autocomplete="current-password"
									disabled={isSubmitting}
									placeholder="Пароль от Dikidi"
									bind:value={$formData.dikidi_password}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<!-- Скрытый submit, вызывается через submit() из попапа -->
					<button type="submit" class="sr-only" aria-hidden="true" tabindex="-1"></button>
				</form>

				<Button
					class="mt-2 w-full py-5 text-md uppercase tracking-widest font-semibold"
					disabled={isSubmitting}
					onclick={openDialog}
				>
					{isSubmitting ? 'Подключение...' : 'Подключить'}
				</Button>
			</div>
		</div>
	</div>
</div>

<AlertDialog.Root bind:open={showDialog}>
	<AlertDialog.Content class="max-w-sm rounded-3xl px-6 py-8">
		<AlertDialog.Header class="flex flex-col items-center gap-3 text-center">
			<TriangleAlert class="h-14 w-14 text-[#FF9F0A]" strokeWidth={1.5} />
			<AlertDialog.Title class="text-lg font-bold leading-snug">Вы уверены?</AlertDialog.Title>
			<AlertDialog.Description class="text-sm leading-relaxed text-muted-foreground text-center">
				Продолжая, вы подтверждаете, что осознаёте риски: возможную блокировку учётной записи и
				ответственность за нарушение пользовательского соглашения платформы.
				<br /><br />
				<span class="font-medium text-foreground">
					 Сервис не несёт
				ответственности за последствия.
				</span>
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer class="mt-6 flex flex-col gap-3">
			<AlertDialog.Action
				class="w-full py-5"
				disabled={countdown > 0}
				onclick={handleConfirm}
			>
				{countdown > 0 ? `Продолжить (${countdown})` : 'Продолжить'}
			</AlertDialog.Action>
			<AlertDialog.Cancel class="w-full py-5" onclick={handleCancel}>
				Отменить
			</AlertDialog.Cancel>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>