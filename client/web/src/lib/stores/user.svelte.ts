import type { UserResponse } from '$lib/api/schema/user.js';
import type { AuthRequest } from '$lib/api/schema/auth.js';
import { api } from '$lib/api/client.js';

class UserStore {
	data = $state<UserResponse | undefined>(undefined);

	async auth(data: AuthRequest) {
		try {
			await api.auth(data);
			await this.load();
		} catch (error) {
			throw new Error('Не удалось авторизоваться');
		}
	}

	async load() {
		try {
			this.data = await api.getUser();
		} catch (error) {
			throw new Error('Не удалось загрузить данные');
		}
	}
}

export const user = new UserStore();
