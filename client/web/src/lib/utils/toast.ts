import {type ExternalToast} from "svelte-sonner";

const defaultOptions: ExternalToast = {
    duration: 4000,
    position: "top-center"
}

export const toast = {
    success(message: string, options?: ExternalToast): any {
        return toast.success(message, {
            ...defaultOptions,
            ...options,
        });
    },

    error(message: string, options?: ExternalToast): any {
        return toast.error(message, {
            ...defaultOptions,
            duration: 5000,
            ...options,
        });
    },

    info(message: string, options?: ExternalToast): any {
        return toast.info(message, {
            ...defaultOptions,
            ...options,
        });
    },

    warning(message: string, options?: ExternalToast): any {
        return toast.warning(message, {
            ...defaultOptions,
            ...options,
        });
    },

    loading(message: string, options?: ExternalToast): any {
        return toast.loading(message, {
            ...defaultOptions,
            duration: Infinity,
            ...options,
        });
    },

    message(message: string, options?: ExternalToast): any {
        return toast.message(message, {
            ...defaultOptions,
            ...options,
        });
    },

    dismiss(toastId?: string | number): any {
        return toast.dismiss(toastId);
    },

    promise<T>(
        promise: Promise<T>,
        messages: {
            loading: string;
            success: string | ((data: T) => string);
            error: string | ((error: any) => string);
        },
        options?: ExternalToast
    ): any {
        return toast.promise(promise, messages, {
            ...defaultOptions,
            ...options,
        });
    },
}

export function handleApiError(error: unknown, defaultMessage: string = 'Произошла ошибка'): void {
    import('$lib/api/errors').then(({ApiError, NetworkError}) => {
        if (error instanceof NetworkError) {
            toast.error('Проверьте подключение к интернету', {
                description: error.message,
            });
        } else if (error instanceof ApiError) {
            if (error.isValidationError()) {
                toast.error('Проверьте правильность заполнения полей', {
                    description: error.message,
                });
            } else if (error.isServerError()) {
                toast.error('Ошибка на сервере', {
                    description: 'Попробуйте позже или обратитесь в поддержку',
                });
            } else {
                toast.error(error.message || defaultMessage);
            }
        } else if (error instanceof Error) {
            toast.error(defaultMessage, {
                description: error.message,
            });
        } else {
            toast.error(defaultMessage);
        }
    });
}