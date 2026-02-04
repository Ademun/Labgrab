<script lang="ts">
    import {Field, Group, Label, Legend, Set} from "$lib/components/ui/field";
    import {Input} from "$lib/components/ui/input";
    import {Button} from "$lib/components/ui/button";
    import {fade, fly, scale} from 'svelte/transition';
    import {goto} from "$app/navigation";
    import type {UserDetails} from "$lib/types/user.ts";

    let name = $state<string>()
    let surname = $state<string>()
    let patronymic = $state<string>()
    let group_code = $state<string>()
    let phone_number = $state<string>()

    const submitDetails = async () => {
        const details: UserDetails = {
            name: name,
            surname: surname,
            patronymic: patronymic,
            group_code: group_code,
            phone_number: phone_number,
        }

        const response = await fetch("/api/users", {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify(details)
        })

        if (!response.ok) {
            console.error(response)
            return
        }

        await goto("/account");
    }


    const skip = async () => {
        await goto("/account");
    }
</script>

<div class="flex flex-col items-center h-full w-full mx-auto px-8 py-8">
    <div class="my-auto bg-card text-center rounded-2xl shadow-lg p-6"
         in:scale={{
            delay: 100,
            duration: 300,
            start: 0.95
         }}
         out:scale={{
            duration: 250,
            start: 0.95,
            delay: 0
         }}>

        <h1 class="text-xl font-bold font-archivo-black tracking-wider" in:fade={{
                delay: 200,
                duration: 300
            }}
            out:fade={{
                duration: 250,
                delay: 0
            }}>
            Уточним информацию?
        </h1>

        <div in:fade={{
                delay: 250,
                duration: 300
             }}
             out:fade={{
                duration: 250,
                delay: 50
             }}>
            <Set class="py-6">

                <div in:fade={{
                        delay: 300,
                        duration: 250
                    }}
                     out:fade={{
                        duration: 200,
                        delay: 100
                    }}>
                    <Legend class="text-left">
                        <span class="text-muted-foreground text-sm font-semibold tracking-wide">
                            Эта информация необходима для автозаписи
                        </span>
                    </Legend>
                </div>

                <div in:fade={{
                      delay: 350,
                      duration: 300
                   }}
                     out:fade={{
                      duration: 250,
                      delay: 150
                   }}>
                    <Group onsubmit={submitDetails}>

                        <div in:fly={{
                                y: 10,
                                duration: 250,
                                delay: 400,
                                opacity: 0
                            }}
                             out:fly={{
                                y: 10,
                                duration: 200,
                                delay: 0,
                                opacity: 0
                            }}>
                            <Field>
                                <Label class="text-sm font-medium">
                                    Имя <span class="text-primary">*</span>
                                </Label>
                                <Input required type="text" placeholder="Иван" bind:value={name}/>
                            </Field>
                        </div>

                        <div in:fly={{
                                y: 10,
                                duration: 250,
                                delay: 450,
                                opacity: 0
                            }}
                             out:fly={{
                                y: 10,
                                duration: 200,
                                delay: 50,
                                opacity: 0
                            }}>
                            <Field>
                                <Label class="text-sm font-medium">
                                    Фамилия <span class="text-primary">*</span>
                                </Label>
                                <Input required type="text" placeholder="Иванов" bind:value={surname}/>
                            </Field>
                        </div>

                        <div in:fly={{
                                y: 10,
                                duration: 250,
                                delay: 500,
                                opacity: 0
                            }}
                             out:fly={{
                                y: 10,
                                duration: 200,
                                delay: 100,
                                opacity: 0
                            }}>
                            <Field>
                                <Label class="text-sm font-medium">
                                    Отчество <span class="text-primary">*</span>
                                </Label>
                                <Input required type="text" placeholder="Иванович" bind:value={patronymic}/>
                            </Field>
                        </div>

                        <div in:fly={{
                                y: 10,
                                duration: 250,
                                delay: 550,
                                opacity: 0
                            }}
                             out:fly={{
                                y: 10,
                                duration: 200,
                                delay: 150,
                                opacity: 0
                            }}>
                            <Field>
                                <Label class="text-sm font-medium">
                                    Группа <span class="text-primary">*</span>
                                </Label>
                                <Input required type="text" placeholder="ИН-24-8" bind:value={group_code}/>
                            </Field>
                        </div>

                        <div in:fly={{
                                y: 10,
                                duration: 250,
                                delay: 600,
                                opacity: 0
                            }}
                             out:fly={{
                                y: 10,
                                duration: 200,
                                delay: 200,
                                opacity: 0
                            }}>
                            <Field>
                                <Label class="text-sm font-medium">
                                    Телефон <span class="text-primary">*</span>
                                </Label>
                                <Input required type="tel" placeholder="+7 978 621 48 21" bind:value={phone_number}/>
                            </Field>
                        </div>
                    </Group>
                </div>
            </Set>
        </div>

        <div in:scale={{
                delay: 650,
                duration: 300,
                start: 0.9
            }}
             out:scale={{
                duration: 250,
                start: 0.9,
                delay: 250
            }}>
            <Field>
                <div in:scale={{
                        delay: 700,
                        duration: 250,
                        start: 0.95
                    }}
                     out:scale={{
                        duration: 200,
                        start: 0.95,
                        delay: 0
                    }}>
                    <Button type="submit" class="w-full py-5 font-semibold text-sm uppercase tracking-wide"
                            onclick={submitDetails}>
                        ПРОДОЛЖИТЬ
                    </Button>
                </div>

                <div in:fade={{
                        delay: 750,
                        duration: 250
                    }}
                     out:fade={{
                        duration: 200,
                        delay: 50
                    }}>
                    <Button variant="outline" class="w-full py-5 font-semibold text-sm uppercase tracking-wide"
                            onclick={skip}>
                        ПРОПУСТИТЬ
                    </Button>
                </div>
            </Field>
        </div>
    </div>
</div>