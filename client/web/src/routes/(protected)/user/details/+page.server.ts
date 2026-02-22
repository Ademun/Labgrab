import type { PageServerLoad } from './$types.js';
import { fail, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { updateUserFormSchema } from '$lib/api/schema/user.js';
import { api } from '$lib/api/client.js';
import { AuthError, NetworkError, ValidationError } from '$lib/api/errors.js';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ parent }) => {
	const { user } = await parent();

	const form = await superValidate(
		{
			name: user?.name ?? '',
			surname: user?.surname ?? '',
			patronymic: user?.patronymic ?? '',
			group_code: user?.group_code ?? '',
			phone_number: user?.phone_number ?? undefined
		},
		zod4(updateUserFormSchema)
	);

	return { form };
};

export const actions = {
	updateDetails: async ({ fetch, request }) => {
		const form = await superValidate(request, zod4(updateUserFormSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await api.updateUser(
				{
					name: form.data.name,
					surname: form.data.surname,
					patronymic: form.data.patronymic,
					group_code: form.data.group_code,
					phone_number: form.data.phone_number
				},
				fetch
			);
		} catch (e) {
			if (e instanceof AuthError) {
				throw redirect(303, '/auth');
			}
			if (e instanceof ValidationError) {
				return fail(422, { form, error: 'Неверный формат данных.' });
			}
			if (e instanceof NetworkError) {
				return fail(503, { form, error: 'Сервер недоступен. Попробуйте позже.' });
			}
			return fail(500, { form, error: 'Внутренняя ошибка сервера.' });
		}

		throw redirect(303, '/user');
	}
};
