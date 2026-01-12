# Визуализация работы поиска подписок на примере данных

## Исходные данные

Давайте представим, что в нашей системе есть следующие данные:

### Таблица subscriptions (подписки)

| subscription_uuid | lab_type | lab_topic | lab_number | lab_auditorium | user_uuid | closed_at |
|------------------|----------|-----------|------------|----------------|-----------|-----------|
| sub-001 | Defence | Virtual | 3 | 201 | user-alice | NULL |
| sub-002 | Defence | Virtual | 3 | 201 | user-bob | NULL |
| sub-003 | Defence | Virtual | 3 | 201 | user-charlie | NULL |
| sub-004 | Defence | Mechanics | 5 | 105 | user-dave | NULL |

### Таблица time_preferences (временные предпочтения)

| user_uuid | day_of_week | lessons |
|-----------|-------------|---------|
| user-alice | MON | [1, 2, 3] |
| user-alice | WED | [2, 4] |
| user-bob | MON | [1] |
| user-bob | TUE | [3, 4] |
| user-charlie | MON | [5, 6] |
| user-charlie | WED | [2] |

### Таблица teacher_preferences (чёрный список преподавателей)

| user_uuid | blacklisted_teachers |
|-----------|---------------------|
| user-alice | ["Ivanov"] |
| user-bob | ["Petrov", "Sidorov"] |
| user-charlie | [] |

### Таблица details (дополнительная информация)

| user_uuid | successful_subscriptions | last_successful_subscription |
|-----------|-------------------------|----------------------------|
| user-alice | 5 | 2025-01-10 |
| user-bob | 2 | 2025-01-05 |
| user-charlie | 0 | NULL |

### Входящие параметры запроса

Представим, что появились новые слоты для лабораторной Defence/Virtual/3/201 (это как раз соответствует первым трём подпискам):

**AvailableSlots:**
```json
{
  "MON": {
    "1": ["Ivanov", "Petrov"],
    "2": ["Sidorov", "Kozlov"],
    "3": ["Petrov"]
  },
  "WED": {
    "2": ["Ivanov", "Kozlov"]
  }
}
```

## Шаг 1: CTE available_slots_expanded

После выполнения первого CTE, который разворачивает JSON структуру, мы получаем плоскую таблицу:

| day_of_week | lesson | teachers |
|-------------|--------|----------|
| MON | 1 | ["Ivanov", "Petrov"] |
| MON | 2 | ["Sidorov", "Kozlov"] |
| MON | 3 | ["Petrov"] |
| WED | 2 | ["Ivanov", "Kozlov"] |

Здесь мы превратили вложенную структуру в четыре строки. Каждая строка представляет один конкретный временной слот с его преподавателями.

## Шаг 2: CTE matching_subscriptions (начало)

Сначала применяются WHERE условия к таблице subscriptions. Мы ищем подписки с параметрами Defence/Virtual/3/201, которые не закрыты:

**Отфильтрованные подписки:**

| subscription_uuid | user_uuid |
|-------------------|-----------|
| sub-001 | user-alice |
| sub-002 | user-bob |
| sub-003 | user-charlie |

Подписка sub-004 отфильтровалась, так как у неё другие параметры лабораторной (Mechanics/5/105).

## Шаг 3: CROSS JOIN с available_slots_expanded

Теперь делаем декартово произведение трёх подписок на четыре слота. Получаем 3 × 4 = 12 строк:

| subscription_uuid | user_uuid | day_of_week | lesson | teachers |
|------------------|-----------|-------------|--------|----------|
| sub-001 | user-alice | MON | 1 | ["Ivanov", "Petrov"] |
| sub-001 | user-alice | MON | 2 | ["Sidorov", "Kozlov"] |
| sub-001 | user-alice | MON | 3 | ["Petrov"] |
| sub-001 | user-alice | WED | 2 | ["Ivanov", "Kozlov"] |
| sub-002 | user-bob | MON | 1 | ["Ivanov", "Petrov"] |
| sub-002 | user-bob | MON | 2 | ["Sidorov", "Kozlov"] |
| sub-002 | user-bob | MON | 3 | ["Petrov"] |
| sub-002 | user-bob | WED | 2 | ["Ivanov", "Kozlov"] |
| sub-003 | user-charlie | MON | 1 | ["Ivanov", "Petrov"] |
| sub-003 | user-charlie | MON | 2 | ["Sidorov", "Kozlov"] |
| sub-003 | user-charlie | MON | 3 | ["Petrov"] |
| sub-003 | user-charlie | WED | 2 | ["Ivanov", "Kozlov"] |

