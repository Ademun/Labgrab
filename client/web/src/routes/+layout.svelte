<script lang="ts">
    import '../app.css';
    import favicon from '$lib/assets/favicon.svg';
    import {ModeWatcher} from "mode-watcher";
    import {Navigation} from "$lib/components/navigation";
    import {configStore} from "$lib/stores/config.ts";
    import {onMount} from "svelte";
    import {userStore} from "$lib/stores/user.ts";
    import {page} from '$app/state';
    import {Toaster, toast} from "svelte-sonner";

    let {children} = $props();

    onMount(async () => {
        try {
            await configStore.load()
        } catch (error) {
            console.log('Failed to load app config', error)
            toast.error('Не удалось загрузить настройки приложения', {
                description: 'Некоторые функции могут работать некорректно'
            });
        }
    })

    $effect(() => {
        if (page.data.user) {
            userStore.setUser(page.data.user);
        } else {
            userStore.reset()
        }
    })
</script>

<svelte:head>
    <link rel="icon" href={favicon}/>
</svelte:head>
<ModeWatcher/>
<Toaster
        richColors
        closeButton
        position="bottom-center"
        theme={undefined}
/>
<div class="h-screen pb-24">
    {@render children?.()}
</div>
<Navigation/>
