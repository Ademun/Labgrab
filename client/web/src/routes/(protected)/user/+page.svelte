<script lang="ts">
	import { Fallback, Image, Root } from '$lib/components/ui/avatar/index.js';
	import { Ban, Calendar, ChartNoAxesCombined, ChevronRight, LogOut, User } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	let { data } = $props();
	const user = $derived(data.user);
</script>

<div class="flex h-full flex-col items-center justify-center px-6">
	{#if user}
		<div class="flex w-full max-w-sm flex-col items-center justify-center">
			<div>
				<Root class="my-6 h-25 w-25 outline-3 outline-offset-2 outline-primary">
					<Image src={user.photo_url} alt="Фото профиля" />
					<Fallback>
						<User />
					</Fallback>
				</Root>
			</div>

			<span class="text-lg font-bold">
				{user.surname}
				{user.name}
				{user.patronymic}
			</span>

			{#if user.group_code}
				<span class="text-md font-semibold text-primary">
					{user.group_code}
				</span>
			{/if}

			{#if user.phone_number}
				<span class="text-md text-muted-foreground">
					{user.phone_number}
				</span>
			{/if}

			<div>
				<a href="/user/details"><Button class="mt-6 py-5">Редактировать профиль</Button></a>
			</div>

			<nav
				class="mt-8 w-full overflow-hidden rounded-2xl border border-border/40 bg-card px-4 py-4 shadow-sm"
			>
				<ul class="flex flex-col gap-6 font-medium">
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
			<nav
				class="mt-4 w-full overflow-hidden rounded-2xl border border-border/40 bg-card px-4 py-4 shadow-sm"
			>
				<ul class="flex flex-col gap-6 font-medium">
					<li>
						<a href="/info" class="flex w-full items-center justify-between">
							<span class="flex items-center gap-2"> О сервисе </span>
							<ChevronRight class="h-5 w-5" />
						</a>
					</li>
					<li>
						<a href="/support" class="flex w-full items-center justify-between">
							<span class="flex items-center gap-2"> Техподдержка </span>
							<ChevronRight class="h-5 w-5" />
						</a>
					</li>
					<li>
						<a href="/donate" class="flex w-full items-center justify-between">
							<span class="flex items-center gap-2"> Поддержать разработчика </span>
							<ChevronRight class="h-5 w-5" />
						</a>
					</li>
				</ul>
			</nav>

			<div class="w-full">
				<Button variant="outline" class="text-md  mt-8 w-full py-6 font-semibold text-destructive">
					<span class="flex items-center gap-2">
						<LogOut class="h-5 w-5" />
						Выйти
					</span>
				</Button>
			</div>
		</div>
	{/if}
</div>
