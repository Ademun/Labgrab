import { z } from 'zod';
import { labTopicSchema, labTypeSchema } from '$lib/api/schema/subscription.js';

export const lessonSchema = z.object({
	number: z.number(),
	start_time: z.string(),
	end_time: z.string()
})

export type Lesson = z.infer<typeof lessonSchema>

export const lessonArraySchema = z.array(lessonSchema)

export type LessonArray = z.infer<typeof lessonArraySchema>

export const appConfigSchema = z.object({
	lab_types: z.array(labTypeSchema),
	lab_topics: z.array(labTopicSchema),
	lessons: lessonArraySchema,
});

export type AppConfig = z.infer<typeof appConfigSchema>;

export const apiErrorSchema = z.object({
	error: z.string(),
	message: z.string(),
	details: z.record(z.string(), z.array(z.string()))
});

export type ApiError = z.infer<typeof apiErrorSchema>;
