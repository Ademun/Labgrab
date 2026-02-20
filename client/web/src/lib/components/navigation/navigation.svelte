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

<footer class="fixed right-0 bottom-0 left-0 border-t border-border bg-background">
	<nav class="mx-auto flex max-w-screen-sm items-center justify-around px-2 py-3">
		{#each navItems as item}
			<button
				type="button"
				class={cn(
					'flex flex-col items-center justify-center gap-1 rounded-md px-3 py-2 transition-all hover:bg-accent hover:text-primary',
					isActive(item.href) ? 'text-primary' : 'mb-1 text-accent-foreground'
				)}
				on:click={() => handleNavClick(item.href)}
			>
				<svelte:component this={item.icon} class="h-6 w-6" />
				<span class="text-xs font-medium">{item.label}</span>
			</button>
		{/each}
	</nav>
</footer>
