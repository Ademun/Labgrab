<script lang="ts">
	import { Field, Group, Label, Legend, Set } from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Content, Item, Root, Trigger } from '$lib/components/ui/select';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils.ts';
	import { fade } from 'svelte/transition';

	import { findLabType, labTopics, labTypes } from '$lib/stores/config';
	import type { Infer, SuperForm } from 'sveltekit-superforms';
	import type { createSubscriptionRequestSchema } from '$lib/api/schema/subscription.ts';

	let {
		formApi,
		isSubmitting
	}: { formApi: SuperForm<Infer<typeof createSubscriptionRequestSchema>>; isSubmitting: boolean } =
		$props();
	const { form, errors, constraints, enhance, message } = formApi;

	$effect(() => {
		if (!$form.lab_type) {
			$form.lab_type = 'Performance';
		}
	});

	const needsAuditorium = $derived(() => {
		const type = findLabType($form.lab_type);
		return type?.needs_auditorium ?? false;
	});

	const isPerformance = $derived($form.lab_type === 'Performance');

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
		if (!needsAuditorium() && $form.lab_auditorium !== undefined) {
			$form.lab_auditorium = undefined;
		}
	});
</script>

<form method="POST" use:enhance class="space-y-6">
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
							variant={$form.lab_type === type.id ? 'default' : 'outline'}
							onclick={() => ($form.lab_type = type.id)}
							disabled={isSubmitting}
						>
							{type.name_ru}
						</Button>
					{/each}
				</div>
				{#if $errors.lab_type}
					<span class="text-destructive text-sm">{$errors.lab_type}</span>
				{/if}
			</Field>

			<Field>
				<Label class="text-sm font-medium mb-2">
					Тема работы <span class="text-primary">*</span>
				</Label>
				<Root required type="single" bind:value={$form.lab_topic} disabled={isSubmitting}>
					<Trigger class="w-full">
						<span>
							{#if $form.lab_topic}
								{$labTopics.find((t) => t.id === $form.lab_topic)?.name_ru || $form.lab_topic}
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
				{#if $errors.lab_topic}
					<span class="text-destructive text-sm">{$errors.lab_topic}</span>
				{/if}
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
					bind:value={$form.lab_number}
					{...$constraints.lab_number}
					disabled={isSubmitting}
					class="py-5"
				/>
				{#if $errors.lab_number}
					<span class="text-destructive text-sm">{$errors.lab_number}</span>
				{/if}
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
							bind:value={$form.lab_auditorium}
							{...$constraints.lab_auditorium}
							disabled={isSubmitting}
							class="py-5"
						/>
						{#if $errors.lab_auditorium}
							<span class="text-destructive text-sm">{$errors.lab_auditorium}</span>
						{/if}
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
					bind:checked={$form.auto_enroll}
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
					bind:checked={$form.any_date}
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
