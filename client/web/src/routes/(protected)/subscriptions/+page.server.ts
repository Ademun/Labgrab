import type { PageServerLoad } from './$types.js';
import { api } from '$lib/api/client.js';
import { zod4 } from 'sveltekit-superforms/adapters';
import { fail, message, superValidate } from 'sveltekit-superforms';
import { editSubscriptionRequestSchema } from '$lib/api/schema/subscription.js';
import {
	AuthError,
	ConflictError,
	NotFoundError,
	NetworkError,
	ValidationError
} from '$lib/api/errors.js';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ parent, fetch }) => {
	const [subs, editForm] = await Promise.all([
		api.getSubscriptions(fetch),
		superValidate(zod4(editSubscriptionRequestSchema))
	]);

	let bookings = undefined;
	const { user } = await parent();
	if (user.api_ready) {
		bookings = await api.getBookings(fetch);
	}

	return { subs, bookings, editForm };
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
	}
};
