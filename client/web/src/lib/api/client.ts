import {ApiError, NetworkError} from "$lib/api/errors.ts";
import {goto} from "$app/navigation";
import type {
    AppConfig,
    CreateSubscriptionRequest, CreateSubscriptionResponse,
    Subscription,
    TelegramAuthData, UpdateSubscriptionRequest,
    User,
    UserUpdateRequest
} from "$lib/api/types.ts";

interface ApiClientConfig {
    baseUrl?: string;
    timeout?: number;
}

class ApiClient {
    private readonly baseUrl: string;
    private readonly timeout: number;

    constructor(config: ApiClientConfig = {}) {
        this.baseUrl = config.baseUrl || ''
        this.timeout = config.timeout || 10000
    }

    private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        try {
            const url = `${this.baseUrl}${endpoint}`

            const response = await fetch(endpoint, {
                ...options,
                signal: controller.signal,
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            })

            clearTimeout(timeoutId)

            let data: any
            const contentType = response.headers.get('content-type')

            if (contentType && contentType.includes('application/json')) {
                data = await response.json()
            } else {
                data = await response.text()
            }

            if (!response.ok) {
                const error = new ApiError(response.status, data)

                if (error.isUnauthorized()) {
                    await goto('/auth', { replaceState: true })
                }

                throw error
            }

            return data as T
        } catch (error) {
            clearTimeout(timeoutId)

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

    async getCurrentUser(): Promise<User> {
        return this.request<User>('/api/users');
    }

    async updateUser(data: UserUpdateRequest): Promise<User> {
        return this.request<User>('/api/users', {
            method: 'PATCH',
            body: JSON.stringify(data),
        });
    }

    async authenticateWithTelegram(authData: TelegramAuthData): Promise<void> {
        await this.request<void>('/api/users/auth', {
            method: 'POST',
            body: JSON.stringify(authData),
        });
    }

    async logout(): Promise<void> {
        try {
            await this.request<void>('/api/users/logout', {
                method: 'POST',
            });
        } finally {
            await goto('/', { replaceState: true });
        }
    }

    async getSubscriptions(): Promise<Subscription[]> {
        return this.request<Subscription[]>('/api/subscriptions');
    }

    async getSubscription(uuid: string): Promise<Subscription> {
        return this.request<Subscription>(`/api/subscriptions/${uuid}`);
    }

    async createSubscription(data: CreateSubscriptionRequest): Promise<Subscription> {
        const response = await this.request<CreateSubscriptionResponse>(
            '/api/subscriptions',
            {
                method: 'POST',
                body: JSON.stringify(data),
            }
        );
        return response.subscription;
    }

    async updateSubscription(uuid: string, data: UpdateSubscriptionRequest): Promise<Subscription> {
        return this.request<Subscription>(`/api/subscriptions/${uuid}`, {
            method: 'PATCH',
            body: JSON.stringify(data),
        });
    }

    async deleteSubscription(uuid: string): Promise<void> {
        await this.request<void>(`/api/subscriptions/${uuid}`, {
            method: 'DELETE',
        });
    }

    async pauseSubscription(uuid: string): Promise<Subscription> {
        return this.updateSubscription(uuid, { status: 'paused' });
    }

    async resumeSubscription(uuid: string): Promise<Subscription> {
        return this.updateSubscription(uuid, { status: 'active' });
    }
}

export const api = new ApiClient();

export { ApiClient };