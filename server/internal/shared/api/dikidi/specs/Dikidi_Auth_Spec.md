# Dikidi — Auth API Specification

> **Base URL:** `https://dikidi.net` / `https://auth.dikidi.net`  
> **Company ID:** `550001`  
> Authentication flow создаёт новый HTTP-клиент без предварительных кук.  
> После успешной авторизации сохраняются куки: `cid`, `lang`, `cookie_name`, `token`.

---

## Flow Overview

```
1. Fetch Main Page     GET  https://dikidi.net/550001?p=0.pi-ssm
2. Acquire CSRF Token  POST https://auth.dikidi.net/ajax/check/auth/
3. Send Auth Request   POST https://auth.dikidi.net/ajax/user/auth/
```

После Step 3 из куки `cookie_name` извлекается сессия:  
`cookie_name = f129621b...~4d31c435...` → сессия = строка после `~`

---

## Step 1 — Fetch Main Page

**`GET https://dikidi.net/550001?p=0.pi-ssm`**

Загружает главную страницу компании. Из HTML извлекается `telegram_csrf`.

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`document`|
|`Sec-Fetch-Mode`|`navigate`|
|`Sec-Fetch-Site`|`none`|
|`Sec-Fetch-User`|`?1`|

### Parsing

Из HTML извлекается значение атрибута `value` элемента:

```html
<input type="hidden" name="telegram_csrf" value="91d23e7a517...~1770803949" />
```

---

## Step 2 — Acquire CSRF Token

**`POST https://auth.dikidi.net/ajax/check/auth/`**

Отправляет номер телефона и `telegram_csrf`. В ответе возвращается HTML-фрагмент формы с токеном `csrf`.

### Form Data

|Field|Value / Description|
|---|---|
|`telegram_csrf`|Токен из Step 1|
|`number`|Номер телефона, только цифры|

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`empty`|
|`Sec-Fetch-Mode`|`cors`|
|`Sec-Fetch-Site`|`same-site`|
|`Origin`|`https://dikidi.net`|
|`Referer`|`https://dikidi.net/550001?p=0.pi-ssm`|

### Response `200 OK`

```json
{
  "container": ".modal.auth .pocket",
  "html": "...<input type=\"hidden\" name=\"csrf\" value=\"c06d1fce19d7...\" />..."
}
```

|Field|Type|Description|
|---|---|---|
|`html`|`string`|HTML-фрагмент с `<input name="csrf" .../>`|

Из `html` извлекается `csrf`:

```html
<input type="hidden" name="csrf" value="<csrf_token>" />
```

---

## Step 3 — Send Auth Request

**`POST https://auth.dikidi.net/ajax/user/auth/`**

Финальная отправка данных авторизации.

### Form Data

|Field|Value / Description|
|---|---|
|`telegram_csrf`|Токен из Step 1|
|`number`|Номер телефона, только цифры|
|`csrf`|Токен из Step 2|
|`password`|Пароль пользователя|
|`pdAgreement`|`1`|

### Headers

|Header|Value|
|---|---|
|`Sec-Fetch-Dest`|`empty`|
|`Sec-Fetch-Mode`|`cors`|
|`Sec-Fetch-Site`|`same-site`|
|`Origin`|`https://dikidi.net`|
|`Referer`|`https://dikidi.net/550001?p=0.pi-ssm`|

### Post-auth Cookies

После успешного ответа из jar HTTP-клиента извлекаются следующие куки:

|Cookie|Домен|Описание|
|---|---|---|
|`cookie_name`|`dikidi.net` / `auth.dikidi.net`|Содержит session ID после `~`|
|`token`|`dikidi.net` / `auth.dikidi.net`|Auth token|
|`cid`|`dikidi.net`|Client ID|
|`lang`|`dikidi.net`|Язык интерфейса|

**Извлечение session ID:**

```
cookie_name = "<session_token>~<session_token>"
session     = "<session_token>"  // строка после ~
```