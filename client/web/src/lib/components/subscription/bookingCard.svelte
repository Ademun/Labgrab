<script lang="ts">
	import { cn, formatTimeString } from '$lib/utils.js';
	import type { Booking } from '$lib/api/schema/booking.js';
	import type { LabTopic, LabType } from '$lib/api/schema/subscription.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Ban } from '@lucide/svelte';
	import { api } from '$lib/api/client.js';
	import { toast } from 'svelte-sonner';
	import { getErrorMessage } from '$lib/utils/toast-errors.js';

	let {
		booking,
		labTypes,
		labTopics,
		onCancelled
	}: {
		booking: Booking;
		labTypes: LabType[];
		labTopics: LabTopic[];
		onCancelled?: (id: number) => void;
	} = $props();

	const isClosed = $derived(booking.status === 'Closed');
	const isPerformance = $derived(booking.type === 'Performance');

	const accentTextClasses = $derived(
		isPerformance ? 'text-accent-performance' : 'text-accent-defence'
	);
	const headerBgClasses = $derived(isPerformance ? 'bg-accent-performance' : 'bg-accent-defence');

	const resolvedTypeName = $derived(
		labTypes.find((t) => t.id === booking.type)?.name_ru ?? booking.type
	);
	const resolvedTopicName = $derived(
		labTopics.find((t) => t.id === booking.topic)?.name_ru ?? booking.topic
	);

	const formattedStart = $derived(formatTimeString(booking.start_time));

	let showCancelDialog = $state(false);
	let isCancelling = $state(false);

	async function handleCancelConfirm() {
		showCancelDialog = false;
		isCancelling = true;
		try {
			await api.cancelBooking(booking.number);
			onCancelled?.(booking.number);
		} catch (e) {
			toast.error(getErrorMessage(e));
		} finally {
			isCancelling = false;
		}
	}
</script>

<div
	class={cn(
		'w-full overflow-hidden rounded-2xl border border-border/40 bg-card shadow-xl transition-opacity',
		isClosed && 'opacity-60'
	)}
>
	<div
		class={cn(
			'flex items-center justify-between px-5 py-3 text-sm font-semibold text-white',
			headerBgClasses,
			isClosed && 'opacity-70'
		)}
	>
		<span class="tracking-wide uppercase">
			{isClosed ? 'Завершена' : 'Активна'}
		</span>
		<span>{formattedStart}</span>
	</div>

	<div class="px-5 py-6">
		<div class="mb-5">
			<div class="mb-3 flex items-baseline gap-3">
				<span class="text-3xl leading-none font-black">
					№{booking.number}
				</span>
				<span
					class={cn('text-lg leading-none font-black tracking-tight uppercase', accentTextClasses)}
				>
					{resolvedTypeName}
				</span>
			</div>
			<span class="text-sm font-medium text-muted-foreground">
				{resolvedTopicName}
			</span>
		</div>

		<div class="flex flex-wrap items-start gap-x-12 gap-y-4 border-t border-b border-border/50 py-4">
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Аудитория
				</span>
				<span class="text-sm font-semibold">{booking.auditorium}</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Место
				</span>
				<span class="text-sm font-semibold">{booking.spot ?? '—'}</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Пара
				</span>
				<span class="text-sm font-semibold">{booking.lesson}</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Время
				</span>
				<span class="text-sm font-semibold">
					{booking.start_time.slice(11, 16)} — {booking.end_time.slice(11, 16)}
				</span>
			</div>
		</div>

		{#if !isClosed}
			<div class="mt-4">
				<Button
					variant="outline"
					class="w-full py-4 text-sm font-semibold text-destructive"
					disabled={isCancelling}
					onclick={() => (showCancelDialog = true)}
				>
					<Ban class="mr-2 h-4 w-4" />
					{isCancelling ? 'Отмена...' : 'Отменить запись'}
				</Button>
			</div>
		{/if}
	</div>
</div>

<AlertDialog.Root bind:open={showCancelDialog}>
	<AlertDialog.Content class="max-w-sm rounded-3xl px-6 py-8">
		<AlertDialog.Header class="flex flex-col items-center gap-3 text-center">
			<Ban class="h-14 w-14 text-destructive" strokeWidth={1.5} />
			<AlertDialog.Title class="text-lg leading-snug font-bold">Отменить запись?</AlertDialog.Title>
			<AlertDialog.Description class="text-center text-sm leading-relaxed text-muted-foreground">
				Запись на <span class="font-medium text-foreground">№{booking.number} {resolvedTypeName}</span>
				будет отменена. Это действие нельзя отменить.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer class="mt-6 flex flex-col gap-3">
			<AlertDialog.Action
				class="w-full py-5 bg-destructive text-destructive-foreground hover:bg-destructive/90"
				onclick={handleCancelConfirm}
			>
				Отменить запись
			</AlertDialog.Action>
			<AlertDialog.Cancel class="w-full py-5">
				Назад
			</AlertDialog.Cancel>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>