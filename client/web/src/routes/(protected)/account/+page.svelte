<script lang="ts">
	import { Fallback, Image, Root } from '$lib/components/ui/avatar';
	import {
		Ban,
		Calendar,
		ChartNoAxesCombined,
		ChevronRight,
		CircleDollarSign,
		CircleQuestionMark,
		Headset,
		LogOut,
		User
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { fade, scale } from 'svelte/transition';
	import {
		Action as AlertAction,
		Cancel as AlertCancel,
		Content as AlertContent,
		Description as AlertDescription,
		Root as AlertRoot,
		Title as AlertTitle,
		Trigger as AlertTrigger
	} from '$lib/components/ui/alert-dialog';

	import { fullName, userStore } from '$lib/stores/user.ts';
	import { toast } from '$lib/utils/toast';

	let isLoggingOut = $state<boolean>(false);

	async function handleLogout() {
		isLoggingOut = true;

		try {
			await userStore.logout();

			toast.success('Вы вышли из аккаунта');
		} catch (error) {
			console.error('Failed to logout:', error);
			toast.error('Произошла ошибка при выходе', {
				description: 'Попробуйте обновить страницу'
			});
		} finally {
			isLoggingOut = false;
		}
	}

	function navigateToSchedule() {
		toast.info('Раздел расписания в разработке');
	}

	function navigateToBlacklist() {
		toast.info('Черный список в разработке');
	}

	function navigateToStats() {
		toast.info('Статистика в разработке');
	}

	function navigateToAbout() {
		toast.info('Информация о сервисе в разработке');
	}

	function navigateToSupport() {
		toast.info('Техподдержка в разработке', {
			description: 'Пока можете написать в Telegram: @your_support_username'
		});
	}

	function navigateToDonate() {
		toast.info('Поддержка разработчика в разработке', {
			description: 'Спасибо за желание помочь проекту!'
		});
	}

	function navigateToEditProfile() {
		toast.info('Редактирование профиля в разработке');
	}
</script>

<div class="flex flex-col items-center w-full mx-auto px-8 py-8">
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
		}}
	>
		<Root class="w-25 h-25 outline-3 outline-offset-2 outline-primary my-6">
			{#if $userStore.data?.photo_url}
				<Image src={$userStore.data.photo_url} alt="Фото профиля" />
			{/if}
			<Fallback>
				<User />
			</Fallback>
		</Root>
	</div>

	{#if $userStore.data}
		<span
			class="text-lg font-bold"
			in:fade={{
				delay: 200,
				duration: 300
			}}
			out:fade={{
				duration: 250,
				delay: 0
			}}
		>
			{$fullName || $userStore.data.username}
		</span>

		{#if $userStore.data.group_code}
			<span
				class="text-md font-semibold text-primary"
				in:fade={{
					delay: 225,
					duration: 300
				}}
				out:fade={{
					duration: 250,
					delay: 0
				}}
			>
				{$userStore.data.group_code}
			</span>
		{/if}

		{#if $userStore.data.phone_number}
			<span
				class="text-md text-muted-foreground"
				in:fade={{
					delay: 250,
					duration: 300
				}}
				out:fade={{
					duration: 250,
					delay: 0
				}}
			>
				{$userStore.data.phone_number}
			</span>
		{/if}
	{/if}

	<div
		in:scale={{
			delay: 300,
			duration: 300,
			start: 0.9
		}}
		out:scale={{
			duration: 250,
			start: 0.9,
			delay: 0
		}}
	>
		<Button class="py-5 mt-6" onclick={navigateToEditProfile}>Редактировать профиль</Button>
	</div>

	<nav
		class="bg-card rounded-2xl w-full overflow-hidden
            shadow-[0_2px_8px_rgba(0,0,0,0.04),0_1px_2px_rgba(0,0,0,0.06)]
            border border-border/40
            hover:shadow-[0_4px_12px_rgba(0,0,0,0.06),0_2px_4px_rgba(0,0,0,0.08)]
            transition-shadow duration-200 px-4 py-4 mt-8"
		in:fade={{
			delay: 350,
			duration: 300
		}}
		out:fade={{
			duration: 250,
			delay: 50
		}}
	>
		<ul class="flex flex-col gap-6 font-medium">
			<li>
				<button
					class="flex justify-between items-center w-full hover:text-primary transition-colors"
					onclick={navigateToSchedule}
				>
					<span class="flex items-center gap-2">
						<Calendar class="w-5 h-5 text-[#007AFF]" />
						Расписание
					</span>
					<ChevronRight class="w-5 h-5" />
				</button>
			</li>
			<li>
				<button
					class="flex justify-between items-center w-full hover:text-primary transition-colors"
					onclick={navigateToBlacklist}
				>
					<span class="flex items-center gap-2">
						<Ban class="w-5 h-5  text-[#FF3B30]" />
						Черный список
					</span>
					<ChevronRight class="w-5 h-5" />
				</button>
			</li>
			<li>
				<button
					class="flex justify-between items-center w-full hover:text-primary transition-colors"
					onclick={navigateToStats}
				>
					<span class="flex items-center gap-2">
						<ChartNoAxesCombined class="w-5 h-5 text-[#34C759]" />
						Статистика
					</span>
					<ChevronRight class="w-5 h-5" />
				</button>
			</li>
		</ul>
	</nav>

	<nav
		class="bg-card rounded-2xl w-full overflow-hidden
            shadow-[0_2px_8px_rgba(0,0,0,0.04),0_1px_2px_rgba(0,0,0,0.06)]
            border border-border/40
            hover:shadow-[0_4px_12px_rgba(0,0,0,0.06),0_2px_4px_rgba(0,0,0,0.08)]
            transition-shadow duration-200 px-4 py-4 mt-4"
		in:fade={{
			delay: 400,
			duration: 300
		}}
		out:fade={{
			duration: 250,
			delay: 100
		}}
	>
		<ul class="flex flex-col gap-6 font-medium">
			<li>
				<button
					class="flex justify-between items-center w-full hover:text-primary transition-colors"
					onclick={navigateToAbout}
				>
					<span class="flex items-center gap-2"> О сервисе </span>
					<ChevronRight class="w-5 h-5" />
				</button>
			</li>
			<li>
				<button
					class="flex justify-between items-center w-full hover:text-primary transition-colors"
					onclick={navigateToSupport}
				>
					<span class="flex items-center gap-2"> Техподдержка </span>
					<ChevronRight class="w-5 h-5" />
				</button>
			</li>
			<li>
				<button
					class="flex justify-between items-center w-full hover:text-primary transition-colors"
					onclick={navigateToDonate}
				>
					<span class="flex items-center gap-2"> Поддержать разработчика </span>
					<ChevronRight class="w-5 h-5" />
				</button>
			</li>
		</ul>
	</nav>

	<div
		class="w-full"
		in:fade={{
			delay: 450,
			duration: 300
		}}
		out:fade={{
			duration: 250,
			delay: 150
		}}
	>
		<AlertRoot>
			<AlertTrigger class="w-full">
				<Button
					variant="outline"
					class="text-destructive text-md font-semibold w-full py-6 mt-8
                        hover:bg-destructive hover:text-destructive-foreground"
					disabled={isLoggingOut}
				>
					{#if isLoggingOut}
						<span class="flex items-center gap-2">
							<span class="animate-spin">⏳</span>
							Выход...
						</span>
					{:else}
						<span class="flex items-center gap-2">
							<LogOut class="w-5 h-5" />
							Выйти
						</span>
					{/if}
				</Button>
			</AlertTrigger>
			<AlertContent>
				<AlertTitle>Выйти из аккаунта?</AlertTitle>
				<AlertDescription>
					Вам придётся войти заново через Telegram для доступа к сервису.
				</AlertDescription>
				<div class="flex justify-between gap-3 mt-4">
					<AlertCancel class="flex-1">Отмена</AlertCancel>
					<AlertAction class="flex-1" onclick={handleLogout}>Да, выйти</AlertAction>
				</div>
			</AlertContent>
		</AlertRoot>
	</div>
</div>
