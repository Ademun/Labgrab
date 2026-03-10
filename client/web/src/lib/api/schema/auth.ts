import { z } from 'zod';

export const authRequestSchema = z.object({
	id: z.optional(z.number()),
	first_name: z.optional(z.string()),
	last_name: z.optional(z.string()),
	username: z.optional(z.string()),
	photo_url: z.optional(z.string()),
	auth_date: z.optional(z.number()),
	hash: z.optional(z.string())
});

export type AuthRequest = z.infer<typeof authRequestSchema>;

export const createUserDataRequestSchema = z.object({
	dikidi_password: z.string(),
	dikidi_phone_number: z.string()
});

export type CreateUserDataRequest = z.infer<typeof createUserDataRequestSchema>;

export const getUserInfoResponseSchema = z.object({
	phone_number: z.string(),
	password: z.string(),
	api_authed: z.boolean(),
	last_auth: z.nullish(z.string()),
})

export type GetUserInfoResponse = z.infer<typeof getUserInfoResponseSchema>