import { api } from '$lib/api/client.js';
import type { PageServerLoad } from './$types.js';

export const load: PageServerLoad = async ({ fetch }) => {
	const schedule = await api.getTimePreferences(fetch);
	const preferences = schedule.preferences;
	return { preferences };
};
