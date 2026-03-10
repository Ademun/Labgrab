import { getUserInfoResponseSchema, type AuthRequest, type CreateUserDataRequest, type GetUserInfoResponse, type UpdateUserDataRequest } from '$lib/api/schema/auth.js';
import {
	type UpdateUserRequest,
	type UserResponse,
	userResponseSchema
} from '$lib/api/schema/user.js';
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
import { type AppConfig, appConfigSchema, apiErrorSchema } from '$lib/api/schema/app.js';
import { z, ZodError } from 'zod';
import { createApiError, NetworkError, ValidationError } from '$lib/api/errors.js';
import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { type BookingArray, bookingArraySchema } from '$lib/api/schema/booking.js';

interface ApiClientConfig {
	baseUrl: string;
	timeout: number;
}

class ApiClient {
	private readonly baseUrl: string;
	private readonly timeout: number;

	constructor(
		config: ApiClientConfig = {
			baseUrl: PUBLIC_API_BASE_URL,
			timeout: 10000
		}
	) {
		this.baseUrl = config.baseUrl;
		this.timeout = config.timeout;
	}

	private hasJsonBody(response: Response): boolean {
    const contentType = response.headers.get('Content-Type');
    const contentLength = response.headers.get('Content-Length');

    if (response.status === 204) return false;
    if (contentLength === '0') return false;
    if (!contentType) return false;

    return contentType.includes('application/json');
}

private async request<T>(
    endpoint: string,
    schema: z.ZodSchema<T>,
    fetchFn?: typeof fetch,
    options: RequestInit = {},
    timeout?: number
): Promise<T> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeout ?? this.timeout);

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

        if (!response.ok) {
            let errorBody;
            try {
                const raw = await response.json();
                errorBody = apiErrorSchema.parse(raw);
            } catch {}
            throw createApiError(response.status, errorBody);
        }

        if (!this.hasJsonBody(response)) {
            const acceptsUndefined = schema.safeParse(undefined).success;
            if (!acceptsUndefined) {
                throw new ValidationError(
                    `Expected JSON body but got empty response for ${endpoint} (status ${response.status})`
                );
            }
            return schema.parse(undefined) as T;
        }

        let data: unknown;
        try {
            data = await response.json();
        } catch (e) {
            throw new ValidationError(
                `Failed to parse JSON response on ${endpoint}: ${e instanceof Error ? e.message : String(e)}`
            );
        }

        try {
            return schema.parse(data);
        } catch (e) {
            if (e instanceof ZodError) {
                throw new ValidationError(`Response schema mismatch on ${endpoint}: ${e.message}`, e);
            }
            throw e;
        }
    } catch (error) {
        if (
            error instanceof NetworkError ||
            error instanceof ValidationError ||
            (error instanceof Error && error.name.endsWith('Error') && 'status' in error)
        ) {
            throw error;
        }

        if (error instanceof DOMException && error.name === 'AbortError') {
            throw new NetworkError(`Request to ${endpoint} timed out after ${this.timeout}ms`, error);
        }

        throw new NetworkError(
            `Network failure on ${endpoint}: ${error instanceof Error ? error.message : String(error)}`,
            error
        );
    } finally {
        clearTimeout(timeoutId);
    }
}

	async auth(data: AuthRequest, fetchFn?: typeof fetch): Promise<void> {
		return this.request('/auth/user', z.void(), fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async createUserData(data: CreateUserDataRequest, fetchFn?: typeof fetch): Promise<void> {
		return this.request('/auth/data', z.void(), fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async dikidiAuth(fetchFn?: typeof fetch): Promise<void> {
		return this.request('/auth/dikidi', z.void(), fetchFn, {}, 30_000);
	}

	async getUserInfo(fetchFn?: typeof fetch): Promise<GetUserInfoResponse | void> {
		return this.request('/auth/data', z.union([getUserInfoResponseSchema, z.void()]), fetchFn)
	}

	async updateUserData(data: UpdateUserDataRequest, fetchFn?: typeof fetch): Promise<void> {
		return this.request('/auth/data', z.void(), fetchFn, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	async getUser(fetchFn?: typeof fetch): Promise<UserResponse> {
		return this.request<UserResponse>('/user', userResponseSchema, fetchFn);
	}

	async updateUser(data: UpdateUserRequest, fetchFn?: typeof fetch): Promise<void> {
		return this.request('/user', z.void(), fetchFn, {
			method: 'PATCH',
			body: JSON.stringify(data)
		});
	}

	async deleteUser(fetchFn?: typeof fetch): Promise<void> {
		return this.request('/user', z.void(), fetchFn, {
			method: "DELETE"
		})
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
		return this.editSubscription({ uuid, status: 'Paused' }, fetchFn);
	}

	async restoreSubscription(
		uuid: string,
		fetchFn?: typeof fetch
	): Promise<EditSubscriptionResponse> {
		return this.editSubscription({ uuid, status: 'Active' }, fetchFn);
	}

	async closeSubscription(uuid: string, fetchFn?: typeof fetch): Promise<EditSubscriptionResponse> {
		return this.editSubscription({ uuid, status: 'Closed' }, fetchFn);
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
		return this.request('/subscriptions/preferences/teachers', z.void(), fetchFn, {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async getBookings(fetchFn?: typeof fetch): Promise<BookingArray> {
		return this.request<BookingArray>('/bookings', bookingArraySchema, fetchFn);
	}

	async cancelBooking(id: number, fetchFn?: typeof fetch): Promise<void> {
		return this.request<void>(`/bookings/${id}`, z.void(), fetchFn, {
			method: "DELETE"
		})
	}

	async getConfig(fetchFn?: typeof fetch): Promise<AppConfig> {
		return this.request<AppConfig>('/web/config', appConfigSchema, fetchFn);
	}
}

export const api = new ApiClient();

export { ApiClient };
