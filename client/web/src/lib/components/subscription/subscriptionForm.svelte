<script lang="ts">
	import { Field, Group, Label, Legend, Set } from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Content, Item, Root, Trigger } from '$lib/components/ui/select';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils.ts';
	import { fade } from 'svelte/transition';
	import type { CreateSubscriptionRequest } from '$lib/api/types';

	import { api } from '$lib/api/client';
	import { toast } from '$lib/utils/toast';
	import { findLabType, labTopics, labTypes } from '$lib/stores/config';

	let {
		open = $bindable(false),
		onCreated
	}: {
		open: boolean;
		onCreated?: () => void | Promise<void>;
	} = $props();

	let labType = $state<string>('Performance');
	let labTopic = $state<string | undefined>();
	let labNum = $state<number | undefined>();
	let labAuditorium = $state<number | undefined>();
	let autoSign = $state<boolean>(false);
	let anyDate = $state<boolean>(false);

	let isSubmitting = $state<boolean>(false);

	const needsAuditorium = $derived(() => {
		const type = findLabType(labType);
		return type?.needs_auditorium ?? false;
	});

	const isPerformance = $derived(labType === 'Performance');

	const accentBgClasses = $derived(
		isPerformance
			? 'bg-accent-performance hover:bg-accent-performance-hover'
			: 'bg-accent-defence hover:bg-accent-defence-hover'
	);

	const accentCheckboxClasses = $derived(
		isPerformance
			? 'data-[state=checked]:bg-accent-performance data-[state=checked]:border-accent-performance'
			: 'data-[state=checked]:bg-accent-defence data-[state=checked]:border-accent-defence'
	);

	$effect(() => {
		if (!needsAuditorium()) {
			labAuditorium = undefined;
		}
	});

	function validateForm(): boolean {
		if (!labTopic) {
			toast.error('Выберите тему лабораторной работы');
			return false;
		}

		if (labNum === undefined || labNum < 1) {
			toast.error('Укажите корректный номер работы', {
				description: 'Номер должен быть больше нуля'
			});
			return false;
		}

		if (needsAuditorium() && !labAuditorium) {
			toast.error('Укажите аудиторию', {
				description: 'Для работы типа "Выполнение" аудитория обязательна'
			});
			return false;
		}

		if (labAuditorium !== undefined && labAuditorium < 1) {
			toast.error('Укажите корректный номер аудитории', {
				description: 'Номер должен быть больше нуля'
			});
			return false;
		}

		return true;
	}

	const createSubscription = async (event: Event) => {
		event.preventDefault();

		if (!validateForm()) {
			return;
		}

		const subscriptionData: CreateSubscriptionRequest = {
			lab_type: labType,
			lab_topic: labTopic!,
			lab_number: labNum!,
			lab_auditorium: labAuditorium,
			auto_enroll: autoSign,
			any_date: anyDate,
			created_at: Date.now()
		};

		isSubmitting = true;

		try {
			await toast.promise(api.createSubscription(subscriptionData), {
				loading: 'Создаём подписку...',
				success: 'Подписка создана успешно!',
				error: (err) => {
					if (err && typeof err === 'object' && 'message' in err) {
						return err.message;
					}
					return 'Не удалось создать подписку';
				}
			});

			resetForm();
			open = false;
			await onCreated?.();
		} catch (error) {
			console.error('Failed to create subscription:', error);
		} finally {
			isSubmitting = false;
		}
	};

	function resetForm() {
		labType = 'Performance';
		labTopic = undefined;
		labNum = undefined;
		labAuditorium = undefined;
		autoSign = false;
		anyDate = false;
	}
</script>

