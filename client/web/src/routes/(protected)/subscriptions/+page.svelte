<script lang="ts">
	import { SubscriptionCard, SubscriptionForm } from '$lib/components/subscription';
	import { buttonVariants } from '$lib/components/ui/button';
	import { Content, Description, Header, Root, Title, Trigger } from '$lib/components/ui/dialog';
	import { cn } from '$lib/utils/utils.ts';
	import type { PageProps } from '../../../../.svelte-kit/types/src/routes/subscriptions/$types';
	import { fade, fly, scale } from 'svelte/transition';
	import { backOut } from 'svelte/easing';

	let { data }: PageProps = $props();
	let open = $state(false);
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
					}}>{data.subscriptions.length}</span
				>
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
				<Root bind:open>
					<Trigger class={cn(buttonVariants({ variant: 'default' }), 'px-12')}>СОЗДАТЬ</Trigger>
					<Content>
						<Header class="text-left">
							<Title>Новая подписка</Title>
							<Description>Настройте параметры отслеживания лабораторной работы</Description>
						</Header>
						<SubscriptionForm bind:open />
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
			{#each data.subscriptions as subscription, index}
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
					<SubscriptionCard {subscription} />
				</div>
			{/each}
		</div>
	</div>
</div>
