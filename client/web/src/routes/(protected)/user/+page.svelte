<script lang="ts">
	import { Fallback, Image, Root } from '$lib/components/ui/avatar/index.js';
	import { Ban, Calendar, ChartNoAxesCombined, ChevronRight, Link, LogOut, User } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { fly, fade } from 'svelte/transition';
	import { Header } from '$lib/components/navigation/index.js';

	let { data } = $props();
	const user = $derived(data.user);
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Аккаунт" />

	<div class="flex flex-1 flex-col items-center justify-center px-6">
		{#if user}
			<div class="flex w-full max-w-sm flex-col items-center justify-center">
				<!-- Аватар -->
				<div in:fly={{ y: -16, duration: 300, delay: 0 }}>
					<Root class="my-6 h-25 w-25 outline-3 outline-offset-2 outline-primary">
						<Image src={user.telegram_photo_url} alt="Фото профиля" />
						<Fallback>
							<User />
						</Fallback>
					</Root>
				</div>

				<!-- Имя -->
				<span class="text-lg font-bold" in:fly={{ y: 12, duration: 280, delay: 60 }}>
					{user.surname}
					{user.name}
					{user.patronymic}
				</span>

				{#if user.group_code}
					<span
						class="text-md font-semibold text-primary"
						in:fly={{ y: 10, duration: 260, delay: 100 }}
					>
						{user.group_code}
					</span>
				{/if}

				{#if user.phone_number}
					<span class="text-md text-muted-foreground" in:fly={{ y: 10, duration: 260, delay: 120 }}>
						{user.phone_number}
					</span>
				{/if}

				<!-- Кнопка редактировать -->
				<div in:fly={{ y: 10, duration: 260, delay: 140 }}>
					<a href="/user/details"><Button class="mt-6 py-5">Редактировать профиль</Button></a>
				</div>

				<!-- Навигационное меню -->
				<nav
					class="mt-8 w-full overflow-hidden rounded-2xl border border-border/40 bg-card px-4 py-4 shadow-sm"
					in:fly={{ y: 16, duration: 300, delay: 180 }}
				>
					<ul class="flex flex-col gap-6 font-medium">
						<li>
							<a href="/user/integration" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2">
									<Link class="h-5 w-5 text-[#FF9F0A]" />
									Интеграция
								</span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
						<li>
							<a href="/user/schedule" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2">
									<Calendar class="h-5 w-5 text-[#007AFF]" />
									Расписание
								</span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
						<li>
							<a href="/user/blacklist" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2">
									<Ban class="h-5 w-5 text-[#FF3B30]" />
									Черный список
								</span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
						<li>
							<a href="/user/stats" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2">
									<ChartNoAxesCombined class="h-5 w-5 text-[#34C759]" />
									Статистика
								</span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
					</ul>
				</nav>

				<!-- Дополнительное меню -->
				<nav
					class="mt-4 w-full overflow-hidden rounded-2xl border border-border/40 bg-card px-4 py-4 shadow-sm"
					in:fly={{ y: 16, duration: 300, delay: 220 }}
				>
					<ul class="flex flex-col gap-6 font-medium">
						<li>
							<a href="/user/about" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2"> О сервисе </span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
						<li>
							<a href="/user/support" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2"> Техподдержка </span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
						<li>
							<a href="/user/donate" class="flex w-full items-center justify-between">
								<span class="flex items-center gap-2"> Поддержать разработчика </span>
								<ChevronRight class="h-5 w-5" />
							</a>
						</li>
					</ul>
				</nav>

				<!-- Кнопка выйти -->
				<div class="w-full" in:fly={{ y: 12, duration: 260, delay: 260 }}>
					<Button variant="outline" class="text-md mt-8 w-full py-6 font-semibold text-destructive">
						<span class="flex items-center gap-2">
							<LogOut class="h-5 w-5" />
							Выйти
						</span>
					</Button>
				</div>
			</div>
		{/if}
	</div>
</div>