Каждая подписка теперь "попробовала" соединиться с каждым слотом.

## Шаг 4: INNER JOIN с time_preferences

Теперь применяется первый фильтр - временные предпочтения. JOIN проверяет три условия одновременно:
- user_uuid совпадает
- day_of_week совпадает
- lesson входит в массив lessons

Давайте проверим каждую строку:

**Строка 1:** user-alice, MON, 1 → Alice имеет предпочтение MON [1,2,3], пара 1 входит → **ПРОХОДИТ** ✓

**Строка 2:** user-alice, MON, 2 → Alice имеет предпочтение MON [1,2,3], пара 2 входит → **ПРОХОДИТ** ✓

**Строка 3:** user-alice, MON, 3 → Alice имеет предпочтение MON [1,2,3], пара 3 входит → **ПРОХОДИТ** ✓

**Строка 4:** user-alice, WED, 2 → Alice имеет предпочтение WED [2,4], пара 2 входит → **ПРОХОДИТ** ✓

**Строка 5:** user-bob, MON, 1 → Bob имеет предпочтение MON [1], пара 1 входит → **ПРОХОДИТ** ✓

**Строка 6:** user-bob, MON, 2 → Bob имеет предпочтение MON [1], пара 2 НЕ входит → **ОТСЕИВАЕТСЯ** ✗

**Строка 7:** user-bob, MON, 3 → Bob имеет предпочтение MON [1], пара 3 НЕ входит → **ОТСЕИВАЕТСЯ** ✗

**Строка 8:** user-bob, WED, 2 → Bob НЕ имеет предпочтений для WED (только TUE) → **ОТСЕИВАЕТСЯ** ✗

**Строка 9:** user-charlie, MON, 1 → Charlie имеет предпочтение MON [5,6], пара 1 НЕ входит → **ОТСЕИВАЕТСЯ** ✗

**Строка 10:** user-charlie, MON, 2 → Charlie имеет предпочтение MON [5,6], пара 2 НЕ входит → **ОТСЕИВАЕТСЯ** ✗

**Строка 11:** user-charlie, MON, 3 → Charlie имеет предпочтение MON [5,6], пара 3 НЕ входит → **ОТСЕИВАЕТСЯ** ✗

**Строка 12:** user-charlie, WED, 2 → Charlie имеет предпочтение WED [2], пара 2 входит → **ПРОХОДИТ** ✓

После этого фильтра у нас остаётся только 5 строк из 12:

| subscription_uuid | user_uuid | day_of_week | lesson | teachers | blacklisted_teachers |
|------------------|-----------|-------------|--------|----------|---------------------|
| sub-001 | user-alice | MON | 1 | ["Ivanov", "Petrov"] | ["Ivanov"] |
| sub-001 | user-alice | MON | 2 | ["Sidorov", "Kozlov"] | ["Ivanov"] |
| sub-001 | user-alice | MON | 3 | ["Petrov"] | ["Ivanov"] |
| sub-001 | user-alice | WED | 2 | ["Ivanov", "Kozlov"] | ["Ivanov"] |
| sub-002 | user-bob | MON | 1 | ["Ivanov", "Petrov"] | ["Petrov", "Sidorov"] |
| sub-003 | user-charlie | WED | 2 | ["Ivanov", "Kozlov"] | [] |

Обратите внимание, я добавил колонку blacklisted_teachers - она появилась после JOIN с teacher_preferences.

## Шаг 5: Проверка EXISTS (фильтр по преподавателям)

Теперь для каждой оставшейся строки проверяем, есть ли хотя бы один преподаватель, который НЕ в чёрном списке:

**Строка 1:** user-alice, teachers=["Ivanov", "Petrov"], blacklist=["Ivanov"]
- Ivanov в чёрном списке? ДА → не подходит
- Petrov в чёрном списке? НЕТ → **подходит!** ✓

**Строка 2:** user-alice, teachers=["Sidorov", "Kozlov"], blacklist=["Ivanov"]
- Sidorov в чёрном списке? НЕТ → **подходит!** ✓

**Строка 3:** user-alice, teachers=["Petrov"], blacklist=["Ivanov"]
- Petrov в чёрном списке? НЕТ → **подходит!** ✓

**Строка 4:** user-alice, teachers=["Ivanov", "Kozlov"], blacklist=["Ivanov"]
- Ivanov в чёрном списке? ДА → не подходит
- Kozlov в чёрном списке? НЕТ → **подходит!** ✓

