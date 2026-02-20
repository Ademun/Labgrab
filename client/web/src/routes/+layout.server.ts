import type { LayoutServerLoad } from './$types.js';
import { api } from '$lib/api/client.js';

export const load: LayoutServerLoad = async ({ fetch }) => {
	const config = await api.getConfig(fetch);
	return { config };
};
