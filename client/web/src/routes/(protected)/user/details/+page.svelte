<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { toast } from 'svelte-sonner';
	import { getErrorMessage } from '$lib/utils/toast-errors.js';
	import { DetailsForm } from '$lib/components/user/index.js';
	import { Header } from '$lib/components/navigation/index.js';
	import { invalidateAll } from '$app/navigation';

	let { data } = $props();

	let isSubmitting = $state(false);

	const form = superForm(data.form, {
		onSubmit: () => {
			isSubmitting = true;
		},
		onResult: async ({ result }) => {
			if (result.type === 'failure' && result.data?.error) {
				toast.error(result.data.error as string);
			}
			isSubmitting = false;
			await invalidateAll()
		}
	});
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Редактировать профиль" back="/user" />

	<div class="flex flex-1 flex-col items-center justify-center px-6 py-8">
		<div class="w-full max-w-sm">
			<div class="rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
				<DetailsForm {form} {isSubmitting} />
			</div>
		</div>
	</div>
</div>
