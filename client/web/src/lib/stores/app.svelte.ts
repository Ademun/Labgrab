import type { AppConfig } from '$lib/api/schema/app.js';
import { api } from '$lib/api/client.js';
import type { LabTopic, LabType } from '$lib/api/schema/subscription.js';

class AppStore {
	data = $state<AppConfig | undefined>(undefined);

	async load() {
		try {
			this.data = await api.getConfig();
		} catch (error) {
			throw new Error('Не удалось загрузить конфигурацию приложения');
		}
	}

	findLabType(labType: string): LabType | undefined {
		if (!this.data) {
			return undefined;
		}
		return this.data.lab_types.find((type) => type.id === labType);
	}

	findLabTopic(labTopic: string): LabTopic | undefined {
		if (!this.data) {
			return undefined;
		}
		return this.data.lab_topics.find((topic) => topic.id === labTopic);
	}
}

export const app = new AppStore();
