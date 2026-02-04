<script lang="ts">
    import {Field, Group, Label, Legend, Set} from "$lib/components/ui/field";
    import {Input} from "$lib/components/ui/input";
    import {Content, Item, Root, Trigger} from "$lib/components/ui/select";
    import {Checkbox} from "$lib/components/ui/checkbox";
    import {Button} from "$lib/components/ui/button";
    import {cn} from "$lib/utils/utils.ts";
    import {LAB_TYPE_COLORS, needsAuditorium, type NewSubscription} from "$lib/types/subscription.ts";
    import {invalidateAll} from "$app/navigation";
    import {fade} from "svelte/transition";
    import {translateTopicToEnglish} from "$lib/utils/translations.ts";

    let {open = $bindable(false)}: { open: boolean } = $props();

    let labType = $state<string>('Performance')
    let labTopic = $state<string>()
    let labNum = $state<number>()
    let labAuditorium = $state<number>()
    let autoSign = $state<boolean>(false)
    let anyDate = $state<boolean>(false)

    const createSubscription = async (event: Event) => {
        event.preventDefault()

        const subscription: NewSubscription = {
            lab_type: labType,
            lab_topic: translateTopicToEnglish(labTopic ?? ""),
            lab_number: Number(labNum ?? -1),
            lab_auditorium: labAuditorium != null ? Number(labAuditorium) : undefined,
            created_at: Math.round(new Date().getTime() / 1000)
        }

        const response = await fetch("/api/subscriptions", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify(subscription)
        })

        if (!response.ok) {
            console.error(response)
            return
        }

        open = false
        await invalidateAll()
    }

    $effect(() => {
        if (labType === 'Defence') {
            labAuditorium = undefined;
        }
    });

    const isDefence = $derived(labType === 'Defence');
    const colors = $derived(isDefence ? LAB_TYPE_COLORS.Defence : LAB_TYPE_COLORS.Performance);
</script>

<div>
    <form onsubmit={createSubscription} class="space-y-6">
        <Set class="bg-muted/30 rounded-xl p-6">
            <Legend class="mb-4">
                <span class="text-muted-foreground text-sm font-semibold uppercase tracking-wide">
                    Информация о работе
                </span>
            </Legend>

            <Group>
                <Field>
                    <Label class="text-sm font-medium mb-2">
                        Тип работы <span class="text-primary">*</span>
                    </Label>
                    <div class="flex justify-between items-center gap-3">
                        <Button
                                type="button"
                                class={cn(
                                "w-[48%] py-5 font-semibold text-sm uppercase tracking-wide",
                                labType === 'Performance' && colors.bg,
                                labType === 'Performance' && colors.hover
                            )}
                                variant={labType === 'Performance' ? 'default' : 'outline'}
                                onclick={() => labType = 'Performance'}
                        >
                            Выполнение
                        </Button>
                        <Button
                                type="button"
                                class={cn(
                                "w-[48%] py-5 font-semibold text-sm uppercase tracking-wide",
                                labType === 'Defence' && colors.bg,
                                labType === 'Defence' && colors.hover
                            )}
                                variant={labType === 'Defence' ? 'default' : 'outline'}
                                onclick={() => labType = 'Defence'}
                        >
                            Защита
                        </Button>
                    </div>
                </Field>

                <Field>
                    <Label class="text-sm font-medium mb-2">
                        Тема работы <span class="text-primary">*</span>
                    </Label>
                    <Root required type='single' bind:value={labTopic}>
                        <Trigger class="w-full">
                            <span>{labTopic || "Выберите тему из списка"}</span>
                        </Trigger>
                        <Content>
                            <Item value="Механика">Механика</Item>
                            <Item value="Электричество">Электричество</Item>
                            <Item value="Виртуальная">Виртуальная</Item>
                            <Item value="Оптика">Оптика</Item>
                            <Item value="Твёрдое тело">Твёрдое тело</Item>
                        </Content>
                    </Root>
                </Field>

                <Field>
                    <Label class="text-sm font-medium mb-2">
                        Номер работы <span class="text-primary">*</span>
                    </Label>
                    <Input
                            required
                            type="number"
                            min="1"
                            placeholder="Например: 3"
                            bind:value={labNum}
                            class="py-5"
                    />
                </Field>

                {#if needsAuditorium(labType)}
                    <div transition:fade={{duration: 300}}>
                        <Field>
                            <Label class="text-sm font-medium mb-2">
                                Аудитория
                                <span class="text-primary">
                                    *
                                </span>
                            </Label>
                            <Input
                                    type="number"
                                    placeholder="205"
                                    bind:value={labAuditorium}
                                    class="py-5"
                            />
                        </Field>
                    </div>
                {/if}
            </Group>
        </Set>

        <Set class="bg-muted/30 rounded-xl p-5">
            <Legend class="mb-4">
                <span class="text-muted-foreground text-sm font-semibold uppercase tracking-wide">
                    Настройки
                </span>
            </Legend>

            <Group class="space-y-4">
                <Field orientation="horizontal" class="flex items-start gap-3">
                    <Checkbox
                            id="subscription-auto"
                            class={cn(
                            "mt-0.5",
                            isDefence && "data-[state=checked]:bg-blue-500 data-[state=checked]:border-blue-500"
                        )}
                            bind:checked={autoSign}
                    />
                    <Label for="subscription-auto" class="cursor-pointer">
                        <div class="flex flex-col gap-1">
                            <span class="text-sm font-medium">Автоматическая запись</span>
                            <span class="text-xs text-muted-foreground leading-relaxed">
                                Система автоматически запишет вас когда появится свободное место
                            </span>
                        </div>
                    </Label>
                </Field>

                <Field orientation="horizontal" class="flex items-start gap-3">
                    <Checkbox
                            id="subscription-any-date"
                            class={cn(
                            "mt-0.5",
                            isDefence && "data-[state=checked]:bg-blue-500 data-[state=checked]:border-blue-500"
                        )}
                            bind:checked={anyDate}
                    />
                    <Label for="subscription-any-date" class="cursor-pointer">
                        <div class="flex flex-col gap-1">
                            <span class="text-sm font-medium">Записаться на любую дату</span>
                            <span class="text-xs text-muted-foreground leading-relaxed">
                                Не учитывать предпочтения по датам, взять первое доступное место
                            </span>
                        </div>
                    </Label>
                </Field>
            </Group>
        </Set>

        <Field>
            <Button
                    type="submit"
                    class={cn(
                    "w-full py-5 font-semibold text-sm uppercase tracking-wide",
                    colors.bg,
                    colors.hover
                )}
            >
                Создать подписку
            </Button>
        </Field>
    </form>
</div>