<script lang="ts">
	import { SubscriptionCard, SubscriptionForm } from '$lib/components/subscription';
	import { buttonVariants } from '$lib/components/ui/button';
	import { Content, Description, Header, Root, Title, Trigger } from '$lib/components/ui/dialog';
	import { cn } from '$lib/utils.ts';
	import { fade, fly } from 'svelte/transition';
	import { backOut } from 'svelte/easing';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import {
		createSubscriptionRequestSchema,
		type SubscriptionResponseArray
	} from '$lib/api/schema/subscription.ts';

	import { api } from '$lib/api/client';
	import { handleApiError, toast } from '$lib/utils/toast';
	import { onMount } from 'svelte';

	let { data } = $props();

	let subscriptions = $state<SubscriptionResponseArray>([]);
	let isLoadingSubscriptions = $state<boolean>(false);
	let isDialogOpen = $state(false);

	const formResult = superForm(data.form, {
		validators: zod4Client(createSubscriptionRequestSchema),
		onUpdate: ({ form }) => {
			console.log(form);
			if (form.valid && form.message?.success) {
				isDialogOpen = false;
				loadSubscriptions();
			} else {
				toast.error('Произошла ошибка при создании подписки');
			}
		}
	});

	const { form, errors, enhance, constraints, delayed } = formResult;

	async function loadSubscriptions() {
		isLoadingSubscriptions = true;
		try {
			subscriptions = await api.getSubscriptions();
		} catch (error) {
			console.error('Failed to load subscriptions:', error);
			handleApiError(error, 'Не удалось загрузить подписки');
		} finally {
			isLoadingSubscriptions = false;
		}
	}

	async function onSubscriptionDeleted(uuid: string) {
		toast.success('Подписка отменена', {
			description: 'Система больше не будет проверять слоты для этой работы'
		});
		subscriptions = subscriptions.filter((sub) => sub.uuid !== uuid);
	}

	async function onSubscriptionPaused(uuid: string) {
		toast.info('Подписка поставлена на паузу', {
			description: 'Вы можете возобновить её в любой момент'
		});
		subscriptions = subscriptions.map((sub) =>
			sub.uuid === uuid ? { ...sub, status: 'Paused' as const } : sub
		);
	}

	async function onSubscriptionResumed(uuid: string) {
		toast.success('Подписка возобновлена', {
			description: 'Система снова проверяет слоты'
		});
		subscriptions = subscriptions.map((sub) =>
			sub.uuid === uuid ? { ...sub, status: 'Active' as const } : sub
		);
	}

	onMount(() => {
		if (subscriptions.length === 0) {
			loadSubscriptions();
		}
	});
</script>

<div
	class="flex flex-col items-center w-full px-8 py-8"
	in:fly={{ y: 20, duration: 400, easing: backOut, opacity: 0 }}
	out:fly={{ y: 20, duration: 300, easing: backOut, opacity: 0 }}
>
	<div class="w-full">
		<div class="flex justify-between items-center">
			<span class="text-muted-foreground">
				<span class="text-primary font-bold">
					{#if isLoadingSubscriptions}
						<span class="animate-pulse">—</span>
					{:else}
						{subscriptions.length}
					{/if}
				</span>
				ПОДПИСОК
			</span>

			<Root bind:open={isDialogOpen}>
				<Trigger class={cn(buttonVariants({ variant: 'default' }), 'px-12')}>СОЗДАТЬ</Trigger>
				<Content class="max-w-lg overflow-y-scroll max-h-screen">
					<Header class="text-left">
						<Title>Новая подписка</Title>
						<Description>Настройте параметры отслеживания лабораторной работы</Description>
					</Header>

					<SubscriptionForm formApi={formResult} isSubmitting={false} />
				</Content>
			</Root>
		</div>

		<hr class="w-full my-6" />

		<div class="flex flex-col items-center gap-12">
			{#if isLoadingSubscriptions}
				<div class="w-full space-y-12">
					{#each [1, 2] as i}
						<div
							class="bg-card rounded-2xl w-full h-96 animate-pulse"
							in:fade={{ delay: i * 100, duration: 300 }}
						></div>
					{/each}
				</div>
			{:else if subscriptions.length === 0}
				<div class="text-center py-12" in:fade={{ duration: 300 }}>
					<p class="text-lg text-muted-foreground mb-2">У вас пока нет активных подписок</p>
					<p class="text-sm text-muted-foreground">
						Создайте первую подписку чтобы начать автоматический поиск слотов
					</p>
				</div>
			{:else}
				{#each subscriptions as subscription, index (subscription.uuid)}
					<div
						class="w-full"
						in:fly={{ y: 20, duration: 300, delay: 50 + index * 50, opacity: 0 }}
						out:fly={{ y: 20, duration: 250, delay: index * 30, opacity: 0 }}
					>
						<SubscriptionCard
							{subscription}
							onDeleted={onSubscriptionDeleted}
							onPaused={onSubscriptionPaused}
							onResumed={onSubscriptionResumed}
						/>
					</div>
				{/each}
			{/if}
		</div>
	</div>
</div>
