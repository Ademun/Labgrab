import type {PageServerLoad} from "./$types.js"
import {api} from "$lib/api/client.js";
import {zod4} from "sveltekit-superforms/adapters";
import {fail, message, superValidate} from "sveltekit-superforms";
import {createSubscriptionRequestSchema, editSubscriptionRequestSchema} from "$lib/api/schema/subscription.js";

export const load: PageServerLoad = async ({fetch}) => {
    const subs = await api.getSubscriptions(fetch)
    const editForm = await superValidate(zod4(editSubscriptionRequestSchema))
    const createForm = await superValidate(zod4(createSubscriptionRequestSchema))
    return {subs, editForm, createForm}
}

export const actions = {
    editSubscription: async ({fetch, request}) => {
        const form = await superValidate(request, zod4(editSubscriptionRequestSchema))
        if (!form.valid) {
            console.log(form.errors)
            return fail(400, {form})
        }

        try {
            await api.editSubscription(form.data, fetch)
        } catch (e) {
            return fail(500, {form})
        }

        return message(form, "Form posted successfully")
    },
    createSubscription: async ({fetch, request}) => {
        const form = await superValidate(request, zod4(createSubscriptionRequestSchema))
        if (!form.valid) {
            console.log(form.errors)
            return fail(400, {form})
        }

        try {
            await api.createSubscription(form.data, fetch)
        } catch (e) {
            console.log(e)
            return fail(500, {form})
        }

        return message(form, "Form created successfully")
    }
}