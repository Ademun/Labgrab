# Dikidi — Get Records API Specification

> **Base URL:** `https://dikidi.net`  
> **Company ID:** `550001`  
> All requests require session cookies and standard browser-impersonation headers.

---

## Get Records

**`GET /ru/mobile/ajax/newrecord/get_records/`**

Возвращает текущие и прошлые записи пользователя.

### Query Parameters

|Parameter|Type|Value / Description|
|---|---|---|
|`company`|`string`|`550001`|
|`session`|`string`|User session|
|`social_key`|`string`|`""` (empty)|
|`fresh`|`string`|`both` — вернуть и новые и старые записи|

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`empty`|
|`Sec-Fetch-Mode`|`cors`|
|`Sec-Fetch-Site`|`same-origin`|
|`X-Requested-With`|`XMLHttpRequest`|
|`Referer`|`https://dikidi.net/550001?p=0.pi-ssm`|

### Response `200 OK`

```json
{
  "error": { "code": 0, "message": null },
  "data": {
    "new": {
      "more": false,
      "list": [
        {
          "id": "<record_id>",
          "type": "normal",
          "time": "2026-03-02 08:50:00",
          "time_to": "2026-03-02 10:20:00",
          "status": "1",
          "appointment_status": "1",
          "duration": "90",
          "duration_string": "1 час 30 минут",
          "services": [{ "id": "5298527", "name": "Лабораторная работа № 4 (Оптика)", "..." : "..." }],
          "employees": [{ "id": "1373891", "username": "Аудиторное занятие (233 ауд.) ...", "..." : "..." }],
          "showCancel": true
        }
      ]
    },
    "old": {
      "more": false,
      "list": [ "..." ]
    }
  }
}
```

|Field|Type|Description|
|---|---|---|
|`error.code`|`int`|`0` — успех, любое другое значение — ошибка API|
|`data.new.list`|`array`|Предстоящие записи (`status: "1"`)|
|`data.old.list`|`array`|Прошедшие/отменённые записи (`status: "2"`)|
|`data.*.more`|`bool`|Есть ли ещё записи (пагинация)|
|`list[].id`|`string`|ID записи|
|`list[].time`|`string`|Время начала — `YYYY-MM-DD HH:MM:SS`|
|`list[].time_to`|`string`|Время окончания — `YYYY-MM-DD HH:MM:SS`|
|`list[].showCancel`|`bool`|Доступна ли отмена записи|
|`list[].services`|`array`|Список услуг|
|`list[].employees`|`array`|Список исполнителей (аудитории)|