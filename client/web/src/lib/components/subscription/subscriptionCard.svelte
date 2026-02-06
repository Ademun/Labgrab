<script lang="ts">
    import {Button} from "$lib/components/ui/button";
    import {cn} from "$lib/utils.ts";
    import {fade, scale} from 'svelte/transition';
    import {
        Action as AlertAction,
        Cancel as AlertCancel,
        Content as AlertContent,
        Description as AlertDescription,
        Root as AlertRoot,
        Title as AlertTitle,
        Trigger as AlertTrigger,
    } from "$lib/components/ui/alert-dialog";

    import {api} from "$lib/api/client";
    import {handleApiError, toast} from "$lib/utils/toast";
    import {getLabTopicName, getLabTypeName} from "$lib/stores/config";
    import type {Subscription} from "$lib/api/types.ts";
    import {Spinner} from "$lib/components/ui/spinner";

    let {
        subscription,
        onDeleted,
        onPaused,
        onResumed,
    }: {
        subscription: Subscription;
        onDeleted?: (uuid: string) => void;
        onPaused?: (uuid: string) => void;
        onResumed?: (uuid: string) => void;
    } = $props();

    let isPausing = $state<boolean>(false);
    let isResuming = $state<boolean>(false);
    let isDeleting = $state<boolean>(false);

    const isPerformance = $derived(subscription.lab_type === "Performance");
    const colors = $derived(isPerformance ? "primary" : "blue-500");

    const isPaused = $derived(subscription.status === 'paused');

    async function handlePause() {
        isPausing = true;
        try {
            await api.pauseSubscription(subscription.uuid);

            onPaused?.(subscription.uuid);
        } catch (error) {
            console.error('Failed to pause subscription:', error);
            handleApiError(error, 'Не удалось приостановить подписку');
        } finally {
            isPausing = false;
        }
    }

    async function handleResume() {
        isResuming = true;
        try {
            await api.resumeSubscription(subscription.uuid);

            onResumed?.(subscription.uuid);
        } catch (error) {
            console.error('Failed to resume subscription:', error);
            handleApiError(error, 'Не удалось возобновить подписку');
        } finally {
            isResuming = false;
        }
    }

    async function handleDelete() {
        isDeleting = true;
        try {
            await api.deleteSubscription(subscription.uuid);

            onDeleted?.(subscription.uuid);
        } catch (error) {
            console.error('Failed to delete subscription:', error);
            handleApiError(error, 'Не удалось удалить подписку');
        } finally {
            isDeleting = false;
        }
    }
</script>

<div
        class="bg-card rounded-2xl w-full overflow-hidden shadow-xl"
        in:fade={{
        delay: 0,
        duration: 300
    }}
        out:fade={{
        duration: 250,
        delay: 0
    }}
