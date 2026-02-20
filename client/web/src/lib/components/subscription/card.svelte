<script lang="ts">
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { cn, formatTimeString } from '$lib/utils.js';
	import type {
		EditSubscriptionRequest,
		LabTopic,
		LabType,
		SubscriptionResponse
	} from '$lib/api/schema/subscription.ts';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { SuperForm } from 'sveltekit-superforms';

	AlertDialog;
	Dialog;

	let {
		subscription,
		labTypes,
		labTopics,
		isEditDialogOpen = $bindable(),
		onEditDialogOpen,
		onPaused,
		onRestored,
		onClosed,
		editFormData,
		editErrors,
		editEnhance
	}: {
		subscription: SubscriptionResponse;
		labTypes: LabType[];
		labTopics: LabTopic[];
		isEditDialogOpen: boolean;
		onEditDialogOpen: (sub: SubscriptionResponse) => void;
		onPaused: (uuid: string) => void;
		onRestored: (uuid: string) => void;
		onClosed: (uuid: string) => void;
		editFormData: SuperForm<EditSubscriptionRequest>['form'];
		editErrors: SuperForm<EditSubscriptionRequest>['errors'];
		editEnhance: SuperForm<EditSubscriptionRequest>['enhance'];
	} = $props();

	const isPaused = $derived(subscription.status === 'Paused');
	const isPerformance = $derived(subscription.lab_type === 'Performance');

	const accentBgClasses = $derived(
		isPerformance
			? 'bg-accent-performance hover:bg-accent-performance-hover'
			: 'bg-accent-defence hover:bg-accent-defence-hover'
	);

	const accentTextClasses = $derived(
		isPerformance ? 'text-accent-performance' : 'text-accent-defence'
	);

	const headerBgClasses = $derived(isPerformance ? 'bg-accent-performance' : 'bg-accent-defence');

	const resolvedTypeName = $derived(
		labTypes.find((t) => t.id === subscription.lab_type)?.name_ru ?? subscription.lab_type
	);
	const resolvedTopicName = $derived(
		labTopics.find((t) => t.id === subscription.lab_topic)?.name_ru ?? subscription.lab_topic
	);
</script>

<div class="w-full overflow-hidden rounded-2xl border border-border/40 bg-card shadow-xl">
	<div
		class={cn(
			'flex items-center justify-between px-5 py-3 text-sm font-semibold text-white',
			headerBgClasses
		)}
	>
		<span class="tracking-wide uppercase">
			{isPaused ? 'На паузе' : 'Активна'}
		</span>
		<span>
			{formatTimeString(subscription.created_at)}
		</span>
	</div>

	<div class="px-5 py-6">
		<div class="mb-5">
			<div class="mb-3 flex items-baseline gap-3">
				<span class="text-3xl leading-none font-black">
					№{subscription.lab_number}
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

		<div class="flex items-start gap-12 border-t border-b border-border/50 py-4">
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Аудитория
				</span>
				<span class="text-sm font-semibold">
					{subscription.lab_auditorium ?? 'Любая'}
				</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Авто
				</span>
				<span class="text-sm font-semibold">
					{subscription.auto_enroll ? 'Да' : 'Нет'}
				</span>
			</div>
			<div class="flex flex-col gap-1.5">
				<span class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
					Любая дата
				</span>
				<span class="text-sm font-semibold">
					{subscription.any_date ? 'Да' : 'Нет'}
				</span>
			</div>
		</div>

		<div class="mt-6 flex flex-col items-center gap-2.5">
			<div class="w-full">
				{#if isPaused}
					<Button
						variant="default"
						class={cn('w-full py-5 text-sm font-semibold tracking-wide uppercase', accentBgClasses)}
						onclick={() => onRestored(subscription.uuid)}
					>
						Возобновить
					</Button>
				{:else}
					<Button
						variant="outline"
						class="w-full py-5 text-sm font-semibold tracking-wide uppercase hover:bg-accent"
						onclick={() => onPaused(subscription.uuid)}
					>
						Пауза
					</Button>
				{/if}
			</div>

			<div class="w-full">
				<Dialog.Root bind:open={isEditDialogOpen}>
					<Dialog.Trigger
						onclick={() => onEditDialogOpen(subscription)}
						class={cn(
							buttonVariants({ variant: 'default' }),
							accentBgClasses,
							'w-full py-5 text-sm font-semibold tracking-wide uppercase'
						)}
					>
						Настроить
					</Dialog.Trigger>
					<Dialog.Content>
						<Dialog.Header>
							<Dialog.Title>Настройка подписки</Dialog.Title>
						</Dialog.Header>
						<form class="mt-6" method="POST" action="?/editSubscription" use:editEnhance>
							<input type="hidden" name="uuid" value={$editFormData.uuid} />

							<div class="flex flex-col gap-6">
								<div class="flex items-start gap-3">
									<Checkbox
										id="auto-enroll"
										name="auto_enroll"
										bind:checked={$editFormData.auto_enroll}
									/>
									<Label for="auto-enroll">Автоматическая запись</Label>
								</div>
								<div class="flex items-start gap-3">
									<Checkbox id="any-date" name="any_date" bind:checked={$editFormData.any_date} />
									<div class="grid gap-2">
										<Label for="any-date">Любая дата</Label>
										<p class="text-sm text-muted-foreground">
											Первое доступное время, независимо от вашего расписания
										</p>
									</div>
								</div>
							</div>
							<Dialog.Footer>
								<Button
									type="submit"
									class={cn(
										'mt-6 w-full py-5 text-sm font-semibold tracking-wide uppercase',
										accentBgClasses
									)}
								>
									Сохранить
								</Button>
							</Dialog.Footer>
						</form>
					</Dialog.Content>
				</Dialog.Root>
			</div>

			<div class="w-full">
				<AlertDialog.Root>
					<AlertDialog.Trigger
						class={cn(
							buttonVariants({ variant: 'outline' }),
							'w-full py-5 text-sm font-semibold tracking-wide uppercase'
						)}
					>
						Отменить
					</AlertDialog.Trigger>
					<AlertDialog.Content>
						<AlertDialog.Header>
							<AlertDialog.Title>Вы уверены?</AlertDialog.Title>
							<AlertDialog.Description>
								Система перестанет отслеживать записи на эту работу. Это действие нельзя отменить.
							</AlertDialog.Description>
						</AlertDialog.Header>
						<AlertDialog.Footer>
							<AlertDialog.Cancel>Отменить</AlertDialog.Cancel>
							<AlertDialog.Action onclick={() => onClosed(subscription.uuid)}>
								Удалить
							</AlertDialog.Action>
						</AlertDialog.Footer>
					</AlertDialog.Content>
				</AlertDialog.Root>
			</div>
		</div>
	</div>
</div>
