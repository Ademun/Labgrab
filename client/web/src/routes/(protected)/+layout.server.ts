import type {LayoutLoad} from "./$types.ts"
import {user} from "$lib/stores/user.svelte.js";

export const load: LayoutLoad = async ({data}) => {
    if (!user.isLoggedIn) {
        throw redirect()
    }
}