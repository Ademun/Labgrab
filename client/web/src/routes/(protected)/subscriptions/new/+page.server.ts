import type { PageServerLoad } from './$types.js';
import { api } from '$lib/api/client.js';
import { zod4 } from 'sveltekit-superforms/adapters';
import { fail, superValidate } from 'sveltekit-superforms';
import { createSubscriptionRequestSchema } from '$lib/api/schema/subscription.js';
import {
	AuthError,
	ConflictError,
	NetworkError,
	ValidationError
} from '$lib/api/errors.js';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async () => {
	const form = await superValidate(zod4(createSubscriptionRequestSchema));
	return { form };
};

export const actions = {
	createSubscription: async ({ fetch, request }) => {
		const form = await superValidate(request, zod4(createSubscriptionRequestSchema));
		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.createSubscription(form.data, fetch);
		} catch (e) {
			if (e instanceof AuthError) {
				throw redirect(303, '/auth');
			}
			if (e instanceof ConflictError) {
				return fail(409, { form, error: e.body?.message ?? 'Такая подписка уже существует.' });
			}
			if (e instanceof ValidationError) {
				return fail(422, { form, error: 'Неверный формат данных.' });
			}
			if (e instanceof NetworkError) {
				return fail(503, { form, error: 'Сервер недоступен. Попробуйте позже.' });
			}
			return fail(500, { form, error: 'Не удалось создать подписку.' });
		}

		throw redirect(303, '/subscriptions');
	}
};