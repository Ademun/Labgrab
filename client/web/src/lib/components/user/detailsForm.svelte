<script lang="ts">
	import { type SuperForm } from 'sveltekit-superforms';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import type { UpdateUserFormData } from '$lib/api/schema/user.js';

	let {
		form,
		isSubmitting
	}: {
		form: SuperForm<UpdateUserFormData>;
		isSubmitting: boolean;
	} = $props();

	// svelte-ignore state_referenced_locally
	let { form: formData, enhance } = form;
</script>

<form method="POST" action="?/updateDetails" use:enhance>
	<fieldset class="my-4 space-y-6">
		<legend class="text-md my-4 text-left tracking-wide text-muted-foreground uppercase">
			Информация об аккаунте
		</legend>

		<Form.Field {form} name="name">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label>
						<div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Имя
						</div>
					</Form.Label>
					<Input
						{...props}
						disabled={isSubmitting}
						placeholder="Иван"
						bind:value={$formData.name}
					/>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>

		<Form.Field {form} name="surname">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label>
						<div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Фамилия
						</div>
					</Form.Label>
					<Input
						{...props}
						disabled={isSubmitting}
						placeholder="Иванов"
						bind:value={$formData.surname}
					/>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>

		<Form.Field {form} name="patronymic">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label>
						<div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Отчество
						</div>
					</Form.Label>
					<Input
						{...props}
						disabled={isSubmitting}
						placeholder="Иванович"
						bind:value={$formData.patronymic}
					/>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>

		<Form.Field {form} name="group_code">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label>
						<div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Группа
						</div>
					</Form.Label>
					<Input
						{...props}
						disabled={isSubmitting}
						placeholder="ИН-24-8"
						bind:value={$formData.group_code}
					/>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>

		<Form.Field {form} name="phone_number">
			<Form.Control>
				{#snippet children({ props })}
					<Form.Label>
						<div class="text-md flex items-center gap-1">
							<span class="text-xl text-red-500">*</span>Номер телефона
						</div>
					</Form.Label>
					<Input
						{...props}
						disabled={isSubmitting}
						type="number"
						placeholder="+7 999 999 99 99"
						bind:value={$formData.phone_number}
					/>
				{/snippet}
			</Form.Control>
			<Form.FieldErrors />
		</Form.Field>
	</fieldset>

	<Form.Button disabled={isSubmitting} class="text-md w-full py-6 uppercase">Сохранить</Form.Button>
</form>
