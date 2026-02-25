# Dikidi — Enrollment API Specification

> **Base URL:** `https://dikidi.net`  
> **Company ID:** `550001`  
> All requests require session cookies and standard browser-impersonation headers.

---

## Flow Overview

```
1. Reserve Slot          GET  /ru/ajax/newrecord/time_reservation/
2. Pre-submission Check  POST /ru/mobile/newrecord/check/
3. Verify Reservation    GET  /ru/mobile/ajax/newrecord/records_info/
4. Create Record         POST /ru/ajax/newrecord/record/
```

---

## Step 1 — Reserve Slot

**`GET /ru/ajax/newrecord/time_reservation/`**

Reserves a 1.5-hour slot. Returns the reservation ID used in all subsequent steps.

### Query Parameters

|Parameter|Type|Value / Description|
|---|---|---|
|`company_id`|`string`|`550001`|
|`master_id`|`int`|Level-2 card ID (master/specialist)|
|`services_id[]`|`int`|Level-1 card ID (service)|
|`time`|`string`|Target time — `YYYY-MM-DD HH:MM:SS`|
|`action_source`|`string`|`direct_link`|
|`session`|`string`|User session|

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`empty`|
|`Sec-Fetch-Mode`|`cors`|
|`Sec-Fetch-Site`|`same-origin`|
|`X-Requested-With`|`XMLHttpRequest`|
|`Referer`|`https://dikidi.net/550001?p=2.pi-ssm-sd&o=7&s={services_id}&rl=0_undefined`|

### Response `200 OK`

```json
{
  "record_id": <record_id>,
  "master_id": "1516167",
  "duration_string": "1 час 30 минут"
}
```

|Field|Type|Description|
|---|---|---|
|`record_id`|`int`|Reservation ID — used downstream|
|`master_id`|`string`|Level-2 card ID (echoed back)|
|`duration_string`|`string`|Human-readable slot duration|

---

## Step 2 — Pre-submission Check

**`POST /ru/mobile/newrecord/check/`**

Validates user data before the final record submission.

### Query Parameters

|Parameter|Type|Value / Description|
|---|---|---|
|`company`|`string`|`550001`|
|`session`|`string`|User session|
|`social_key`|`string`|`""` (empty)|

### Form Data

|Field|Type|Value / Description|
|---|---|---|
|`company`|`string`|`550001`|
|`type`|`string`|`normal`|
|`session`|`string`|User session|
|`social_key`|`string`|`""` (empty)|
|`share_id`|`int`|`0`|
|`phone`|`string`|Phone number, digits only|
|`first_name`|`string`|First name + patronymic|
|`last_name`|`string`|Last name|
|`comments`|`string`|Study group|
|`promocode_appointment_id`|`string`|`""` (empty)|

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`empty`|
|`Sec-Fetch-Mode`|`cors`|
|`Sec-Fetch-Site`|`same-site`|
|`Origin`|`https://dikidi.net`|
|`Referer`|`https://dikidi.net/550001?p=3.pi-ssm-sd-cf&o=7&m={master_id}&s={services_id}&d={YYYYMMDDHHmm}&r={record_id}&rl=0_{record_id}&sdr=`|

### Response `200 OK`

```json
{
  "error": 0,
  "warning": { "{record_id}": "..." },
  "...": "..."
}
```

|Field|Type|Description|
|---|---|---|
|`error`|`int`|`0` — успех, любое другое значение — ошибка API|

> Остальные поля в рамках флоу не используются.

---

## Step 3 — Verify Reservation

**`GET /ru/mobile/ajax/newrecord/records_info/`**

Fetches reservation details for display/validation before final submission.

### Query Parameters

|Parameter|Type|Value / Description|
|---|---|---|
|`company_id`|`string`|`550001`|
|`record_id_list[]`|`int`|Reservation ID (Step 1)|
|`session`|`string`|User session|

### Headers

Same as Step 2.

### Response `200 OK`

```json
{
  "error": { "code": 0, "message": null },
  "data": { "..." : "..." }
}
```

|Field|Type|Description|
|---|---|---|
|`error.code`|`int`|`0` — успех, любое другое значение — ошибка API|

> Остальные поля в рамках флоу не используются.

---

## Step 4 — Create Record

**`POST /ru/ajax/newrecord/record/`**

Final submission. Creates the appointment under the reserved slot ID.

### Query Parameters

|Parameter|Type|Value / Description|
|---|---|---|
|`company_id`|`string`|`550001`|
|`session`|`string`|User session|
|`social_key`|`string`|`""` (empty)|
|`action`|`string`|`send_code_info_continue_1`|
|`unique_num`|`int`|`1`|

### Form Data

|Field|Type|Value / Description|
|---|---|---|
|`type`|`string`|`normal`|
|`name`|`string`|First name + patronymic|
|`first_name`|`string`|First name + patronymic|
|`last_name`|`string`|Last name|
|`phone`|`string`|Phone number, digits only|
|`code`|`string`|`""` (empty)|
|`comments`|`string`|Study group|
|`is_show_all_times`|`int`|`3`|
|`captcha_token`|`string`|`""` (empty)|
|`action_source`|`string`|`direct link`|
|`session`|`string`|User session|
|`social_key`|`string`|`""` (empty)|
|`active_cart_id`|`int`|`0`|
|`active_method`|`int`|`0`|
|`agreement`|`int`|`1`|

### Headers

Same as Step 2.

> **Note:** The record is created under the `record_id` from Step 1.

### Response `200 OK`

```json
{
  "bookings": [
    {
      "id": "<record_id>",
      "status": "1",
      "appointment_status": "1",
      "time": "2026-03-02 08:50:00",
      "time_to": "2026-03-02 10:20:00",
      "duration_string": "1 час 30 минут",
      "...": "..."
    }
  ]
}
```

|Field|Type|Description|
|---|---|---|
|`bookings`|`array`|Список созданных записей; пустой массив — запись не создана|
|`bookings[].status`|`string`|`"1"` — запись успешно создана ⚠️ точное условие ошибки не верифицировано|

---

## Shared Context

### User Identity Fields

|Field|Source in codebase|Description|
|---|---|---|
|`phone`|DB — `dikidi_phone_number`|Digits only (sanitized)|
|`first_name`|DB — `first_name`|First name + patronymic|
|`last_name`|DB — `last_name`|Last name|
|`comments`|DB — `comments`|Study group|
|`session`|Decrypted from `Session`|Per-user session string|

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