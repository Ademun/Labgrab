<script lang="ts">
    import {Button} from "$lib/components/ui/button";
    import {cn, formatShortDate} from "$lib/utils/utils.ts";
    import {LAB_TYPE_COLORS, type Subscription} from "$lib/types/subscription.ts";
    import {translateLabTopic, translateLabType} from "$lib/utils/translations.ts";
    import {fade, scale} from 'svelte/transition';

    export let subscription: Subscription;
    const isPerformance = subscription.lab_type === "Performance";
    const colors = isPerformance ? LAB_TYPE_COLORS.Performance : LAB_TYPE_COLORS.Defence;
</script>

<div class="bg-card rounded-2xl w-full overflow-hidden
    shadow-[0_2px_8px_rgba(0,0,0,0.04),0_1px_2px_rgba(0,0,0,0.06)]
    border border-border/40
    hover:shadow-[0_4px_12px_rgba(0,0,0,0.06),0_2px_4px_rgba(0,0,0,0.08)]
    hover:-translate-y-0.5
    transition-all duration-200"
     in:fade={{
        delay: 0,
        duration: 300
    }}
     out:fade={{
        duration: 250,
        delay: 0
    }}>

    <div class={cn(
        "text-sm font-semibold flex items-center justify-between px-5 py-3",
        colors.bg
    )}
         in:fade={{
            delay: 50,
            duration: 250
         }}
         out:fade={{
            duration: 200,
            delay: 0
         }}>
        <div class="flex items-center gap-2"
             in:scale={{
                delay: 100,
                duration: 200,
                start: 0.95
             }}
             out:scale={{
                duration: 150,
                start: 0.95,
                delay: 0
             }}>
            <span class="w-1.5 h-1.5 bg-foreground rounded-full animate-pulse"></span>
            <span class="uppercase tracking-wide">Активна</span>
        </div>
        <span
              in:fade={{
                delay: 150,
                duration: 200
              }}
              out:fade={{
                duration: 150,
                delay: 0
              }}>
            {formatShortDate(subscription.created_at)}
        </span>
    </div>

    <div class="px-5 py-6">
        <div class="mb-5"
             in:fade={{
                delay: 100,
                duration: 300
             }}
             out:fade={{
                duration: 200,
                delay: 50
             }}>
            <div class="flex items-baseline gap-3 mb-3"
                 in:scale={{
                    delay: 150,
                    duration: 250,
                    start: 0.98
                 }}
                 out:scale={{
                    duration: 200,
                    start: 0.98,
                    delay: 0
                 }}>
                <span class="text-[2rem] font-black leading-none">
                    №{subscription.lab_number}
                </span>
                <span class={cn(
                    "text-lg font-black uppercase leading-none tracking-tight",
                    colors.text
                )}>
                    {translateLabType(subscription.lab_type)}
                </span>
            </div>
            <span class="text-sm text-muted-foreground font-medium"
                  in:fade={{
                    delay: 200,
                    duration: 250
                  }}
                  out:fade={{
                    duration: 150,
                    delay: 100
                  }}>
                {translateLabTopic(subscription.lab_topic)}
            </span>
        </div>

        <div class="flex items-start gap-12 py-4 border-t border-b border-border/50"
             in:fade={{
                delay: 250,
                duration: 300
             }}
             out:fade={{
                duration: 200,
                delay: 150
             }}>
            <div class="flex flex-col gap-1.5"
                 in:scale={{
                    delay: 300,
                    duration: 200,
                    start: 0.95
                 }}
                 out:scale={{
                    duration: 150,
                    start: 0.95,
                    delay: 0
                 }}>
                <span class="text-[0.65rem] text-muted-foreground uppercase tracking-wider font-medium">
                    Аудитория
                </span>
                <span class="text-[0.9375rem] font-semibold text-foreground">
                    {subscription.lab_auditorium ? subscription.lab_auditorium : 'Любая'}
                </span>
            </div>
            <div class="flex flex-col gap-1.5"
                 in:scale={{
                    delay: 350,
                    duration: 200,
                    start: 0.95
                 }}
                 out:scale={{
                    duration: 150,
                    start: 0.95,
                    delay: 50
                 }}>
                <span class="text-[0.65rem] text-muted-foreground uppercase tracking-wider font-medium">
                    Проверок
                </span>
                <span class="text-[0.9375rem] font-semibold text-foreground">
                    4
                </span>
            </div>
        </div>

        <div class="flex flex-col items-center gap-2.5 mt-6"
             in:scale={{
                delay: 400,
                duration: 300,
                start: 0.9
             }}
             out:scale={{
                duration: 250,
                start: 0.9,
                delay: 200
             }}>
            <div class="w-full" in:fade={{
                    delay: 450,
                    duration: 200
                }}
                 out:fade={{
                    duration: 150,
                    delay: 0
                }}>
                <Button
                        variant="outline"
                        class="w-full py-5 font-semibold text-sm uppercase tracking-wide hover:bg-accent"
                >
                    Пауза
                </Button>
            </div>

            <div class="w-full" in:scale={{
                    delay: 500,
                    duration: 250,
                    start: 0.95
                }}
                 out:scale={{
                    duration: 200,
                    start: 0.95,
                    delay: 50
                }}>
                <Button
                        class={cn(
                    "w-full py-5 font-semibold text-sm uppercase tracking-wide",
                    colors.bg,
                    colors.hover
                )}
                >
                    Настроить
                </Button>
            </div>

            <div class="w-full" in:fade={{
                    delay: 550,
                    duration: 200
                }}
                 out:fade={{
                    duration: 150,
                    delay: 100
                }}>
                <Button
                        variant="outline"
                        class="w-full py-5 font-semibold text-sm uppercase tracking-wide
                text-muted-foreground
                hover:bg-destructive hover:text-destructive-foreground hover:border-destructive"
                >
                    Отменить
                </Button>
            </div>
        </div>
    </div>
</div>