import type { PageServerLoad } from './$types.js';
import { fail, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { createUserDataRequestSchema, updateUserDataRequestSchema } from '$lib/api/schema/auth.js';
import { api } from '$lib/api/client.js';
import { AuthError, NetworkError, ValidationError } from '$lib/api/errors.js';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ parent, fetch }) => {
	const { user } = await parent();

	try {
		const userInfo = await api.getUserInfo(fetch);

		if (userInfo) {
			const updateForm = await superValidate(
				{ dikidi_phone_number: userInfo.phone_number, dikidi_password: '' },
				zod4(updateUserDataRequestSchema)
			);
			return { user, userInfo, connectForm: null, updateForm };
		}

		const connectForm = await superValidate(zod4(createUserDataRequestSchema));
		return { user, userInfo: null, connectForm, updateForm: null };
	} catch (e) {
		if (e instanceof AuthError) throw redirect(303, '/auth');
		throw e;
	}
};

export const actions = {
	connect: async ({ fetch, request }) => {
		const form = await superValidate(request, zod4(createUserDataRequestSchema));
		if (!form.valid) return fail(400, { form });

		try {
			await api.createUserData(
				{
					dikidi_phone_number: form.data.dikidi_phone_number,
					dikidi_password: form.data.dikidi_password
				},
				fetch
			);
		} catch (e) {
			if (e instanceof AuthError) throw redirect(303, '/auth');
			if (e instanceof ValidationError) return fail(422, { form, error: 'Неверный формат данных.' });
			if (e instanceof NetworkError) return fail(503, { form, error: 'Сервер недоступен. Попробуйте позже.' });
			return fail(500, { form, error: 'Внутренняя ошибка сервера.' });
		}

		throw redirect(303, '/user/integration');
	},

	update: async ({ fetch, request }) => {
		const form = await superValidate(request, zod4(updateUserDataRequestSchema));
		if (!form.valid) return fail(400, { form });

		try {
			await api.updateUserData(
				{
					dikidi_phone_number: form.data.dikidi_phone_number,
					dikidi_password: form.data.dikidi_password
				},
				fetch
			);
		} catch (e) {
			if (e instanceof AuthError) throw redirect(303, '/auth');
			if (e instanceof ValidationError) return fail(422, { form, error: 'Неверный формат данных.' });
			if (e instanceof NetworkError) return fail(503, { form, error: 'Сервер недоступен. Попробуйте позже.' });
			return fail(500, { form, error: 'Внутренняя ошибка сервера.' });
		}

		return { form, success: true };
	}
};