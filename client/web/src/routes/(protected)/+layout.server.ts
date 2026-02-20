import type { LayoutServerLoad } from './$types.ts';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const sessionCookie = cookies.get('session_id');

	if (!sessionCookie) {
		throw redirect(303, '/auth');
	}
};
