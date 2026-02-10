import { error, redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from '../../../.svelte-kit/types/src/routes/(protected)/$types';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const sessionCookie = cookies.get('session_id');

	if (!sessionCookie) {
		throw redirect(303, '/auth');
	}

	try {
		const response = await fetch('/api/users', {
			credentials: 'include'
		});

		if (response.status === 401) {
			throw redirect(303, '/auth');
		}

		if (!response.ok) {
			console.error('Failed to load user:', response.status, await response.text());
			throw error(500, 'Не удалось загрузить данные пользователя');
		}

		const user = await response.json();

		return {
			user
		};
	} catch (err) {
		if (err instanceof Response && err.status >= 300 && err.status < 400) {
			throw err;
		}

		console.error('Unexpected error in protected layout:', err);

		if (import.meta.env.DEV) {
			throw error(500, {
				message: 'Ошибка при загрузке данных'
			});
		}

		throw redirect(303, '/auth');
	}
};
