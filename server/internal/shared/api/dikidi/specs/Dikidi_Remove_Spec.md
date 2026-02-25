# Dikidi — Remove Record API Specification

> **Base URL:** `https://dikidi.net`  
> **Company ID:** `550001`  
> All requests require session cookies and standard browser-impersonation headers.

---

## Remove Record

**`GET /ru/mobile/newrecord/remove_record/`**

Отменяет запись пользователя.

### Query Parameters

|Parameter|Type|Value / Description|
|---|---|---|
|`id`|`string`|ID записи — `data.new.list[].id` из get_records|
|`session`|`string`|User session|
|`social_key`|`string`|`""` (empty)|

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`empty`|
|`Sec-Fetch-Mode`|`cors`|
|`Sec-Fetch-Site`|`same-origin`|
|`X-Requested-With`|`XMLHttpRequest`|
|`Referer`|`https://dikidi.net/550001?p=1.pi-ur`|

### Response `200 OK`

```json
{
  "error": 0,
  "record": {
    "_sale": null,
    "allowModify": true
  }
}
```

|Field|Type|Description|
|---|---|---|
|`error`|`int`|`0` — успех, любое другое значение — ошибка API|