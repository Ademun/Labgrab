import type { User } from '$lib/types/user.ts';
import type { PageLoad } from '../../../../.svelte-kit/types/src/routes/(protected)/account/$types';

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch('/api/users');
	if (!res.ok) {
		console.error('Failed to load user');
	}
	const body = await res.json();
	console.log(body);
	return body as User;
};
