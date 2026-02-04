import {derived, writable} from 'svelte/store';
import {api} from '$lib/api/client';
import type {User, UserUpdateRequest} from '$lib/api/types';

interface UserState {
    data: User | null;
    loading: boolean;
    error: Error | null;
}

const initialState: UserState = {
    data: null,
    loading: false,
    error: null,
};

function createUserStore() {
    const {subscribe, set, update} = writable<UserState>(initialState);

    return {
        subscribe,

        setUser(user: User): void {
            set({
                data: user,
                loading: false,
                error: null,
            });
        },

        async updateProfile(updates: UserUpdateRequest): Promise<void> {
            update(state => ({
                ...state,
                loading: true,
                error: null,
            }));

            try {
                const updatedUser = await api.updateUser(updates);

                set({
                    data: updatedUser,
                    loading: false,
                    error: null,
                });
            } catch (error) {
                update(state => ({
                    ...state,
                    loading: false,
                    error: error instanceof Error ? error : new Error('Failed to update profile'),
                }));

                throw error;
            }
        },

        async logout(): Promise<void> {
            try {
                await api.logout();
            } finally {
                // Очищаем состояние в любом случае
                set(initialState);
            }
        },

        reset(): void {
            set(initialState);
        },

        clearError(): void {
            update(state => ({
                ...state,
                error: null,
            }));
        },
    };
}

export const userStore = createUserStore();

export const isAuthenticated = derived<typeof userStore, boolean>(
    userStore,
    $user => $user.data !== null
);

export const fullName = derived<typeof userStore, string>(
    userStore,
    $user => {
        if (!$user.data) return '';

        const parts = [
            $user.data.surname,
            $user.data.name,
            $user.data.patronymic,
        ].filter(Boolean); // Убираем undefined/null значения

        return parts.join(' ');
    }
);

export const shortName = derived<typeof userStore, string>(
    userStore,
    $user => {
        if (!$user.data) return '';

        const parts = [
            $user.data.name,
            $user.data.surname,
        ].filter(Boolean);

        return parts.join(' ');
    }
);

export const userLoading = derived<typeof userStore, boolean>(
    userStore,
    $user => $user.loading
);

export const userError = derived<typeof userStore, Error | null>(
    userStore,
    $user => $user.error
);