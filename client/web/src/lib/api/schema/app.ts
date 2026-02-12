import { z } from 'zod';
import { labTopicSchema, labTypeSchema } from '$lib/api/schema/subscription.ts';

export const appConfigSchema = z.object({
	lab_types: z.array(labTypeSchema),
	lab_topics: z.array(labTopicSchema)
});

export type AppConfig = z.infer<typeof appConfigSchema>;

export const apiErrorSchema = z.object({
	error: z.string(),
	message: z.string(),
	details: z.record(z.string(), z.array(z.string()))
});

export type ApiError = z.infer<typeof apiErrorSchema>;
