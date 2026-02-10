import { twMerge } from 'tailwind-merge';
import { type ClassValue, clsx } from 'clsx';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function formatTimeString(timeStr: string, locale: string = 'ru-RU'): string {
	const date = new Date(timeStr);

	if (isNaN(date.getTime())) {
		throw new Error('Invalid date string');
	}

	return date.toLocaleString(locale, { day: 'numeric', month: 'short', year: 'numeric' });
}
