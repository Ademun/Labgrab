<script lang="ts">
	import { cn, formatTimeString } from '$lib/utils.js';
	import type { Booking } from '$lib/api/schema/booking.js';
	import type { LabTopic, LabType } from '$lib/api/schema/subscription.js';

	let {
		booking,
		labTypes,
		labTopics
	}: {
		booking: Booking;
		labTypes: LabType[];
		labTopics: LabTopic[];
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
				<span class="text-sm font-semibold">
					{booking.auditorium}
				</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Место
				</span>
				<span class="text-sm font-semibold">
					{booking.spot ?? '—'}
				</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Пара
				</span>
				<span class="text-sm font-semibold">
					{booking.lesson}
				</span>
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
	</div>
</div>
