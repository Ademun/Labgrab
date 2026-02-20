import { type AuthRequest } from '$lib/api/schema/auth.js';
import { type UserResponse, userResponseSchema } from '$lib/api/schema/user.js';
import {
	type CreateSubscriptionRequest,
	type CreateSubscriptionResponse,
	createSubscriptionResponseSchema,
	type EditSubscriptionRequest,
	type EditSubscriptionResponse,
	editSubscriptionResponseSchema,
	type SetTimePreferencesRequest,
	type SubscriptionResponseArray,
	subscriptionResponseArraySchema,
	teacherPrefencesSchema,
	type TeacherPreferences,
	type TimePreferencesResponse,
	timePreferencesResponseSchema
} from '$lib/api/schema/subscription.js';
import { type AppConfig, appConfigSchema } from '$lib/api/schema/app.js';
import { z } from 'zod';

interface ApiClientConfig {
	baseUrl: string;
	timeout: number;
}

class ApiClient {
	private readonly baseUrl: string;
	private readonly timeout: number;

	constructor(
		config: ApiClientConfig = {
			baseUrl: 'http://127.0.0.1:8080/api',
			timeout: 10000
		}
	) {
		this.baseUrl = config.baseUrl;
		this.timeout = config.timeout;
	}

	private async request<T>(
		endpoint: string,
		schema: z.ZodSchema<T>,
		fetchFn?: typeof fetch,
		options: RequestInit = {}
	): Promise<T> {
		const controller = new AbortController();
		const timeoutId = setTimeout(() => controller.abort(), this.timeout);

		try {
			const fetcher = fetchFn ?? fetch;

			const url = `${this.baseUrl}${endpoint}`;

			const response = await fetcher(url, {
				...options,
				signal: controller.signal,
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json',
					...options.headers
				}
			});

			clearTimeout(timeoutId);

			if (!response.ok) {
				throw new Error(`API error: ${response.status}`);
			}

			const isVoid = schema instanceof z.ZodVoid;

			if (response.status === 204) {
				if (!isVoid) {
					throw new Error(`Expected body but got 204 No Content for ${endpoint}`);
				}
				return schema.parse(undefined) as T;
			}

			const data = await response.json();
			return schema.parse(data);
		} catch (error) {
			console.error(error);
			throw new Error('Unknown error');
		}
	}

	async auth(data: AuthRequest, fetchFn?: typeof fetch): Promise<boolean> {
		return this.request<boolean>('/users/auth', z.boolean(), fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async getUser(fetchFn?: typeof fetch): Promise<UserResponse> {
		return this.request<UserResponse>('/users', userResponseSchema, fetchFn);
	}

	async getSubscriptions(fetchFn?: typeof fetch): Promise<SubscriptionResponseArray> {
		return this.request<SubscriptionResponseArray>(
			'/subscriptions',
			subscriptionResponseArraySchema,
			fetchFn
		);
	}

	async getSubscription(uuid: string, fetchFn?: typeof fetch): Promise<SubscriptionResponseArray> {
		return this.request<SubscriptionResponseArray>(
			`/subscriptions/${uuid}`,
			subscriptionResponseArraySchema,
			fetchFn
		);
	}

	async editSubscription(
		data: EditSubscriptionRequest,
		fetchFn?: typeof fetch
	): Promise<EditSubscriptionResponse> {
		return this.request<EditSubscriptionResponse>(
			`/subscriptions/${data.uuid}`,
			editSubscriptionResponseSchema,
			fetchFn,
			{
				method: 'PATCH',
				body: JSON.stringify(data)
			}
		);
	}

	async pauseSubscription(uuid: string, fetchFn?: typeof fetch): Promise<EditSubscriptionResponse> {
		return this.editSubscription(
			{
				uuid: uuid,
				status: 'Paused'
			},
			fetchFn
		);
	}

	async restoreSubscription(
		uuid: string,
		fetchFn?: typeof fetch
	): Promise<EditSubscriptionResponse> {
		return this.editSubscription(
			{
				uuid: uuid,
				status: 'Active'
			},
			fetchFn
		);
	}

	async closeSubscription(uuid: string, fetchFn?: typeof fetch): Promise<EditSubscriptionResponse> {
		return this.editSubscription(
			{
				uuid: uuid,
				status: 'Closed'
			},
			fetchFn
		);
	}

	async createSubscription(
		data: CreateSubscriptionRequest,
		fetchFn?: typeof fetch
	): Promise<CreateSubscriptionResponse> {
		return this.request('/subscriptions', createSubscriptionResponseSchema, fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async getTimePreferences(fetchFn?: typeof fetch): Promise<TimePreferencesResponse> {
		return this.request('/subscriptions/preferences/time', timePreferencesResponseSchema, fetchFn);
	}

	async setTimePreferences(data: SetTimePreferencesRequest, fetchFn?: typeof fetch): Promise<void> {
		return this.request('/subscriptions/preferences/time', z.void(), fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async getTeacherPreferences(fetchFn?: typeof fetch): Promise<TeacherPreferences> {
		return this.request('/subscriptions/preferences/teachers', teacherPrefencesSchema, fetchFn);
	}

	async setTeacherPreferences(data: TeacherPreferences, fetchFn?: typeof fetch): Promise<void> {
		return this.request('/subscriptions/prefrences/teachers', z.void(), fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async getConfig(fetchFn?: typeof fetch): Promise<AppConfig> {
		return this.request<AppConfig>('/web/config', appConfigSchema, fetchFn);
	}
}

export const api = new ApiClient();

export { ApiClient };
