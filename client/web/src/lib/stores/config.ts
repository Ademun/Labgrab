import type {AppConfig, LabTopic, LabType} from "$lib/api/types.ts";
import {derived, get, writable} from "svelte/store";
import {api} from "$lib/api/client.ts";

interface ConfigState {
    data: AppConfig | null;
    loading: boolean;
    error: Error | null;
    lastUpdated: number | null;
}

const initialState: ConfigState = {
    data: null,
    loading: false,
    error: null,
    lastUpdated: null,
};

function createConfigStore() {
    const { subscribe, set, update } = writable<ConfigState>(initialState);

    return {
        subscribe,
        async load(force: boolean = false): Promise<void> {
            const currentState = get({ subscribe });

            if (currentState.data && !force) {
                return;
            }

            if (currentState.loading) {
                return;
            }

            update(state => ({
                ...state,
                loading: true,
                error: null,
            }));

            try {
                const data = await api.getConfig();

                set({
                    data,
                    loading: false,
                    error: null,
                    lastUpdated: Date.now(),
                });
            } catch (error) {
                update(state => ({
                    ...state,
                    loading: false,
                    error: error instanceof Error ? error : new Error('Failed to load config'),
                }));

                throw error;
            }
        },

        reset(): void {
            set(initialState);
        },
    }
}

export const configStore = createConfigStore();

export const labTypes = derived<typeof configStore, LabType[]>(
    configStore,
    $config => $config.data?.lab_types || []
);

export const labTopics = derived<typeof configStore, LabTopic[]>(
    configStore,
    $config => $config.data?.lab_topics || []
);

export const configLoading = derived<typeof configStore, boolean>(
    configStore,
    $config => $config.loading
);

export const configError = derived<typeof configStore, Error | null>(
    configStore,
    $config => $config.error
);

export function findLabType(id: string): LabType | undefined {
    const types = get(labTypes);
    return types.find(type => type.id === id);
}

export function findLabTopic(id: string): LabTopic | undefined {
    const topics = get(labTopics);
    return topics.find(topic => topic.id === id);
}

export function getLabTypeName(id: string, locale: 'ru' | 'en' = 'ru'): string {
    const type = findLabType(id);
    if (!type) return id;
    return locale === 'ru' ? type.name_ru : type.name_en;
}

export function getLabTopicName(id: string, locale: 'ru' | 'en' = 'ru'): string {
    const topic = findLabTopic(id);
    if (!topic) return id;
    return locale === 'ru' ? topic.name_ru : topic.name_en;
}