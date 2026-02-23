<script lang="ts">
	import { Card } from '$lib/components/subscription/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import {
		type SubscriptionResponse,
		type SubscriptionResponseArray
	} from '$lib/api/schema/subscription.js';
	import { superForm } from 'sveltekit-superforms';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { api } from '$lib/api/client.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { cn } from '$lib/utils.js';
	import { toast } from 'svelte-sonner';
	import { getErrorMessage } from '$lib/utils/toast-errors.js';
	import { fly, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import { Header } from '$lib/components/navigation/index.js';

	let { data } = $props();
	let subscriptions = $state<SubscriptionResponseArray>(data.subs);
	let isLoading = $state(false);
	let isEditDialogOpen = $state(false);
	let currentEditingUuid = $state<string | null>(null);

	const labTypes = $derived(data.config.lab_types);
	const labTopics = $derived(data.config.lab_topics);

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
						toast.error(getErrorMessage(e));
					}
				}

				currentEditingUuid = null;
			}
		},
		onResult: ({ result }) => {
			if (result.type === 'failure' && result.data?.error) {
				toast.error(result.data.error as string);
			}
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
			toast.error(getErrorMessage(e));
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
			toast.error(getErrorMessage(e));
		}
	}

	async function onClosed(uuid: string) {
		try {
			await api.closeSubscription(uuid);
			subscriptions = subscriptions.filter((sub) => sub.uuid !== uuid);
		} catch (e) {
			toast.error(getErrorMessage(e));
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

<div class="flex h-full w-full flex-col">
	<Header title="Подписки" />

	<div class="flex w-full flex-col items-center px-8 py-8">
		<div class="w-full">
			<div class="flex items-center justify-between" in:fly={{ y: -10, duration: 240 }}>
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
				<a
					href="/subscriptions/new"
					class={cn(
						buttonVariants({ variant: 'default' }),
						'text-md px-10 py-5 font-semibold tracking-wide uppercase'
					)}
				>
					Создать
				</a>
			</div>

			<hr class="my-6 w-full" />

			<div class="flex flex-col items-center gap-12">
				{#if isLoading}
					<Spinner />
				{:else if subscriptions.length === 0}
					<div class="py-12 text-center" in:fade={{ duration: 280, delay: 100 }}>
						<p class="mb-2 text-lg text-muted-foreground">У вас пока нет активных подписок</p>
						<p class="text-sm text-muted-foreground">
							Создайте первую подписку чтобы начать автоматический поиск слотов
						</p>
					</div>
				{:else}
					{#each subscriptions as subscription, i (subscription.uuid)}
						<div
							class="w-full"
							in:fly={{ y: 20, duration: 300, delay: i * 40 }}
							out:fade={{ duration: 200 }}
							animate:flip={{ duration: 300 }}
						>
							<Card
								{subscription}
								{labTypes}
								{labTopics}
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
</div>
