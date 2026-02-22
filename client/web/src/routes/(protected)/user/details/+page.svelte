<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { toast } from 'svelte-sonner';
	import { getErrorMessage } from '$lib/utils/toast-errors.js';
	import { DetailsForm } from '$lib/components/user/index.js';
	import { ArrowLeft } from '@lucide/svelte';
	import { goto } from '$app/navigation';

	let { data } = $props();

	let isSubmitting = $state(false);

	const form = superForm(data.form, {
		onSubmit: () => {
			isSubmitting = true;
		},
		onResult: ({ result }) => {
			if (result.type === 'failure' && result.data?.error) {
				toast.error(result.data.error as string);
			}
			isSubmitting = false;
		}
	});
</script>

<div class="flex h-full w-full flex-col">
	<header
		class="sticky top-0 z-10 flex items-center border-b border-border bg-background px-4 py-4"
	>
		<button
			type="button"
			class="flex items-center gap-2 text-muted-foreground transition-colors hover:text-foreground"
			onclick={() => goto('/user')}
		>
			<ArrowLeft class="h-5 w-5" />
			<span class="text-sm font-medium">Назад</span>
		</button>
		<h1 class="absolute left-1/2 -translate-x-1/2 text-base font-semibold">
			Редактировать профиль
		</h1>
	</header>

	<div class="flex flex-1 flex-col items-center px-6 py-8">
		<div class="w-full max-w-sm">
			<div class="rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
				<DetailsForm {form} {isSubmitting} />
			</div>
		</div>
	</div>
</div>
