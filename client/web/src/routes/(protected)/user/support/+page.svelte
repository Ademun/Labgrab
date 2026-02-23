<script lang="ts">
	import { Header } from '$lib/components/navigation/index.js';
	import { fly } from 'svelte/transition';
	import { SendHorizontal } from '@lucide/svelte';
	import Button from '$lib/components/ui/button/button.svelte';

	const faqs = [
		{
			q: 'На лабораторной был другой преподаватель — почему?',
			a: 'Сервис определяет преподавателя на основе общего расписания и запросов пользователей — это не гарантия того, кто фактически проведёт занятие. Если преподаватель не совпал, отправьте запрос на корректировку: добавьте или измените дату и время в расписании преподавателя в сервисе.'
		},
		{
			q: 'Есть предложение или жалоба',
			a: 'Напишите напрямую в техподдержку.'
		},
		{
			q: 'Можно ли отменить запись прямо из сервиса?',
			a: 'Да.'
		},
		{
			q: 'Может ли сервис записать меня на сегодня?',
			a: 'Нет. Сервис учитывает буфер в 48 часов — чтобы у вас было время подготовиться или при необходимости отменить запись.'
		}
	];
</script>

<div class="flex h-full w-full flex-col">
	<Header title="Техподдержка" back="/user" />

	<div class="flex flex-1 flex-col items-center px-6 py-8">
		<div class="flex w-full max-w-sm flex-col gap-8">
			<!-- Кнопка "Задать вопрос" -->
			<div in:fly={{ y: 16, duration: 300, delay: 40 }} class="flex w-full">
				<a href="https://t.me/Adremun" class="grow"
					><Button class="w-full py-6 uppercase">Задать вопрос</Button></a
				>
			</div>

			<!-- FAQ -->
			<div class="flex flex-col gap-4">
				{#each faqs as item, i}
					<div
						class="rounded-2xl border border-border/40 bg-card px-5 py-5 shadow-sm"
						in:fly={{ y: 12, duration: 260, delay: 120 + i * 60 }}
					>
						<p class="mb-2 text-base leading-snug font-semibold">{item.q}</p>
						<p class="text-sm leading-relaxed text-muted-foreground">{item.a}</p>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>
