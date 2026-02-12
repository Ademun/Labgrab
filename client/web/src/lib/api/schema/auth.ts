import { z } from 'zod';

export const authRequestSchema = z.object({
	id: z.number(),
	first_name: z.string(),
	last_name: z.string(),
	username: z.string(),
	photo_url: z.string(),
	auth_date: z.number(),
	hash: z.string()
});

export type AuthRequest = z.infer<typeof authRequestSchema>;
