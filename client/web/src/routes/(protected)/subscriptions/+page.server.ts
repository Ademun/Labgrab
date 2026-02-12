import { superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { createSubscriptionRequestSchema } from '$lib/api/schema/subscription.ts';
import { api } from '$lib/api/client.ts';
import { fail } from '@sveltejs/kit';

export const load = async () => {
	const form = await superValidate(zod4(createSubscriptionRequestSchema));
	return { form };
};

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(createSubscriptionRequestSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.createSubscription(form.data);
			return { form, success: true };
		} catch (error) {
			console.error('Failed to create subscription:', error);
			return fail(500, {
				form,
				error: error instanceof Error ? error.message : 'Не удалось создать подписку'
			});
		}
	}
};
