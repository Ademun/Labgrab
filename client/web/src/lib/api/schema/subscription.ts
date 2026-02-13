import { z } from 'zod';

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
	lab_auditorium: z.optional(z.number()),
	status: StatusEnum,
	auto_enroll: z.boolean(),
	any_date: z.boolean(),
	created_at: z.date(),
	closed_at: z.optional(z.date())
});

export const subscriptionResponseArraySchema = z.array(subscriptionResponseSchema);

export type SubscriptionResponse = z.infer<typeof subscriptionResponseSchema>;

export type SubscriptionResponseArray = z.infer<typeof subscriptionResponseArraySchema>;

export const createSubscriptionRequestSchema = z.object({
	lab_type: z.string(),
	lab_topic: z.string(),
	lab_number: z.number(),
	lab_auditorium: z.optional(z.number()),
	auto_enroll: z.boolean(),
	any_date: z.boolean()
});

export type CreateSubscriptionRequest = z.infer<typeof createSubscriptionRequestSchema>;

export const updateSubscriptionRequestSchema = z.object({
	status: z.optional(StatusEnum),
	auto_enroll: z.optional(z.boolean()),
	any_date: z.optional(z.boolean())
});

export type UpdateSubscriptionRequest = z.infer<typeof updateSubscriptionRequestSchema>;
