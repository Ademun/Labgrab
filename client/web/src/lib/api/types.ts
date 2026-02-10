export interface LabType {
	id: string;
	name_ru: string;
	name_en: string;
	needs_auditorium: boolean;
}

export interface LabTopic {
	id: string;
	name_ru: string;
	name_en: string;
}

export interface AppConfig {
	lab_types: LabType[];
	lab_topics: LabTopic[];
}

export interface User {
	username: string;
	name?: string;
	surname?: string;
	patronymic?: string;
	group_code?: string;
	phone_number?: string;
	photo_url?: string;
}

export interface UserUpdateRequest {
	name?: string;
	surname?: string;
	patronymic?: string;
	group_code?: string;
	phone_number?: string;
}

export interface TelegramAuthData {
	id: number;
	first_name: string;
	last_name: string;
	username: string;
	photo_url: string;
	auth_date: number;
	hash: string;
}

export interface Subscription {
	uuid: string;
	lab_type: string;
	lab_topic: string;
	lab_number: number;
	lab_auditorium?: number;
	created_at: string;
	closed_at?: string;
	status?: 'active' | 'paused' | 'closed';
	checks_count?: number;
}

export interface CreateSubscriptionRequest {
	lab_type: string;
	lab_topic: string;
	lab_number: number;
	lab_auditorium?: number;
	auto_enroll?: boolean;
	any_date?: boolean;
}

export interface UpdateSubscriptionRequest {
	auto_enroll?: boolean;
	any_date?: boolean;
	status?: 'active' | 'paused';
}

export interface CreateSubscriptionResponse {
	subscription: Subscription;
}

export interface ApiErrorResponse {
	error: string;
	message: string;
	details?: Record<string, string[]>;
}
