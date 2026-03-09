<script lang="ts">
    import { Card, BookingCard } from '$lib/components/subscription/index.js';
    import { buttonVariants } from '$lib/components/ui/button/index.js';
    import {
        type SubscriptionResponse,
        type SubscriptionResponseArray
    } from '$lib/api/schema/subscription.js';
    import type { BookingArray } from '$lib/api/schema/booking.js';
    import { superForm } from 'sveltekit-superforms';
    import { Spinner } from '$lib/components/ui/spinner/index.js';
    import { api } from '$lib/api/client.js';
    import { cn } from '$lib/utils.js';
    import { toast } from 'svelte-sonner';
    import { getErrorMessage } from '$lib/utils/toast-errors.js';
    import { fly, fade } from 'svelte/transition';
    import { flip } from 'svelte/animate';
    import { Header } from '$lib/components/navigation/index.js';
    import * as Tabs from '$lib/components/ui/tabs/index.js';

    let { data } = $props();
    let subscriptions = $state<SubscriptionResponseArray>(data.subs);
    let bookings = $state<BookingArray>(data.bookings);
    let isLoading = $state(false);
    let isEditDialogOpen = $state(false);
    let currentEditingUuid = $state<string | null>(null);

    const labTypes = $derived(data.config.lab_types);
    const labTopics = $derived(data.config.lab_topics);

    $effect(() => {
        subscriptions = data.subs;
    });

    $effect(() => {
        bookings = data.bookings;
    });

    const activeBookings = $derived(bookings.filter((b) => b.status === 'Active'));
    const closedBookings = $derived(bookings.filter((b) => b.status === 'Closed'));
    const sortedBookings = $derived([...activeBookings, ...closedBookings]);

    const {
        form: editFormData,
        errors: editErrors,
        enhance: editEnhance
    } = superForm(data.editForm, {
        onUpdated: async ({ form }) => {
            if (form.valid && form.message) {
                isEditDialogOpen = false;

                if (currentEditingUuid) {
                    try {
                        const updated = await api.getSubscription(currentEditingUuid);
                        const index = subscriptions.findIndex((sub) => sub.uuid === currentEditingUuid);
                        if (index !== -1 && updated.length > 0) {
                            subscriptions = [
                                ...subscriptions.slice(0, index),
                                updated[0],
                                ...subscriptions.slice(index + 1)
                            ];
                        }
                    } catch (e) {
                        toast.error(getErrorMessage(e));
                    }
                }

                currentEditingUuid = null;
            }
        },
        onResult: ({ result }) => {
            if (result.type === 'failure' && result.data?.error) {
                toast.error(result.data.error as string);
            }
        }
    });

    async function onPaused(uuid: string) {
        try {
            await api.pauseSubscription(uuid);
            const updated = await api.getSubscription(uuid);
            const index = subscriptions.findIndex((sub) => sub.uuid === uuid);
            if (index !== -1 && updated.length > 0) {
                subscriptions = [
                    ...subscriptions.slice(0, index),
                    updated[0],
                    ...subscriptions.slice(index + 1)
                ];
            }
        } catch (e) {
            toast.error(getErrorMessage(e));
        }
    }

    async function onRestored(uuid: string) {
        try {
            await api.restoreSubscription(uuid);
            const updated = await api.getSubscription(uuid);
            const index = subscriptions.findIndex((sub) => sub.uuid === uuid);
            if (index !== -1 && updated.length > 0) {
                subscriptions = [
                    ...subscriptions.slice(0, index),
                    updated[0],
                    ...subscriptions.slice(index + 1)
                ];
            }
        } catch (e) {
            toast.error(getErrorMessage(e));
        }
    }

    async function onClosed(uuid: string) {
        try {
            await api.closeSubscription(uuid);
            subscriptions = subscriptions.filter((sub) => sub.uuid !== uuid);
        } catch (e) {
            toast.error(getErrorMessage(e));
        }
    }

    function onEditDialogOpen(sub: SubscriptionResponse) {
        isEditDialogOpen = true;
        currentEditingUuid = sub.uuid;
        $editFormData = {
            uuid: sub.uuid,
            auto_enroll: sub.auto_enroll,
            any_date: sub.any_date
        };
    }
