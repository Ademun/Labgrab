export class ApiError extends Error {
    constructor(
        public readonly status: number,
        public readonly body: any,
        message?: string,
    ) {
        super(message || ApiError.extractMessage(body))
        this.name = 'ApiError'
    }

    private static extractMessage(body: any): string {
        if (typeof body === 'string') {
            return body
        }

        if (body && typeof body === 'object') {
            return body.message || body.error || body.detail || 'Произошла неизвестная ошибка'
        }

        return 'Произошла неизвестная ошибка'
    }

    isUnauthorized(): boolean {
        return this.status === 401
    }

    isValidationError(): boolean {
        return this.status === 400 || this.status === 402
    }

    isServerError(): boolean {
        return this.status >= 500
    }
}

export class NetworkError extends Error {
    constructor(message: string = 'Проблема с подключением к серверу') {
        super(message);
        this.name = 'NetworkError';
    }
}

export class ValidationError extends Error {
    constructor(
        message: string,
        public readonly errors: Record<string, string[]>
    ) {
        super(message);
        this.name = 'ValidationError';
    }
}