import { z } from 'zod';

export const DayOfWeekEnum = z.enum(['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN']);
export type DayOfWeek = z.infer<typeof DayOfWeekEnum>;

export const labTypeSchema = z.object({
	id: z.string(),
	name_ru: z.string(),
	name_en: z.string(),
	needs_auditorium: z.boolean()
});

export type LabType = z.infer<typeof labTypeSchema>;

export const labTopicSchema = z.object({
	id: z.string(),
	name_ru: z.string(),
	name_en: z.string()
});

export type LabTopic = z.infer<typeof labTopicSchema>;

export const StatusEnum = z.enum(['Active', 'Paused', 'Closed']);

export const subscriptionResponseSchema = z.object({
	uuid: z.string(),
	lab_type: z.string(),
	lab_topic: z.string(),
	lab_number: z.number(),
	lab_auditorium: z.nullish(z.number()),
	status: StatusEnum,
	auto_enroll: z.boolean(),
	any_date: z.boolean(),
	created_at: z.string(),
	closed_at: z.nullish(z.string())
});

export const subscriptionResponseArraySchema = z.array(subscriptionResponseSchema);

export type SubscriptionResponse = z.infer<typeof subscriptionResponseSchema>;

export type SubscriptionResponseArray = z.infer<typeof subscriptionResponseArraySchema>;

export const editSubscriptionRequestSchema = z.object({
	uuid: z.string(),
	status: z.optional(StatusEnum),
	auto_enroll: z.optional(z.boolean()),
	any_date: z.optional(z.boolean())
});

export type EditSubscriptionRequest = z.infer<typeof editSubscriptionRequestSchema>;

export const editSubscriptionResponseSchema = z.object({
	uuid: z.string()
});

export type EditSubscriptionResponse = z.infer<typeof editSubscriptionResponseSchema>;

export const createSubscriptionRequestSchema = z.object({
	lab_type: z.string().nonempty(),
	lab_topic: z.string().nonempty(),
	lab_number: z.number().positive().max(100),
	lab_auditorium: z.optional(z.number().positive().max(1000)),
	auto_enroll: z.boolean().default(false),
	any_date: z.boolean().default(false)
});

export type CreateSubscriptionRequest = z.infer<typeof createSubscriptionRequestSchema>;

export const createSubscriptionResponseSchema = z.object({
	uuid: z.string()
});

export type CreateSubscriptionResponse = z.infer<typeof createSubscriptionResponseSchema>;

export const timeRestrictionsResponseSchema = z.object({
	restrictions: z.preprocess(
		(val) => {
			if (val instanceof Map) return val;
			return new Map(
				Object.entries(val as Record<string, unknown>).map(([week, days]) => [
					Number(week),
					new Map(Object.entries(days as Record<string, number[]>))
				])
			);
		},
		z.map(z.number(), z.map(DayOfWeekEnum, z.array(z.number())))
	)
});

export type TimeRestrcitionsResponse = z.infer<typeof timeRestrictionsResponseSchema>;

export const setTimeRestrictionsRequestSchema = z.object({
	restrictions: z.record(z.coerce.number(), z.record(DayOfWeekEnum, z.array(z.number())))
});

export type SetTimeRestrictionsRequest = z.infer<typeof setTimeRestrictionsRequestSchema>;

export const teacherPrefencesSchema = z.object({
	preferences: z.array(z.string())
});

export type TeacherPreferences = z.infer<typeof teacherPrefencesSchema>;