**Строка 5:** user-bob, teachers=["Ivanov", "Petrov"], blacklist=["Petrov", "Sidorov"]
- Ivanov в чёрном списке? НЕТ → **подходит!** ✓

**Строка 6:** user-charlie, teachers=["Ivanov", "Kozlov"], blacklist=[]
- Пустой чёрный список, все подходят → **подходит!** ✓

Все строки прошли проверку! Это результат CTE matching_subscriptions:

| subscription_uuid | user_uuid | successful_subscriptions | last_successful_subscription | day_of_week | lesson |
|------------------|-----------|-------------------------|----------------------------|-------------|--------|
| sub-001 | user-alice | 5 | 2025-01-10 | MON | 1 |
| sub-001 | user-alice | 5 | 2025-01-10 | MON | 2 |
| sub-001 | user-alice | 5 | 2025-01-10 | MON | 3 |
| sub-001 | user-alice | 5 | 2025-01-10 | WED | 2 |
| sub-002 | user-bob | 2 | 2025-01-05 | MON | 1 |
| sub-003 | user-charlie | 0 | NULL | WED | 2 |

## Шаг 6: CTE grouped_by_day

Теперь группируем пары обратно по дням. Используем GROUP BY по (user_uuid, subscription_uuid, день) и собираем уроки в массив:

**Для user-alice, подписка sub-001:**
- День MON: уроки [1, 2, 3] → собираются в JSON массив [1, 2, 3]
- День WED: урок [2] → собирается в JSON массив [2]

**Для user-bob, подписка sub-002:**
- День MON: урок [1] → собирается в JSON массив [1]

**Для user-charlie, подписка sub-003:**
- День WED: урок [2] → собирается в JSON массив [2]

Результат CTE grouped_by_day:

| user_uuid | subscription_uuid | successful_subscriptions | last_successful_subscription | day_of_week | lessons_array |
|-----------|------------------|-------------------------|----------------------------|-------------|---------------|
| user-alice | sub-001 | 5 | 2025-01-10 | MON | [1, 2, 3] |
| user-alice | sub-001 | 5 | 2025-01-10 | WED | [2] |
| user-bob | sub-002 | 2 | 2025-01-05 | MON | [1] |
| user-charlie | sub-003 | 0 | NULL | WED | [2] |

## Шаг 7: Финальный SELECT

На последнем шаге группируем по (user_uuid, subscription_uuid) и собираем все дни в один JSON объект:

**Для user-alice, подписка sub-001:**
- Дни: MON → [1,2,3], WED → [2]
- Собираются в объект: {"MON": [1, 2, 3], "WED": [2]}

**Для user-bob, подписка sub-002:**
- День: MON → [1]
- Собирается в объект: {"MON": [1]}

**Для user-charlie, подписка sub-003:**
- День: WED → [2]
- Собирается в объект: {"WED": [2]}

Затем сортируем по successful_subscriptions и last_successful_subscription:

## Финальный результат (после сортировки)

| user_uuid | subscription_uuid | successful_subscriptions | last_successful_subscription | matching_timeslots |
|-----------|------------------|-------------------------|----------------------------|-------------------|
| user-charlie | sub-003 | 0 | NULL | {"WED": [2]} |
| user-bob | sub-002 | 2 | 2025-01-05 | {"MON": [1]} |
| user-alice | sub-001 | 5 | 2025-01-10 | {"MON": [1, 2, 3], "WED": [2]} |

Обратите внимание на порядок:
1. Charlie идёт первым, потому что у него 0 успешных подписок (минимум)
2. Bob идёт вторым, у него 2 успешных подписки
3. Alice идёт последней, у неё 5 успешных подписок (максимум)

Если бы у Bob и Charlie было одинаковое количество успешных подписок, то Charlie всё равно был бы первым благодаря NULLS FIRST - у него last_successful_subscription = NULL, а у Bob есть дата.

## Резюме

Запрос последовательно:
1. Развернул входные слоты в плоскую таблицу (4 строки)
2. Соединил их с подписками через CROSS JOIN (12 строк)
3. Отфильтровал по временным предпочтениям (осталось 6 строк)
4. Отфильтровал по чёрному списку преподавателей (осталось 6 строк)
5. Сгруппировал пары обратно по дням (4 строки)
6. Сгруппировал дни обратно в объект (3 строки)
7. Отсортировал по приоритету (3 строки в правильном порядке)

Таким образом, из трёх подписок все три оказались подходящими, но с разными наборами подходящих временных слотов.