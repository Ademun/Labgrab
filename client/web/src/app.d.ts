import type { AppConfig } from '$lib/api/schema/app.js';
import type { UserResponse } from '$lib/api/schema/user.js';

declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		interface PageData {
			config: AppConfig; // from root +layout.server.ts, available everywhere
			user?: UserResponse; // from (protected) +layout.server.ts, available in protected routes
		}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
