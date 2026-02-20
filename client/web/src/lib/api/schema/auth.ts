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
