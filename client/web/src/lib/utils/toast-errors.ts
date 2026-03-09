import {
	NetworkError,
	AuthError,
	ForbiddenError,
	NotFoundError,
	ConflictError,
	UnprocessableError,
	ServerError,
	ValidationError
} from '$lib/api/errors.js';

/**
 * Maps a thrown error to a user-facing Russian string suitable for toast.error().
 * Covers all typed API apperr + fallback for unknown throws.
 */
export function getErrorMessage(error: unknown): string {
	if (error instanceof NetworkError) return 'Сервер недоступен. Проверьте подключение к сети.';
	if (error instanceof AuthError) return 'Сессия истекла. Пожалуйста, войдите снова.';
	if (error instanceof ForbiddenError) return 'Доступ запрещён.';
	if (error instanceof NotFoundError) return 'Запрашиваемый ресурс не найден.';
	if (error instanceof ConflictError) return 'Конфликт данных. Попробуйте ещё раз.';
	if (error instanceof UnprocessableError) return 'Неверный формат данных запроса.';
	if (error instanceof ServerError) return 'Ошибка на сервере. Попробуйте позже.';
	if (error instanceof ValidationError) return 'Ошибка обработки ответа. Обратитесь в поддержку.';
	return 'Произошла неизвестная ошибка.';
}