<div>
	<form onsubmit={createSubscription} class="space-y-6">
		<Set class="bg-muted/30 rounded-xl p-6">
			<Group>
				<Legend>
					<span class="text-muted-foreground text-sm font-semibold uppercase tracking-wide">
						Информация о работе
					</span>
				</Legend>

				<Field>
					<Label class="text-sm font-medium mb-2">
						Тип работы <span class="text-primary">*</span>
					</Label>
					<div class="flex items-center gap-3">
						{#each $labTypes as type}
							<Button
								type="button"
								class={cn(
									'flex-1 py-5 font-semibold text-sm uppercase tracking-wide',
									accentBgClasses
								)}
								variant={labType === type.id ? 'default' : 'outline'}
								onclick={() => (labType = type.id)}
								disabled={isSubmitting}
							>
								{type.name_ru}
							</Button>
						{/each}
					</div>
				</Field>

				<Field>
					<Label class="text-sm font-medium mb-2">
						Тема работы <span class="text-primary">*</span>
					</Label>
					<Root required type="single" bind:value={labTopic} disabled={isSubmitting}>
						<Trigger class="w-full">
							<span>
								{#if labTopic}
									{$labTopics.find((t) => t.id === labTopic)?.name_ru || labTopic}
								{:else}
									Выберите тему из списка
								{/if}
							</span>
						</Trigger>
						<Content>
							{#each $labTopics as topic}
								<Item value={topic.id}>{topic.name_ru}</Item>
							{/each}
						</Content>
					</Root>
				</Field>

				<Field>
					<Label class="text-sm font-medium mb-2">
						Номер работы <span class="text-primary">*</span>
					</Label>
					<Input
						required
						type="number"
						min="1"
						placeholder="Например: 3"
						bind:value={labNum}
						disabled={isSubmitting}
						class="py-5"
					/>
				</Field>

				{#if needsAuditorium()}
					<div transition:fade={{ duration: 200 }}>
						<Field>
							<Label class="text-sm font-medium mb-2">
								Аудитория <span class="text-primary">*</span>
							</Label>
							<Input
								required
								type="number"
								min="1"
								placeholder="205"
								bind:value={labAuditorium}
								disabled={isSubmitting}
								class="py-5"
							/>
						</Field>
					</div>
				{/if}
			</Group>
		</Set>

		<Set class="bg-muted/30 rounded-xl p-5">
			<Group class="space-y-4">
				<Legend>
					<span class="text-muted-foreground text-sm font-semibold uppercase tracking-wide">
						Настройки
					</span>
				</Legend>
				<Field orientation="horizontal" class="flex items-start gap-3">
					<Checkbox
						id="subscription-auto"
						class={cn('mt-0.5', accentCheckboxClasses)}
						bind:checked={autoSign}
						disabled={isSubmitting}
					/>
					<Label for="subscription-auto" class="cursor-pointer">
						<div class="flex flex-col gap-1">
							<span class="text-sm font-medium">Автоматическая запись</span>
							<span class="text-xs text-muted-foreground leading-relaxed">
								Система автоматически запишет вас когда появится свободное место
							</span>
						</div>
					</Label>
				</Field>

				<Field orientation="horizontal" class="flex items-start gap-3">
					<Checkbox
						id="subscription-any-date"
						class={cn('mt-0.5', accentCheckboxClasses)}
						bind:checked={anyDate}
						disabled={isSubmitting}
					/>
					<Label for="subscription-any-date" class="cursor-pointer">
						<div class="flex flex-col gap-1">
							<span class="text-sm font-medium">Записаться на любую дату</span>
							<span class="text-xs text-muted-foreground leading-relaxed">
								Не учитывать предпочтения по датам, взять первое доступное место
							</span>
						</div>
					</Label>
				</Field>
			</Group>
		</Set>

		<Field>
			<Button
				type="submit"
				class={cn('w-full py-5 font-semibold text-sm uppercase tracking-wide', accentBgClasses)}
				disabled={isSubmitting}
			>
				{#if isSubmitting}
					<span class="flex items-center gap-2"> Создаём... </span>
				{:else}
					Создать подписку
				{/if}
			</Button>
		</Field>
	</form>
</div>
