import type {LayoutLoad} from "../../../.svelte-kit/types/src/routes/(protected)/$types";
import {redirect} from "@sveltejs/kit";

export const load: LayoutLoad = async ({fetch}) => {
    const response  = await fetch("/api/users", {
        credentials: "include",
    })

    if (response.status === 401) {
        redirect(303, "/auth")
    }

    return {}
}