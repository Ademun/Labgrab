<script lang="ts">
  import '../app.css';
  import favicon from '$lib/assets/favicon.svg';
  import {ModeWatcher} from "mode-watcher";
  import {Navigation} from "$lib/components/navigation";

  let {children} = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon}/>
</svelte:head>
<ModeWatcher />
<div class="h-screen pb-24">
    {@render children?.()}
</div>
<Navigation/>
