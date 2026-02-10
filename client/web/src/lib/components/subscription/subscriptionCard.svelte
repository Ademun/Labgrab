<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils.ts';
	import { fade } from 'svelte/transition';
	import {
		Action as AlertAction,
		Cancel as AlertCancel,
		Content as AlertContent,
		Description as AlertDescription,
		Root as AlertRoot,
		Title as AlertTitle,
		Trigger as AlertTrigger
	} from '$lib/components/ui/alert-dialog';

	import { api } from '$lib/api/client';
	import { handleApiError, toast } from '$lib/utils/toast';
	import { getLabTopicName, getLabTypeName } from '$lib/stores/config';
	import type { Subscription } from '$lib/api/types.ts';
	import { Spinner } from '$lib/components/ui/spinner';
	import { formatTimeString } from '$lib/utils.js';

	let {
		subscription,
		onDeleted,
		onPaused,
		onResumed
	}: {
		subscription: Subscription;
		onDeleted?: (uuid: string) => void;
		onPaused?: (uuid: string) => void;
		onResumed?: (uuid: string) => void;
	} = $props();

	let isPausing = $state<boolean>(false);
	let isResuming = $state<boolean>(false);
	let isDeleting = $state<boolean>(false);

	const isPerformance = $derived(subscription.lab_type === 'Performance');
	const isPaused = $derived(subscription.status === 'Paused');
	const isAnyActionInProgress = $derived(isPausing || isResuming || isDeleting);

	const accentBgClasses = $derived(
		isPerformance
			? 'bg-accent-performance hover:bg-accent-performance-hover'
			: 'bg-accent-defence hover:bg-accent-defence-hover'
	);

	const accentTextClasses = $derived(
		isPerformance ? 'text-accent-performance' : 'text-accent-defence'
	);

	const headerBgClasses = $derived(isPerformance ? 'bg-accent-performance' : 'bg-accent-defence');

	async function handlePause() {
		isPausing = true;
		try {
			await api.pauseSubscription(subscription.uuid);
			onPaused?.(subscription.uuid);
		} catch (error) {
			console.error('Failed to pause subscription:', error);
			handleApiError(error, 'Не удалось приостановить подписку');
		} finally {
			isPausing = false;
		}
	}

	async function handleResume() {
		isResuming = true;
		try {
			await api.resumeSubscription(subscription.uuid);
			onResumed?.(subscription.uuid);
		} catch (error) {
			console.error('Failed to resume subscription:', error);
			handleApiError(error, 'Не удалось возобновить подписку');
		} finally {
			isResuming = false;
		}
	}

	async function handleDelete() {
		isDeleting = true;
		try {
			await api.closeSubscription(subscription.uuid);
			onDeleted?.(subscription.uuid);
		} catch (error) {
			console.error('Failed to delete subscription:', error);
			handleApiError(error, 'Не удалось удалить подписку');
		} finally {
			isDeleting = false;
		}
	}
</script>

<div
	class="bg-card rounded-2xl w-full overflow-hidden shadow-xl"
	in:fade={{ duration: 300 }}
	out:fade={{ duration: 250 }}
>
	<div
		class={cn('text-sm font-semibold flex items-center justify-between px-5 py-3', headerBgClasses)}
	>
		<div class="flex items-center gap-2">
			<span
				class={cn(
					'w-1.5 h-1.5 rounded-full',
					isPaused ? 'bg-foreground/50' : 'bg-foreground animate-pulse'
				)}
			></span>
			<span class="uppercase tracking-wide">
				{isPaused ? 'На паузе' : 'Активна'}
			</span>
		</div>
		<span>
			{formatTimeString(subscription.created_at)}
		</span>
	</div>

	<div class="px-5 py-6">
		<div class="mb-5">
			<div class="flex items-baseline gap-3 mb-3">
				<span class="text-3xl font-black leading-none">
					№{subscription.lab_number}
				</span>
				<span
					class={cn('text-lg font-black uppercase leading-none tracking-tight', accentTextClasses)}
				>
					{getLabTypeName(subscription.lab_type)}
				</span>
			</div>
			<span class="text-sm text-muted-foreground font-medium">
				{getLabTopicName(subscription.lab_topic)}
			</span>
		</div>

		<div class="flex items-start gap-12 py-4 border-t border-b border-border/50">
			<div class="flex flex-col gap-1.5">
				<span class="text-xs text-muted-foreground uppercase tracking-wider font-medium">
					Аудитория
				</span>
				<span class="text-sm font-semibold">
					{subscription.lab_auditorium ?? 'Любая'}
				</span>
			</div>
		</div>

		<div class="flex flex-col items-center gap-2.5 mt-6">
			<div class="w-full">
				{#if isPaused}
					<Button
						variant="default"
						class={cn('w-full py-5 font-semibold text-sm uppercase tracking-wide', accentBgClasses)}
						onclick={handleResume}
						disabled={isAnyActionInProgress}
					>
						{#if isResuming}
							<span class="flex items-center gap-2">Возобновление...</span>
						{:else}
							Возобновить
						{/if}
					</Button>
				{:else}
					<Button
						variant="outline"
						class="w-full py-5 font-semibold text-sm uppercase tracking-wide hover:bg-accent"
						onclick={handlePause}
						disabled={isAnyActionInProgress}
					>
						{#if isPausing}
							<span class="flex items-center gap-2">Пауза...</span>
						{:else}
							Пауза
						{/if}
					</Button>
				{/if}
			</div>

			<div class="w-full">
				<Button
					class={cn('w-full py-5 font-semibold text-sm uppercase tracking-wide', accentBgClasses)}
					disabled={isAnyActionInProgress}
				>
					Настроить
				</Button>
			</div>

			<div class="w-full">
				<AlertRoot>
					<AlertTrigger class="w-full">
						<Button
							variant="outline"
							class="w-full py-5 font-semibold text-sm uppercase tracking-wide"
							disabled={isAnyActionInProgress}
						>
							{#if isDeleting}
								<Spinner />
							{:else}
								Отменить
							{/if}
						</Button>
					</AlertTrigger>
					<AlertContent>
						<AlertTitle>Вы уверены?</AlertTitle>
						<AlertDescription>
							Система перестанет проверять слоты для этой лабораторной работы. Это действие нельзя
							отменить.
						</AlertDescription>
						<div class="flex justify-end gap-3 mt-4">
							<AlertCancel>Отмена</AlertCancel>
							<AlertAction onclick={handleDelete}>Да, отменить подписку</AlertAction>
						</div>
					</AlertContent>
				</AlertRoot>
			</div>
		</div>
	</div>
</div>
