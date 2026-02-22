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

export const updateUserFormSchema = z.object({
	name: z.string().min(1, 'Обязательное поле'),
	surname: z.string().min(1, 'Обязательное поле'),
	patronymic: z.string().min(1, 'Обязательное поле'),
	group_code: z.string().min(1, 'Обязательное поле'),
	phone_number: z.number().int('Обязательное поле').positive()
});

export type UpdateUserFormData = z.infer<typeof updateUserFormSchema>;

export type UpdateUserRequest = {
	name: string;
	surname: string;
	patronymic: string;
	group_code: string;
	phone_number: number;
};
