import { api } from '$lib/api/client.js';
import type { PageServerLoad } from './$types.js';
import { AuthError } from '$lib/api/errors.js';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch }) => {
	try {
		const schedule = await api.getTimeRestrictions(fetch);
		return { restrictions: schedule.restrictions };
	} catch (e) {
		if (e instanceof AuthError) {
			throw redirect(303, '/auth');
		}
		throw e;
	}
};
