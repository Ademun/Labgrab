<script lang="ts">
    import {Fallback, Image, Root} from "$lib/components/ui/avatar";
    import {
        Ban,
        Calendar,
        ChartNoAxesCombined,
        ChevronRight,
        CircleDollarSign,
        CircleQuestionMark,
        Headset,
        User
    } from "lucide-svelte"
    import type {PageProps} from "../../../../.svelte-kit/types/src/routes/account/$types";
    import {Button} from "$lib/components/ui/button";
    import {fade, scale} from 'svelte/transition';

    let {data}: PageProps = $props()

</script>

<div class="flex flex-col items-center w-full mx-auto px-8 py-8">
    <h1 class="text-xl font-medium"
        in:fade={{
            delay: 100,
            duration: 300
        }}
        out:fade={{
            duration: 250,
            delay: 0
        }}>
        Профиль
    </h1>

    <div
            in:scale={{
            delay: 150,
            duration: 300,
            start: 0.9
          }}
            out:scale={{
            duration: 250,
            start: 0.9,
            delay: 0
          }}>
        <Root class="w-25 h-25 outline-3 outline-offset-2 outline-primary my-6">
            <Image src={data.photo_url}/>
            <Fallback>
                <User/>
            </Fallback>
        </Root>
    </div>

    <span class="text-lg font-bold"
          in:fade={{
            delay: 200,
            duration: 300
          }}
          out:fade={{
            duration: 250,
            delay: 0
          }}>{data.surname} {data.name} {data.patronymic}</span>
    <span class="text-md font-semibold text-primary"
          in:fade={{
            delay: 225,
            duration: 300
          }}
          out:fade={{
            duration: 250,
            delay: 0
          }}>{data.group_code}</span>


    <span class="text-md text-muted-foreground"
          in:fade={{
            delay: 250,
            duration: 300
          }}
          out:fade={{
            duration: 250,
            delay: 0
          }}>{data.phone_number}</span>

    <div in:scale={{
                delay: 300,
                duration: 300,
                start: 0.9
            }}
         out:scale={{
                duration: 250,
                start: 0.9,
                delay: 0
            }}>
        <Button class="py-5 mt-6"
        >
            Редактировать профиль
        </Button>
    </div>

    <nav class="bg-card rounded-2xl w-full overflow-hidden
    shadow-[0_2px_8px_rgba(0,0,0,0.04),0_1px_2px_rgba(0,0,0,0.06)]
    border border-border/40
    hover:shadow-[0_4px_12px_rgba(0,0,0,0.06),0_2px_4px_rgba(0,0,0,0.08)]
    hover:-translate-y-0.5
    transition-all duration-200 px-4 py-4 mt-8"
         in:fade={{
            delay: 350,
            duration: 300
         }}
         out:fade={{
            duration: 250,
            delay: 50
         }}>
        <ul class="flex flex-col gap-6 font-medium">
            <li>
                <button class="flex justify-between items-center w-full">
                    <span class="flex items-center gap-2">
                        <Calendar/>
                        Расписание
                    </span>
                    <ChevronRight/>
                </button>
            </li>
            <li>
                <button class="flex justify-between items-center w-full">
                    <span class="flex items-center gap-2">
                        <Ban/>
                        Черный список
                    </span>
                    <ChevronRight/>
                </button>
            </li>
            <li>
                <button class="flex justify-between items-center w-full">
                    <span class="flex items-center gap-2">
                        <ChartNoAxesCombined/>
                        Статистика
                    </span>
                    <ChevronRight/>
                </button>
            </li>
        </ul>
    </nav>

    <nav class="bg-card rounded-2xl w-full overflow-hidden
    shadow-[0_2px_8px_rgba(0,0,0,0.04),0_1px_2px_rgba(0,0,0,0.06)]
    border border-border/40
    hover:shadow-[0_4px_12px_rgba(0,0,0,0.06),0_2px_4px_rgba(0,0,0,0.08)]
    hover:-translate-y-0.5
    transition-all duration-200 px-4 py-4 mt-8"
         in:fade={{
            delay: 400,
            duration: 300
         }}
         out:fade={{
            duration: 250,
            delay: 100
         }}>
        <ul class="flex flex-col gap-6 font-medium">
            <li>
                <button class="flex justify-between items-center w-full">
                    <span class="flex items-center gap-2">
                        <CircleQuestionMark/>
                        О сервисе
                    </span>
                    <ChevronRight/>
                </button>
            </li>
            <li>
                <button class="flex justify-between items-center w-full">
                    <span class="flex items-center gap-2">
                        <Headset/>
                        Техподдержка
                    </span>
                    <ChevronRight/>
                </button>
            </li>
            <li>
                <button class="flex justify-between items-center w-full">
                    <span class="flex items-center gap-2">
                        <CircleDollarSign/>
                        Поддержать разработчика
                    </span>
                </button>
            </li>
        </ul>
    </nav>

    <div class="w-full" in:fade={{
                delay: 450,
                duration: 300
            }}
         out:fade={{
                duration: 250,
                delay: 150
            }}>
        <Button variant="outline" class="text-destructive text-md font-semibold w-full py-6 mt-8"
        >
            Выйти
        </Button>
    </div>
</div>