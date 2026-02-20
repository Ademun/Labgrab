import { DayOfWeekEnum, type DayOfWeek } from '$lib/api/schema/subscription.js';

export type Preferences = Map<number, Map<DayOfWeek, number[]>>;

export function prefsToSet(prefs: Preferences): Set<string> {
	const set = new Set<string>();
	for (const [week, days] of prefs)
		for (const [day, lessons] of days)
			for (const lesson of lessons) {
				set.add(`${week}-${day}-${lesson}`);
			}
	return set;
}

export function setToPrefs(set: Set<string>): Preferences {
	const prefs: Preferences = new Map();
	for (const key of set) {
		const [week, day, lesson] = key.split('-');
		const w = Number(week);

		const parsedDay = DayOfWeekEnum.safeParse(day);
		if (!parsedDay.success) continue;

		if (!prefs.has(w)) prefs.set(w, new Map());
		const days = prefs.get(w)!;
		if (!days.has(parsedDay.data)) days.set(parsedDay.data, []);
		days.get(parsedDay.data)!.push(Number(lesson));
	}
	return prefs;
}

export function prefsToJson(prefs: Preferences): Record<number, Record<string, number[]>> {
	const obj: Record<number, Record<string, number[]>> = {};
	for (const [week, days] of prefs) {
		obj[week] = {};
		for (const [day, lessons] of days) obj[week][day] = lessons;
	}
	return obj;
}

export function hashPrefs(prefs: Preferences): string {
	const obj: Record<string, Record<string, number[]>> = {};
	const weeks = [...prefs.keys()].sort((a, b) => a - b);
	for (const week of weeks) {
		obj[week] = {};
		const days = [...prefs.get(week)!.entries()].sort(([a], [b]) => a.localeCompare(b));
		for (const [day, lessons] of days) obj[week][day] = [...lessons].sort((a, b) => a - b);
	}
	return JSON.stringify(obj);
}
