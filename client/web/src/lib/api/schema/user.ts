import { z } from 'zod';

export const userResponseSchema = z.object({
	name: z.nullish(z.string()),
	surname: z.nullish(z.string()),
	patronymic: z.nullish(z.string()),
	group_code: z.nullish(z.string()),
	phone_number: z.nullish(z.string()),
	telegram_username: z.string(),
	telegram_photo_url: z.nullish(z.string()),
	api_ready: z.boolean()
});

export type UserResponse = z.infer<typeof userResponseSchema>;

export const updateUserFormSchema = z.object({
	name: z.string().min(1, 'Обязательное поле'),
	surname: z.string().min(1, 'Обязательное поле'),
	patronymic: z.string().min(1, 'Обязательное поле'),
	group_code: z.string().min(1, 'Обязательное поле'),
	phone_number: z.string().min(1, 'Обязательное поле'),
});

export type UpdateUserFormData = z.infer<typeof updateUserFormSchema>;

export type UpdateUserRequest = {
	name: string;
	surname: string;
	patronymic: string;
	group_code: string;
	phone_number: string;
};
