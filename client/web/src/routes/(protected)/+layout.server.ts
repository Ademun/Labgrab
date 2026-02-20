import type { LayoutServerLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';
import { api } from '$lib/api/client.js';
import { AuthError } from '$lib/api/errors.js';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const sessionCookie = cookies.get('session_id');

	if (!sessionCookie) {
		throw redirect(303, '/auth');
	}

	try {
		const user = await api.getUser(fetch);
		return { user };
	} catch (e) {
		if (e instanceof AuthError) {
			throw redirect(303, '/auth');
		}
		throw e;
	}
};
