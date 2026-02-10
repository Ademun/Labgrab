<script lang="ts" module>
	import { BookOpenText, House, User } from 'lucide-svelte';
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
		{ icon: User, label: 'Аккаунт', href: '/account' }
	];

	function isActive(href: string): boolean {
		return page.url.pathname === href;
	}

	function handleNavClick(href: string) {
		goto(href);
	}
</script>

<script>
	import { cn } from '$lib/utils.ts';
</script>

<footer class="fixed bottom-0 left-0 right-0 bg-background border-t border-border">
	<nav class="flex justify-around items-center px-2 py-3 max-w-screen-sm mx-auto">
		{#each navItems as item}
			<button
				type="button"
				class={cn(
					'flex flex-col items-center justify-center gap-1 py-2 px-3 rounded-md transition-all hover:bg-accent hover:text-primary',
					isActive(item.href) ? 'text-primary' : 'text-accent-foreground mb-1'
				)}
				on:click={() => handleNavClick(item.href)}
			>
				<svelte:component this={item.icon} class="w-6 h-6" />
				<span class="text-xs font-medium">{item.label}</span>
			</button>
		{/each}
	</nav>
</footer>

<style>
	footer {
		-webkit-tap-highlight-color: transparent;
	}
</style>
