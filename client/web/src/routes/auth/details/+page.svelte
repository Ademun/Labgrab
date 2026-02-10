<script lang="ts">
	import { Field, Group, Label, Legend, Set } from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { fade, scale } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import type { UserUpdateRequest } from '$lib/api/types';

	import { api } from '$lib/api/client';
	import { handleApiError, toast } from '$lib/utils/toast';

	let name = $state<string>('');
	let surname = $state<string>('');
	let patronymic = $state<string>('');
	let group_code = $state<string>('');
	let phone_number = $state<string>('');

	let isSubmitting = $state<boolean>(false);

	function validatePhoneNumber(phone: string): boolean {
		if (!phone) return true;

		const cleaned = phone.replace(/[\s-]/g, '');

		const phoneRegex = /^(\+7|8)\d{10}$/;
		return phoneRegex.test(cleaned);
	}

	const submitDetails = async () => {
		if (!validatePhoneNumber(phone_number)) {
			toast.error('Неверный формат номера телефона', {
				description: 'Используйте формат: +7 978 621 48 21 или 8 978 621 48 21'
			});
			return;
		}

		const updates: UserUpdateRequest = {};
		if (name.trim()) updates.name = name.trim();
		if (surname.trim()) updates.surname = surname.trim();
		if (patronymic.trim()) updates.patronymic = patronymic.trim();
		if (group_code.trim()) updates.group_code = group_code.trim();
		if (phone_number.trim()) updates.phone_number = phone_number.trim();

		if (Object.keys(updates).length === 0) {
			await skip();
			return;
		}

		isSubmitting = true;

		try {
			await api.updateUser(updates);

			await goto('/account');
		} catch (error) {
			console.error('Failed to update profile:', error);
			handleApiError(error, 'Не удалось обновить профиль');
		} finally {
			isSubmitting = false;
		}
	};

	const skip = async () => {
		toast.info('Вы можете заполнить профиль позже в настройках');
		await goto('/account');
	};
</script>

<div class="flex flex-col items-center h-full w-full mx-auto px-8 py-8">
	<div
		class="my-auto bg-card text-center rounded-2xl shadow-lg p-6 max-w-md w-full"
		in:scale={{
			delay: 100,
			duration: 300,
			start: 0.95
		}}
		out:scale={{
			duration: 250,
			start: 0.95,
			delay: 0
		}}
	>
		<h1
			class="text-xl font-bold font-archivo-black tracking-wider mb-2"
			in:fade={{
				delay: 200,
				duration: 300
			}}
			out:fade={{
				duration: 250,
				delay: 0
			}}
		>
			Уточним информацию?
		</h1>

		<p
			class="text-sm text-muted-foreground"
			in:fade={{
				delay: 250,
				duration: 300
			}}
			out:fade={{
				duration: 250,
				delay: 50
			}}
		>
			Эти данные помогут автоматически записывать вас на лабораторные
		</p>

		<div
			in:fade={{
				delay: 250,
				duration: 300
			}}
			out:fade={{
				duration: 250,
				delay: 50
			}}
		>
			<form onsubmit={submitDetails}>
				<Set class="py-2">
					<div
						in:fade={{
							delay: 300,
							duration: 250
						}}
						out:fade={{
							duration: 200,
							delay: 100
						}}
					></div>

					<div
						in:fade={{
							delay: 350,
							duration: 300
						}}
						out:fade={{
							duration: 250,
							delay: 150
						}}
					>
						<Group class="mb-6">
							<Field>
								<Label class="text-sm font-medium text-left">
									Имя <span class="text-primary">*</span>
								</Label>
								<Input
									type="text"
									placeholder="Иван"
									bind:value={name}
									disabled={isSubmitting}
									class="py-5"
								/>
							</Field>

							<Field>
								<Label class="text-sm font-medium text-left">
									Фамилия <span class="text-primary">*</span>
								</Label>
								<Input
									type="text"
									placeholder="Иванов"
									bind:value={surname}
									disabled={isSubmitting}
									class="py-5"
								/>
							</Field>

							<Field>
								<Label class="text-sm font-medium text-left">
									Отчество <span class="text-primary">*</span>
								</Label>
								<Input
									type="text"
									placeholder="Иванович"
									bind:value={patronymic}
									disabled={isSubmitting}
									class="py-5"
								/>
							</Field>

							<Field>
								<Label class="text-sm font-medium text-left">
									Группа <span class="text-primary">*</span>
								</Label>
								<Input
									type="text"
									placeholder="ИН-24-8"
									bind:value={group_code}
									disabled={isSubmitting}
									class="py-5"
								/>
							</Field>

							<Field>
								<Label class="text-sm font-medium text-left">
									Телефон <span class="text-primary">*</span>
								</Label>
								<Input
									type="tel"
									placeholder="+7 978 621 48 21"
									bind:value={phone_number}
									disabled={isSubmitting}
									class="py-5"
								/>
							</Field>
						</Group>
					</div>
				</Set>

				<div
					class="space-y-3"
					in:scale={{
						delay: 650,
						duration: 300,
						start: 0.9
					}}
					out:scale={{
						duration: 250,
						start: 0.9,
						delay: 250
					}}
				>
					<Button
						type="submit"
						class="w-full py-5 font-semibold text-sm uppercase tracking-wide"
						disabled={isSubmitting}
					>
						{#if isSubmitting}
							<span class="flex items-center gap-2">
								<span class="animate-spin">⏳</span>
								Сохраняем...
							</span>
						{:else}
							Продолжить
						{/if}
					</Button>

					<Button
						type="button"
						variant="outline"
						class="w-full py-5 font-semibold text-sm uppercase tracking-wide"
						onclick={skip}
						disabled={isSubmitting}
					>
						Пропустить
					</Button>
				</div>
			</form>
		</div>
	</div>
</div>