>
    <div
            class={cn(
            "text-sm font-semibold flex items-center justify-between px-5 py-3",
            `bg-${colors}`,
            isPaused && "opacity-60"
        )}
            in:fade={{
            delay: 50,
            duration: 250
        }}
            out:fade={{
            duration: 200,
            delay: 0
        }}
    >
        <div
                class="flex items-center gap-2"
                in:scale={{
                delay: 100,
                duration: 200,
                start: 0.95
            }}
                out:scale={{
                duration: 150,
                start: 0.95,
                delay: 0
            }}
        >
            <span class={cn(
                "w-1.5 h-1.5 rounded-full",
                isPaused ? "bg-foreground/50" : "bg-foreground animate-pulse"
            )}></span>
            <span class="uppercase tracking-wide">
                {isPaused ? 'На паузе' : 'Активна'}
            </span>
        </div>
        <span
                in:fade={{
                delay: 150,
                duration: 200
            }}
                out:fade={{
                duration: 150,
                delay: 0
            }}
        >
            {subscription.created_at}
        </span>
    </div>

    <div class="px-5 py-6">
        <div
                class="mb-5"
                in:fade={{
                delay: 100,
                duration: 300
            }}
                out:fade={{
                duration: 200,
                delay: 50
            }}
        >
            <div
                    class="flex items-baseline gap-3 mb-3"
                    in:scale={{
                    delay: 150,
                    duration: 250,
                    start: 0.98
                }}
                    out:scale={{
                    duration: 200,
                    start: 0.98,
                    delay: 0
                }}
            >
                <span class="text-[2rem] font-black leading-none">
                    №{subscription.lab_number}
                </span>
                <span class={cn(
                    "text-lg font-black uppercase leading-none tracking-tight",
                    `text-${colors}`
                )}>
                    {getLabTypeName(subscription.lab_type)}
                </span>
            </div>
            <span
                    class="text-sm text-muted-foreground font-medium"
                    in:fade={{
                    delay: 200,
                    duration: 250
                }}
                    out:fade={{
                    duration: 150,
                    delay: 100
                }}
            >
                {getLabTopicName(subscription.lab_topic)}
            </span>
        </div>

        <div
                class="flex items-start gap-12 py-4 border-t border-b border-border/50"
                in:fade={{
                delay: 250,
                duration: 300
            }}
                out:fade={{
                duration: 200,
                delay: 150
            }}
        >
            <div
                    class="flex flex-col gap-1.5"
                    in:scale={{
                    delay: 300,
                    duration: 200,
                    start: 0.95
                }}
                    out:scale={{
                    duration: 150,
                    start: 0.95,
                    delay: 0
                }}
            >
                <span class="text-[0.65rem] text-muted-foreground uppercase tracking-wider font-medium">
                    Аудитория
                </span>
                <span class="text-[0.9375rem] font-semibold text-foreground">
                    {subscription.lab_auditorium ?? 'Любая'}
                </span>
            </div>
            <div
                    class="flex flex-col gap-1.5"
                    in:scale={{
                    delay: 350,
                    duration: 200,
                    start: 0.95
                }}
                    out:scale={{
                    duration: 150,
                    start: 0.95,
                    delay: 50
                }}
            >
                <span class="text-[0.65rem] text-muted-foreground uppercase tracking-wider font-medium">
                    Проверок
                </span>
                <span class="text-[0.9375rem] font-semibold text-foreground">
                    {subscription.checks_count ?? 0}
                </span>
            </div>
        </div>

        <div
                class="flex flex-col items-center gap-2.5 mt-6"
                in:scale={{
                delay: 400,
                duration: 300,
                start: 0.9
            }}
                out:scale={{
                duration: 250,
                start: 0.9,
                delay: 200
            }}
        >
            <div class="w-full">
                {#if isPaused}
                    <Button
                            variant="default"
                            class={cn(
                            "w-full py-5 font-semibold text-sm uppercase tracking-wide",
                            `bg-${colors}`
                        )}
                            onclick={handleResume}
                            disabled={isResuming || isDeleting}
                    >
                        {#if isResuming}
                            <span class="flex items-center gap-2">
                                <span class="animate-spin">⏳</span>
                                Возобновление...
                            </span>
                        {:else}
                            Возобновить
                        {/if}
                    </Button>
                {:else}
                    <Button
                            variant="outline"
                            class="w-full py-5 font-semibold text-sm uppercase tracking-wide hover:bg-accent"
                            onclick={handlePause}
                            disabled={isPausing || isDeleting}
                    >
                        {#if isPausing}
                            <span class="flex items-center gap-2">
                                <span class="animate-spin">⏳</span>
                                Пауза...
                            </span>
                        {:else}
                            Пауза
                        {/if}
                    </Button>
                {/if}
            </div>

            <div class="w-full">
                <Button
                        class={cn("w-full py-5 font-semibold text-sm uppercase tracking-wide", `bg-${colors}`)}
                        disabled={isPausing || isResuming || isDeleting}
                >
                    Настроить
                </Button>
            </div>

            <div class="w-full">
                <AlertRoot>
                    <AlertTrigger class="w-full">
                        <Button
                                variant="outline"
                                class="w-full py-5 font-semibold text-sm uppercase tracking-wide"
                                disabled={isPausing || isResuming || isDeleting}
                        >
                            {#if isDeleting}
                                <Spinner />
                            {:else}
                                Отменить
                            {/if}
                        </Button>
                    </AlertTrigger>
                    <AlertContent>
                        <AlertTitle>Вы уверены?</AlertTitle>
                        <AlertDescription>
                            Система перестанет проверять слоты для этой лабораторной работы.
                            Это действие нельзя отменить.
                        </AlertDescription>
                        <div class="flex justify-end gap-3 mt-4">
                            <AlertCancel>Отмена</AlertCancel>
                            <AlertAction onclick={handleDelete}>
                                Да, отменить подписку
                            </AlertAction>
                        </div>
                    </AlertContent>
                </AlertRoot>
            </div>
        </div>
    </div>
</div>