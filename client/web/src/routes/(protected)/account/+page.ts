import type { PageLoad } from '../../../../.svelte-kit/types/src/routes/(protected)/account/$types';
import type { UserResponse } from '$lib/api/schema/user.ts';

export const load: PageLoad = async ({ fetch }) => {
	const res = await fetch('/api/users');
	if (!res.ok) {
		console.error('Failed to load user');
	}
	const body = await res.json();
	console.log(body);
	return body as UserResponse;
};
