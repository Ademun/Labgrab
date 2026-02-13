import { ApiError, NetworkError } from '$lib/api/errors.ts';
import { goto } from '$app/navigation';
import type { AppConfig } from '$lib/api/schema/app.ts';
import type { UpdateUserRequest, UserResponse } from '$lib/api/schema/user.ts';
import type { AuthRequest } from '$lib/api/schema/auth.ts';
import type {
	CreateSubscriptionRequest,
	SubscriptionResponse,
	SubscriptionResponseArray,
	UpdateSubscriptionRequest
} from '$lib/api/schema/subscription.ts';
import { redirect } from '@sveltejs/kit';

interface ApiClientConfig {
	baseUrl?: string;
	timeout?: number;
}

class ApiClient {
	private readonly baseUrl: string;
	private readonly timeout: number;

	constructor(config: ApiClientConfig = {}) {
		this.baseUrl = config.baseUrl || 'http://127.0.0.1:8080';
		this.timeout = config.timeout || 10000;
	}

	private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
		const controller = new AbortController();
		const timeoutId = setTimeout(() => controller.abort(), this.timeout);

		try {
			const url = `${this.baseUrl}${endpoint}`;
			console.log(url);

			const response = await fetch(url, {
				...options,
				signal: controller.signal,
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json',
					...options.headers
				}
			});

			clearTimeout(timeoutId);

			let data: any;
			const contentType = response.headers.get('content-type');

			if (contentType && contentType.includes('application/json')) {
				data = await response.json();
			} else {
				data = await response.text();
			}

			if (!response.ok) {
				const error = new ApiError(response.status, data);

				if (error.isUnauthorized()) {
					throw redirect(303, '/auth');
				}

				throw error;
			}

			return data as T;
		} catch (error) {
			clearTimeout(timeoutId);

			if (error instanceof Error && error.name === 'AbortError') {
				throw new NetworkError('Превышено время ожидания ответа от сервера');
			}

			if (error instanceof TypeError) {
				throw new NetworkError('Не удалось обработать данные');
			}

			if (error instanceof ApiError || error instanceof NetworkError) {
				throw error;
			}

			throw new NetworkError(
				error instanceof Error ? error.message : 'Произошла неизвестная ошибка'
			);
		}
	}

	async getConfig(): Promise<AppConfig> {
		return this.request<AppConfig>('/api/web/config');
	}

	async getCurrentUser(): Promise<UserResponse> {
		return this.request<UserResponse>('/api/users');
	}

	async updateUser(data: UpdateUserRequest): Promise<UserResponse> {
		return this.request<UserResponse>('/api/users', {
			method: 'PATCH',
			body: JSON.stringify(data)
		});
	}

	async authenticateWithTelegram(authData: AuthRequest): Promise<void> {
		await this.request<void>('/api/users/auth', {
			method: 'POST',
			body: JSON.stringify(authData)
		});
	}

	async logout(): Promise<void> {
		try {
			await this.request<void>('/api/users/logout', {
				method: 'POST'
			});
		} finally {
			await goto('/', { replaceState: true });
		}
	}

	async getSubscriptions(): Promise<SubscriptionResponseArray> {
		return this.request<SubscriptionResponseArray>('/api/subscriptions');
	}

	async getSubscription(uuid: string): Promise<SubscriptionResponse> {
		return this.request<SubscriptionResponse>(`/api/subscriptions/${uuid}`);
	}

	async createSubscription(data: CreateSubscriptionRequest): Promise<SubscriptionResponse> {
		return await this.request<SubscriptionResponse>('/api/subscriptions', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async updateSubscription(uuid: string, data: UpdateSubscriptionRequest): Promise<void> {
		return this.request<void>(`/api/subscriptions/${uuid}`, {
			method: 'PATCH',
			body: JSON.stringify(data)
		});
	}

	async closeSubscription(uuid: string): Promise<void> {
		return this.updateSubscription(uuid, { status: 'Closed' });
	}

	async pauseSubscription(uuid: string): Promise<void> {
		return this.updateSubscription(uuid, { status: 'Paused' });
	}

	async resumeSubscription(uuid: string): Promise<void> {
		return this.updateSubscription(uuid, { status: 'Active' });
	}
}

export const api = new ApiClient();

export { ApiClient };
