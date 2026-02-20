import type { PageServerLoad } from './$types.js';
import { api } from '$lib/api/client.js';
import { zod4 } from 'sveltekit-superforms/adapters';
import { fail, message, superValidate } from 'sveltekit-superforms';
import {
	createSubscriptionRequestSchema,
	editSubscriptionRequestSchema
} from '$lib/api/schema/subscription.js';
import {
	AuthError,
	ConflictError,
	NotFoundError,
	NetworkError,
	ValidationError
} from '$lib/api/errors.js';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch }) => {
	const subs = await api.getSubscriptions(fetch);
	const editForm = await superValidate(zod4(editSubscriptionRequestSchema));
	const createForm = await superValidate(zod4(createSubscriptionRequestSchema));
	return { subs, editForm, createForm };
};

export const actions = {
	editSubscription: async ({ fetch, request }) => {
		const form = await superValidate(request, zod4(editSubscriptionRequestSchema));
		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.editSubscription(form.data, fetch);
		} catch (e) {
			if (e instanceof AuthError) {
				throw redirect(303, '/auth');
			}
			if (e instanceof NotFoundError) {
				return fail(404, { form, error: 'Подписка не найдена.' });
			}
			if (e instanceof ConflictError) {
				return fail(409, { form, error: e.body?.message ?? 'Конфликт данных.' });
			}
			if (e instanceof ValidationError) {
				return fail(422, { form, error: 'Неверный формат данных.' });
			}
			if (e instanceof NetworkError) {
				return fail(503, { form, error: 'Сервер недоступен. Попробуйте позже.' });
			}
			return fail(500, { form, error: 'Внутренняя ошибка сервера.' });
		}

		return message(form, 'Подписка обновлена');
	},

	createSubscription: async ({ fetch, request }) => {
		const form = await superValidate(request, zod4(createSubscriptionRequestSchema));
		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.createSubscription(form.data, fetch);
		} catch (e) {
			if (e instanceof AuthError) {
				redirect(303, '/auth');
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

		return message(form, 'Подписка создана');
	}
};
