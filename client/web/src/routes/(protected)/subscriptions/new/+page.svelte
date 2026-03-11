<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { toast } from 'svelte-sonner';
	import { CreateForm } from '$lib/components/subscription/index.js';
	import { Header } from '$lib/components/navigation/index.js';

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

	const labTypes = $derived(data.config.lab_types);
	const labTopics = $derived(data.config.lab_topics);
	const apiAuthed = $derived(data.apiAuthed);
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Новая подписка" back="/subscriptions" />

	<div class="flex flex-1 flex-col items-center px-6 py-8">
		<div class="w-full max-w-sm">
			<div class="rounded-2xl border border-border/40 bg-card px-6 py-6 shadow-xl">
				<CreateForm {form} {isSubmitting} {labTypes} {labTopics} {apiAuthed} />
			</div>
		</div>
	</div>
</div>