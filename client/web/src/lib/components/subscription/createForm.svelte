<script lang="ts">
	import { type SuperForm } from 'sveltekit-superforms';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import type {
		CreateSubscriptionRequest,
		LabTopic,
		LabType
	} from '$lib/api/schema/subscription.js';
	import Button from '../ui/button/button.svelte';
	import { cn } from '$lib/utils.js';
	import Checkbox from '../ui/checkbox/checkbox.svelte';
	import Label from '../ui/label/label.svelte';

	let {
		form,
		isSubmitting,
		labTypes,
		labTopics
	}: {
		form: SuperForm<CreateSubscriptionRequest>;
		isSubmitting: boolean;
		labTypes: LabType[];
		labTopics: LabTopic[];
	} = $props();

	// svelte-ignore state_referenced_locally
	let { form: formData, enhance } = form;

	const needsAuditorium = $derived.by(() => {
		const type = labTypes.find((t) => t.id === $formData.lab_type);
		return type?.needs_auditorium ?? false;
	});
	const isPerformance = $derived($formData.lab_type === 'Performance');
	const accentBgClasses = $derived(
		isPerformance
			? 'bg-accent-performance hover:bg-accent-performance-hover'
			: 'bg-accent-defence hover:bg-accent-defence-hover'
	);

	$effect(() => {
		if (!$formData.lab_type) {
			$formData.lab_type = 'Performance';
		}
	});

	$effect(() => {
		if (needsAuditorium && $formData.lab_auditorium !== undefined) {
			$formData.lab_auditorium = undefined;
		}
	});
</script>

<form method="POST" action="?/createSubscription" use:enhance>
	<fieldset class="my-4 space-y-6">
		<legend class="text-md my-4 text-left tracking-wide text-muted-foreground uppercase"
			>Информация о работе</legend
		>
		<Form.Field {form} name="lab_type">
			<Form.Control>
				{#snippet children({ props })}
					<Input {...props} type="hidden" bind:value={$formData.lab_type} />
					<div class="flex justify-between gap-8">
						{#each labTypes as type}
							<Button
								disabled={isSubmitting}
								variant={$formData.lab_type === type.id ? 'default' : 'outline'}
								class={cn(
									'text-md flex-1 py-6 tracking-wide uppercase',
									$formData.lab_type === type.id && accentBgClasses
								)}
								onclick={() => ($formData.lab_type = type.id)}>{type.name_ru}</Button
							>
						{/each}
					</div>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
		<Form.Field {form} name="lab_topic">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label>
						<div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Тема работы
						</div>
					</Form.Label>
					<Select.Root
						{...props}
						disabled={isSubmitting}
						type="single"
						bind:value={$formData.lab_topic}
					>
						<Select.Trigger class="w-full">
							<span>
								{#if $formData.lab_topic}
									{labTopics.find((t) => t.id === $formData.lab_topic)?.name_ru}
								{:else}
									Выберите тему из списка
								{/if}
							</span>
						</Select.Trigger>
						<Select.Content>
							{#each labTopics as topic}
								<Select.Item value={topic.id}>{topic.name_ru}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
		<Form.Field {form} name="lab_number">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label
						><div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Номер работы
						</div></Form.Label
					>
					<Input
						{...props}
						disabled={isSubmitting}
						type="number"
						placeholder="Например 3"
						bind:value={$formData.lab_number}
					/>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
		{#if needsAuditorium}
			<Form.Field {form} name="lab_auditorium">
				<Form.Control>
					{#snippet children({ props })}
						<Form.Label
							><div class="text-md flex items-center gap-1">
								<span class="text-xl text-red-500">*</span>Аудитория
							</div></Form.Label
						>
						<Input
							{...props}
							disabled={isSubmitting}
							type="number"
							placeholder="Например 428"
							bind:value={$formData.lab_auditorium}
						/>
					{/snippet}
				</Form.Control>
				<Form.FieldErrors />
			</Form.Field>
		{/if}
	</fieldset>
	<hr />
	<fieldset class="my-4 space-y-6">
		<legend class="text-md my-4 text-left tracking-wide text-muted-foreground uppercase"
			>Настройки
		</legend>
		<Form.Field {form} name="auto_enroll">
			<Form.Control>
				{#snippet children({ props })}
					<div class="flex items-start gap-3">
						<Checkbox {...props} disabled={isSubmitting} id="auto_enroll" class="mt-1" />
						<div class="grid gap-1">
							<Label for="auto_enroll" class="text-md">Автоматическая запись</Label>
							<p class="text-left text-muted-foreground">
								Система автоматически запишет вас когда появится свободное место
							</p>
						</div>
					</div>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
		<Form.Field {form} name="any_date">
			<Form.Control>
				{#snippet children({ props })}
					<div class="flex items-start gap-3">
						<Checkbox {...props} disabled={isSubmitting} id="any_date" class="mt-1" />
						<div class="grid gap-2">
							<Label for="any_date" class="text-md">Любая дата</Label>
							<p class="text-left text-muted-foreground">
								Не учитывать расписание, взять первое доступное место
							</p>
						</div>
					</div>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
	</fieldset>
	<Form.Button disabled={isSubmitting} class={cn('text-md w-full py-6 uppercase', accentBgClasses)}
		>Создать</Form.Button
	>
</form>
