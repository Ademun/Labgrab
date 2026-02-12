import { z } from 'zod';

export const userResponseSchema = z.object({
	username: z.string(),
	name: z.optional(z.string()),
	surname: z.optional(z.string()),
	patronymic: z.optional(z.string()),
	group_code: z.optional(z.string()),
	phone_number: z.optional(z.number()),
	photo_url: z.optional(z.string())
});

export type UserResponse = z.infer<typeof userResponseSchema>;

export const updateUserRequestSchema = z.object({
	name: z.optional(z.string()),
	surname: z.optional(z.string()),
	patronymic: z.optional(z.string()),
	group_code: z.optional(z.string()),
	phone_number: z.optional(z.string()),
	dikidi_password: z.optional(z.string())
});

export type UpdateUserRequest = z.infer<typeof updateUserRequestSchema>;
