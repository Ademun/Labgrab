<script lang="ts">
	import { Card, CreateForm } from '$lib/components/subscription/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import {
		type SubscriptionResponse,
		type SubscriptionResponseArray
	} from '$lib/api/schema/subscription.js';
	import { superForm } from 'sveltekit-superforms';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { api } from '$lib/api/client.js';
	import { invalidateAll } from '$app/navigation';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { cn } from '$lib/utils.js';
	import { tick } from 'svelte';

	let { data } = $props();
	let subscriptions = $state<SubscriptionResponseArray>(data.subs);
	let isLoading = $state(false);
	let isEditDialogOpen = $state(false);
	let currentEditingUuid = $state<string | null>(null);
	let isCreateDialogOpen = $state(false);
	let isCreateFormSubmitting = $state(false);

	$effect(() => {
		subscriptions = data.subs;
	});

	const {
		form: editFormData,
		errors: editErrors,
		enhance: editEnhance
	} = superForm(data.editForm, {
		onUpdated: async ({ form }) => {
			if (form.valid && form.message) {
				isEditDialogOpen = false;

				if (currentEditingUuid) {
					try {
						const updated = await api.getSubscription(currentEditingUuid);
						const index = subscriptions.findIndex((sub) => sub.uuid === currentEditingUuid);
						if (index !== -1 && updated.length > 0) {
							subscriptions = [
								...subscriptions.slice(0, index),
								updated[0],
								...subscriptions.slice(index + 1)
							];
						}
					} catch (e) {
						console.error('Failed to reload subscription:', e);
					}
				}

				currentEditingUuid = null;
			}
		}
	});

	const createForm = superForm(data.createForm, {
		onSubmit: () => {
			isCreateFormSubmitting = true;
		},
		onResult: async ({ result }) => {
			if (result.type === 'success') {
				isCreateDialogOpen = false;
				await invalidateAll();
			}
			isCreateFormSubmitting = false;
		}
	});

	async function onPaused(uuid: string) {
		try {
			await api.pauseSubscription(uuid);
			const updated = await api.getSubscription(uuid);
			const index = subscriptions.findIndex((sub) => sub.uuid === uuid);
			if (index !== -1 && updated.length > 0) {
				subscriptions = [
					...subscriptions.slice(0, index),
					updated[0],
					...subscriptions.slice(index + 1)
				];
			}
		} catch (e) {
			console.error('Failed to pause subscription:', e);
		}
	}

	async function onRestored(uuid: string) {
		try {
			await api.restoreSubscription(uuid);
			const updated = await api.getSubscription(uuid);
			const index = subscriptions.findIndex((sub) => sub.uuid === uuid);
			if (index !== -1 && updated.length > 0) {
				subscriptions = [
					...subscriptions.slice(0, index),
					updated[0],
					...subscriptions.slice(index + 1)
				];
			}
		} catch (e) {
			console.error('Failed to restore subscription:', e);
		}
	}

	async function onClosed(uuid: string) {
		try {
			await api.closeSubscription(uuid);
			subscriptions = subscriptions.filter((sub) => sub.uuid !== uuid);
		} catch (e) {
			console.error('Failed to close subscription:', e);
		}
	}

	function onEditDialogOpen(sub: SubscriptionResponse) {
		isEditDialogOpen = true;
		currentEditingUuid = sub.uuid;
		$editFormData = {
			uuid: sub.uuid,
			auto_enroll: sub.auto_enroll,
			any_date: sub.any_date
		};
	}
</script>

<div class="flex w-full flex-col items-center px-8 py-8">
	<div class="w-full">
		<div class="flex items-center justify-between">
			<span class="text-muted-foreground">
				<span class="font-bold text-primary">
					{#if isLoading}
						<Spinner />
					{:else}
						{subscriptions.length}
					{/if}
				</span>
				ПОДПИСОК
			</span>
			<Dialog.Root bind:open={isCreateDialogOpen}>
				<Dialog.Trigger
					class={cn(
						buttonVariants({ variant: 'default' }),
						'text-md px-10 py-5 font-semibold tracking-wide uppercase'
					)}
				>
					Создать
				</Dialog.Trigger>
				<Dialog.Content>
					<Dialog.Header>
						<Dialog.Title>Новая подписка</Dialog.Title>
						<CreateForm form={createForm} isSubmitting={isCreateFormSubmitting} />
					</Dialog.Header>
				</Dialog.Content>
			</Dialog.Root>
		</div>

		<hr class="my-6 w-full" />

		<div class="flex flex-col items-center gap-12">
			{#if isLoading}
				<Spinner />
			{:else if subscriptions.length === 0}
				<div class="py-12 text-center">
					<p class="mb-2 text-lg text-muted-foreground">У вас пока нет активных подписок</p>
					<p class="text-sm text-muted-foreground">
						Создайте первую подписку чтобы начать автоматический поиск слотов
					</p>
				</div>
			{:else}
				{#each subscriptions as subscription (subscription.uuid)}
					<div class="w-full">
						<Card
							{subscription}
							bind:isEditDialogOpen
							{onEditDialogOpen}
							{onPaused}
							{onRestored}
							{onClosed}
							{editFormData}
							{editErrors}
							{editEnhance}
						/>
					</div>
				{/each}
			{/if}
		</div>
	</div>
</div>
