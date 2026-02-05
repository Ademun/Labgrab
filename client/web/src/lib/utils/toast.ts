import {type ExternalToast, toast as sonnerToast} from "svelte-sonner";

const defaultOptions: ExternalToast = {
    duration: 4000,
    position: "top-center"
}

export const toast = {
    success(message: string, options?: ExternalToast): any {
        return sonnerToast.success(message, {
            ...defaultOptions,
            ...options,
        });
    },

    error(message: string, options?: ExternalToast): any {
        return sonnerToast.error(message, {
            ...defaultOptions,
            duration: 5000,
            ...options,
        });
    },

    info(message: string, options?: ExternalToast): any {
        return sonnerToast.info(message, {
            ...defaultOptions,
            ...options,
        });
    },

    warning(message: string, options?: ExternalToast): any {
        return sonnerToast.warning(message, {
            ...defaultOptions,
            ...options,
        });
    },

    loading(message: string, options?: ExternalToast): any {
        return sonnerToast.loading(message, {
            ...defaultOptions,
            duration: Infinity,
            ...options,
        });
    },

    message(message: string, options?: ExternalToast): any {
        return sonnerToast.message(message, {
            ...defaultOptions,
            ...options,
        });
    },

    dismiss(toastId?: string | number): any {
        return sonnerToast.dismiss(toastId);
    },

    promise<T>(
        promise: Promise<T>,
        messages: {
            loading: string;
            success: string | ((data: T) => string);
            error: string | ((error: any) => string);
        },
    ): any {
        return sonnerToast.promise(promise, messages);
    },
}

export function handleApiError(error: unknown, defaultMessage: string = 'Произошла ошибка'): void {
    import('$lib/api/errors').then(({ApiError, NetworkError}) => {
        if (error instanceof NetworkError) {
            sonnerToast.error('Проверьте подключение к интернету', {
                description: error.message,
            });
        } else if (error instanceof ApiError) {
            if (error.isValidationError()) {
                sonnerToast.error('Проверьте правильность заполнения полей', {
                    description: error.message,
                });
            } else if (error.isServerError()) {
                sonnerToast.error('Ошибка на сервере', {
                    description: 'Попробуйте позже или обратитесь в поддержку',
                });
            } else {
                sonnerToast.error(error.message || defaultMessage);
            }
        } else if (error instanceof Error) {
            sonnerToast.error(defaultMessage, {
                description: error.message,
            });
        } else {
            sonnerToast.error(defaultMessage);
        }
    });
}