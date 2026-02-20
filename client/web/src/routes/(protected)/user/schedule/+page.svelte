<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { api } from '$lib/api/client.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import { formatShortDayName } from '$lib/utils.js';
	import { hashPrefs, prefsToJson, prefsToSet, setToPrefs } from '$lib/utils/schedule.js';
	import { getErrorMessage } from '$lib/utils/toast-errors.js';
	import { toast } from 'svelte-sonner';

	let { data } = $props();

	const lessons = $derived(data.config.lessons);

	const dayOfWeeks = ['MON', 'TUE', 'WED', 'THU', 'FRI'];
	const weekNumbers = [1, 2];

	const initialHash = $derived(hashPrefs(data.preferences));
	let selected = $state(prefsToSet(data.preferences));

	const isChanged = $derived(hashPrefs(setToPrefs(selected)) !== initialHash);

	const slotKey = (week: number, day: string, lesson: number) => `${week}-${day}-${lesson}`;

	function toggle(key: string) {
		selected.has(key) ? selected.delete(key) : selected.add(key);
		selected = new Set(selected);
	}

	function selectAll(week: number) {
		for (const day of dayOfWeeks)
			for (const lesson of lessons) {
				selected.add(slotKey(week, day, lesson.number));
			}
		selected = new Set(selected);
	}

	function clearAll(week: number) {
		for (const day of dayOfWeeks)
			for (const lesson of lessons) {
				selected.delete(slotKey(week, day, lesson.number));
			}
		selected = new Set(selected);
	}

	const countSelected = (week: number) =>
		[...selected].filter((k) => k.startsWith(`${week}-`)).length;

	async function save() {
		const payload = prefsToJson(setToPrefs(selected));
		try {
			await api.setTimePreferences({ preferences: payload });
			await invalidateAll();
		} catch (e) {
			toast.error(getErrorMessage(e));
		}
	}
</script>

<div class="flex w-full flex-col items-center px-6 py-8">
	<Tabs.Root value="1" class="flex w-full flex-col gap-4">
		<div class="sticky top-0 flex w-full flex-col items-center gap-6 rounded-xl bg-card px-6 py-4">
			<div class="flex w-full items-center justify-between">
				<Tabs.List>
					{#each weekNumbers as week}
						<Tabs.Trigger value={String(week)}>Неделя {week === 1 ? 'I' : 'II'}</Tabs.Trigger>
					{/each}
				</Tabs.List>
				<Button
					variant={isChanged ? 'default' : 'outline'}
					disabled={!isChanged}
					onclick={save}
					class="px-8">Сохранить</Button
				>
			</div>
		</div>

		{#each weekNumbers as week}
			<Tabs.Content value={String(week)}>
				<div class="mb-4 flex w-full gap-4">
					<Button class="flex-1" variant="outline" onclick={() => selectAll(week)}>
						Выбрать все
					</Button>
					<Button class="flex-1" variant="destructive" onclick={() => clearAll(week)}>
						Очистить все ({countSelected(week)})
					</Button>
				</div>

				<div class="flex flex-col gap-6">
					{#each dayOfWeeks as day}
						<div class="w-full rounded-xl bg-card px-6 py-4">
							<span class="font-bold uppercase">{formatShortDayName(day)}</span>
							<div class="mt-4 grid auto-rows-max grid-cols-2 gap-4">
								{#each lessons as lesson}
									{@const key = slotKey(week, day, lesson.number)}
									<Button
										variant={selected.has(key) ? 'default' : 'outline'}
										class="min-h-max py-1"
										onclick={() => toggle(key)}
									>
										<div class="flex flex-col items-center">
											<span class="text-lg font-semibold">{lesson.number} пара</span>
											<span class="text-muted-foreground">
												{lesson.start_time} - {lesson.end_time}
											</span>
										</div>
									</Button>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			</Tabs.Content>
		{/each}
	</Tabs.Root>
</div>
