import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function formatTimeString(dateStr: string, locale: string = 'ru-RU'): string {
	const date = new Date(dateStr);
	return date.toLocaleString(locale, { day: 'numeric', month: 'short', year: 'numeric' });
}

export function formatShortDayName(name: string): string {
	switch (name) {
		case 'MON':
			return 'Понедельник';
		case 'TUE':
			return 'Вторник';
		case 'WED':
			return 'Среда';
		case 'THU':
			return 'Четверг';
		case 'FRI':
			return 'Пятница';
		case 'SAT':
			return 'Суббота';
		case 'SUN':
			return 'Воскресенье';
		default:
			return name;
	}
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
