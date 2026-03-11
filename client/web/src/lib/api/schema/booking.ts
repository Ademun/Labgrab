import { z } from 'zod';

export const bookingStatusEnum = z.enum(['Active', 'Closed']);
export type BookingStatus = z.infer<typeof bookingStatusEnum>;

export const bookingSchema = z.object({
	id: z.number(),
	type: z.string(),
	topic: z.string(),
	number: z.number(),
	auditorium: z.number(),
	spot: z.number().nullable(),
	lesson: z.number(),
	start_time: z.string(),
	end_time: z.string(),
	status: bookingStatusEnum
});

export const bookingArraySchema = z.array(bookingSchema);

export type Booking = z.infer<typeof bookingSchema>;
export type BookingArray = z.infer<typeof bookingArraySchema>;