</script>

<div class="flex h-full w-full flex-col">
    <Header title="Подписки" />

    <div class="flex w-full flex-col items-center px-8 py-8">
        <Tabs.Root value="subscriptions" class="w-full">
            <div class="mb-6 flex items-center justify-between" in:fly={{ y: -10, duration: 240 }}>
                <Tabs.List>
                    <Tabs.Trigger value="subscriptions">Подписки</Tabs.Trigger>
                    <Tabs.Trigger value="bookings">
                        Записи
                        {#if activeBookings.length > 0}
							<span
                                    class="ml-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground"
                            >
								{activeBookings.length}
							</span>
                        {/if}
                    </Tabs.Trigger>
                </Tabs.List>

                <Tabs.Content value="subscriptions" class="flex-none">
                    <a
                            href="/subscriptions/new"
                            class={cn(
							buttonVariants({ variant: 'default' }),
							'text-md px-10 py-5 font-semibold tracking-wide uppercase'
						)}
                    >
                        Создать
                    </a>
                </Tabs.Content>
            </div>

            <!-- Subscriptions tab -->
            <Tabs.Content value="subscriptions">
                <div class="mb-4 flex items-center" in:fly={{ y: -6, duration: 200 }}>
					<span class="text-muted-foreground">
						<span class="font-bold text-primary">
							{#if isLoading}
								<Spinner />
							{:else}
								{subscriptions.length}
							{/if}
						</span>
						ПОДПИСОК
					</span>
                </div>

                <hr class="mb-6 w-full" />

                <div class="flex flex-col items-center gap-12">
                    {#if isLoading}
                        <Spinner />
                    {:else if subscriptions.length === 0}
                        <div class="py-12 text-center" in:fade={{ duration: 280, delay: 100 }}>
                            <p class="mb-2 text-lg text-muted-foreground">У вас пока нет активных подписок</p>
                            <p class="text-sm text-muted-foreground">
                                Создайте первую подписку чтобы начать автоматический поиск слотов
                            </p>
                        </div>
                    {:else}
                        {#each subscriptions as subscription, i (subscription.uuid)}
                            <div
                                    class="w-full"
                                    in:fly={{ y: 20, duration: 300, delay: i * 40 }}
                                    out:fade={{ duration: 200 }}
                                    animate:flip={{ duration: 300 }}
                            >
                                <Card
                                        {subscription}
                                        {labTypes}
                                        {labTopics}
                                        bind:isEditDialogOpen
                                        {onEditDialogOpen}
                                        {onPaused}
                                        {onRestored}
                                        {onClosed}
                                        {editFormData}
                                        {editErrors}
                                        {editEnhance}
                                />
                            </div>
                        {/each}
                    {/if}
                </div>
            </Tabs.Content>

            <!-- Bookings tab -->
            <Tabs.Content value="bookings">
                <div class="mb-4 flex items-center" in:fly={{ y: -6, duration: 200 }}>
					<span class="text-muted-foreground">
						<span class="font-bold text-primary">{bookings.length}</span>
						ЗАПИСЕЙ
					</span>
                </div>

                <hr class="mb-6 w-full" />

                <div class="flex flex-col items-center gap-12">
                    {#if bookings.length === 0}
                        <div class="py-12 text-center" in:fade={{ duration: 280, delay: 100 }}>
                            <p class="mb-2 text-lg text-muted-foreground">Записей пока нет</p>
                            <p class="text-sm text-muted-foreground">
                                Здесь появятся ваши записи на лабораторные после автоматической регистрации
                            </p>
                        </div>
                    {:else}
                        {#each sortedBookings as booking, i}
                            <div
                                    class="w-full"
                                    in:fly={{ y: 20, duration: 300, delay: i * 40 }}
                            >
                                <BookingCard {booking} {labTypes} {labTopics} />
                            </div>
                        {/each}
                    {/if}
                </div>
            </Tabs.Content>
        </Tabs.Root>
    </div>
</div>