<script lang="ts">
    import {SubscriptionCard, SubscriptionForm} from '$lib/components/subscription';
    import {buttonVariants} from '$lib/components/ui/button';
    import {Content, Description, Header, Root, Title, Trigger} from '$lib/components/ui/dialog';
    import {cn} from '$lib/utils.ts';
    import {fade, fly, scale} from 'svelte/transition';
    import {backOut} from 'svelte/easing';

    import {api} from '$lib/api/client';
    import {handleApiError, toast} from '$lib/utils/toast';
    import {onMount} from 'svelte';
    import type {Subscription} from '$lib/api/types';

    let subscriptions = $state<Subscription[]>([]);
    let isLoadingSubscriptions = $state<boolean>(false);

    let isDialogOpen = $state(false);

    async function loadSubscriptions() {
        isLoadingSubscriptions = true;
        try {
            subscriptions = await api.getSubscriptions();
        } catch (error) {
            console.error('Failed to load subscriptions:', error);
            handleApiError(error, 'Не удалось загрузить подписки');
        } finally {
            isLoadingSubscriptions = false;
        }
    }

    async function onSubscriptionCreated() {
        toast.success('Подписка создана', {
            description: 'Система начнёт проверять слоты автоматически'
        });

        isDialogOpen = false;

        await loadSubscriptions();
    }

    async function onSubscriptionDeleted(uuid: string) {
        toast.success('Подписка отменена', {
            description: 'Система больше не будет проверять слоты для этой работы'
        });

        subscriptions = subscriptions.filter(sub => sub.uuid !== uuid);
    }

    async function onSubscriptionPaused(uuid: string) {
        toast.info('Подписка поставлена на паузу', {
            description: 'Вы можете возобновить её в любой момент'
        });

        subscriptions = subscriptions.map(sub =>
            sub.uuid === uuid ? {...sub, status: 'paused' as const} : sub
        );
    }

    async function onSubscriptionResumed(uuid: string) {
        toast.success('Подписка возобновлена', {
            description: 'Система снова проверяет слоты'
        });

        subscriptions = subscriptions.map(sub =>
            sub.uuid === uuid ? {...sub, status: 'active' as const} : sub
        );
    }

    onMount(() => {
        if (!subscriptions || subscriptions.length === 0) {
            loadSubscriptions();
        }
    });
</script>

<div
        class="flex flex-col items-center w-full px-8 py-8"
        in:fly={{
        y: 20,
        duration: 400,
        easing: backOut,
        opacity: 0
    }}
        out:fly={{
        y: 20,
        duration: 300,
        easing: backOut,
        opacity: 0
    }}
>
    <h1
            class="text-xl font-medium"
            in:fade={{
            delay: 100,
            duration: 300
        }}
            out:fade={{
            duration: 250,
            delay: 0
        }}
    >
        Подписки
    </h1>

    <div
            class="my-10 w-full"
            in:fade={{
            delay: 150,
            duration: 300
        }}
            out:fade={{
            duration: 250,
            delay: 50
        }}
    >
        <div
                class="flex justify-between items-center"
                in:scale={{
                delay: 200,
                duration: 300,
                start: 0.95
            }}
                out:scale={{
                duration: 250,
                start: 0.95,
                delay: 0
            }}
        >
            <span class="text-muted-foreground">
                <span
                        class="text-primary font-bold"
                        in:fade={{
                        delay: 250,
                        duration: 200
                    }}
                        out:fade={{
                        duration: 200,
                        delay: 0
                    }}
                >
                    {#if isLoadingSubscriptions}
                        <span class="animate-pulse">—</span>
                    {:else}
                        {subscriptions.length}
                    {/if}
                </span>
                ПОДПИСОК
            </span>

            <div
                    in:scale={{
                    delay: 300,
                    duration: 300,
                    start: 0.9
                }}
                    out:scale={{
                    duration: 250,
                    start: 0.9,
                    delay: 50
                }}
            >
                <Root bind:open={isDialogOpen}>
                    <Trigger class={cn(buttonVariants({ variant: 'default' }), 'px-12')}>
                        СОЗДАТЬ
                    </Trigger>
                    <Content class="max-w-lg">
                        <Header class="text-left">
                            <Title>Новая подписка</Title>
                            <Description>
                                Настройте параметры отслеживания лабораторной работы
                            </Description>
                        </Header>

                        <SubscriptionForm
                                bind:open={isDialogOpen}
                                onCreated={onSubscriptionCreated}
                        />
                    </Content>
                </Root>
            </div>
        </div>

        <hr
                class="w-full my-6"
                in:fade={{
                delay: 350,
                duration: 300
            }}
                out:fade={{
                duration: 200,
                delay: 100
            }}
        />

        <div
                class="flex flex-col items-center gap-12"
                in:fade={{
                delay: 400,
                duration: 300
            }}
                out:fade={{
                duration: 250,
                delay: 150
            }}
        >
            {#if isLoadingSubscriptions}
                <div class="w-full space-y-12">
                    {#each [1, 2] as i}
                        <div
                                class="bg-card rounded-2xl w-full h-96 animate-pulse"
                                in:fade={{ delay: i * 100, duration: 300 }}
                        ></div>
                    {/each}
                </div>
            {:else if subscriptions.length === 0}
                <div
                        class="text-center py-12"
                        in:fade={{ duration: 300 }}
                >
                    <p class="text-lg text-muted-foreground mb-2">
                        У вас пока нет активных подписок
                    </p>
                    <p class="text-sm text-muted-foreground">
                        Создайте первую подписку чтобы начать автоматический поиск слотов
                    </p>
                </div>
            {:else}
                {#each subscriptions as subscription, index (subscription.uuid)}
                    <div
                            class="w-full"
                            in:fly={{
                            y: 20,
                            duration: 300,
                            delay: 450 + index * 100,
                            opacity: 0
                        }}
                            out:fly={{
                            y: 20,
                            duration: 250,
                            delay: index * 50,
                            opacity: 0
                        }}
                    >
                        <SubscriptionCard
                                {subscription}
                                onDeleted={onSubscriptionDeleted}
                                onPaused={onSubscriptionPaused}
                                onResumed={onSubscriptionResumed}
                        />
                    </div>
                {/each}
            {/if}
        </div>
    </div>
</div>