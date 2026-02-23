<script lang="ts">
	import { BookOpenText, House, User } from '@lucide/svelte';
	import { cn } from '$lib/utils.js';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	type NavItem = {
		icon: any;
		label: string;
		href: string;
	};

	const navItems: NavItem[] = [
		{ icon: House, label: 'Главная', href: '/' },
		{ icon: BookOpenText, label: 'Подписки', href: '/subscriptions' },
		{ icon: User, label: 'Аккаунт', href: '/user' }
	];

	function isActive(href: string): boolean {
		return page.url.pathname === href;
	}

	function handleNavClick(href: string) {
		goto(href);
	}
</script>

<footer
	class="fixed right-0 bottom-0 left-0 border-t border-border bg-background"
	style="view-transition-name: nav;"
>
	<nav class="mx-auto flex max-w-screen-sm items-center justify-around px-2 py-3">
		{#each navItems as item}
			{@const active = isActive(item.href)}
			<button
				type="button"
				class={cn(
					'relative flex flex-col items-center justify-center gap-1 rounded-md px-3 py-2',
					'transition-colors duration-200 hover:bg-accent hover:text-primary',
					active ? 'text-primary' : 'mb-1 text-accent-foreground'
				)}
				on:click={() => handleNavClick(item.href)}
			>
				<!-- Иконка с лёгким scale при активации -->
				<span class={cn('transition-transform duration-200', active ? 'scale-110' : 'scale-100')}>
					<svelte:component this={item.icon} class="h-6 w-6" />
				</span>

				<span class="text-xs font-medium">{item.label}</span>

				<!-- iOS-style активный индикатор: точка снизу -->
				{#if active}
					<span
						class="absolute -bottom-0.5 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-primary"
						style="animation: nav-dot-in 200ms cubic-bezier(0.34, 1.56, 0.64, 1) both;"
					></span>
				{/if}
			</button>
		{/each}
	</nav>
</footer>

<style>
	@keyframes nav-dot-in {
		from {
			transform: translateX(-50%) scale(0);
			opacity: 0;
		}
		to {
			transform: translateX(-50%) scale(1);
			opacity: 1;
		}
	}
</style>
