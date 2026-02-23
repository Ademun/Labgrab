<script lang="ts">
	import { Header } from '$lib/components/navigation/index.js';
	import { fly } from 'svelte/transition';
</script>

<div class="flex h-full w-full flex-col">
	<Header title="О сервисе" back="/user" />

	<div class="flex flex-1 flex-col items-center px-6 py-8">
		<div class="flex w-full max-w-sm flex-col gap-4">
			<section
				class="rounded-2xl border border-border/40 bg-card px-6 py-5 shadow-sm"
				in:fly={{ y: 16, duration: 300, delay: 60 }}
			>
				<h2 class="mb-3 text-sm font-semibold tracking-wide text-muted-foreground uppercase">
					Как это работает
				</h2>
				<p class="text-sm leading-relaxed">
					Сервис следит за расписанием лабораторных и сравнивает доступные места с вашими
					предпочтениями. Когда находит подходящее — либо записывает вас автоматически, либо
					отправляет уведомление, в зависимости от выбранного режима. Всё происходит без вашего
					участия.
				</p>
				<p class="mt-3 text-sm leading-relaxed text-muted-foreground">
					Сервис спроектирован так, чтобы работать аккуратно и не создавать лишней нагрузки на сайт.
				</p>
			</section>

			<section
				class="rounded-2xl border border-border/40 bg-card px-6 py-5 shadow-sm"
				in:fly={{ y: 16, duration: 300, delay: 120 }}
			>
				<h2 class="mb-3 text-sm font-semibold tracking-wide text-muted-foreground uppercase">
					Важно понимать
				</h2>
				<p class="mb-3 text-sm leading-relaxed">
					Это неофициальный инструмент, и университетская платформа не предназначена для подобного
					использования. Это означает реальные риски:
				</p>
				<ul class="flex flex-col gap-2 text-sm">
					<li class="flex items-start gap-2">
						<span class="mt-0.5 text-destructive">•</span>
						<span>Блокировка учётной записи</span>
					</li>
					<li class="flex items-start gap-2">
						<span class="mt-0.5 text-destructive">•</span>
						<span>Удаление существующих записей</span>
					</li>
					<li class="flex items-start gap-2">
						<span class="mt-0.5 text-destructive">•</span>
						<span>Ответственность за нарушение пользовательского соглашения</span>
					</li>
				</ul>
				<p class="mt-3 text-sm leading-relaxed text-muted-foreground">
					Сервис не гарантирует успешную запись. Используя его, вы осознаёте и принимаете эти риски.
				</p>
			</section>

			<section
				class="rounded-2xl border border-border/40 bg-card px-6 py-5 shadow-sm"
				in:fly={{ y: 16, duration: 300, delay: 180 }}
			>
				<h2 class="mb-3 text-sm font-semibold tracking-wide text-muted-foreground uppercase">
					Безопасность данных
				</h2>
				<p class="text-sm leading-relaxed">
					Все пароли и секреты шифруются на сервере и не передаются третьим лицам. При удалении
					аккаунта все ваши данные удаляются немедленно и полностью.
				</p>
				<p class="mt-3 text-sm leading-relaxed text-muted-foreground">
					Сервис работает без SLA и формальных гарантий — это инструмент, который поддерживается и
					развивается, но не коммерческий продукт.
				</p>
			</section>
		</div>
	</div>
</div>
