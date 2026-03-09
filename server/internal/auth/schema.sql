create schema if not exists auth_service;

create table if not exists auth_service.user_data
(
    user_uuid           uuid not null,
    dikidi_password     text not null,
    dikidi_phone_number text not null,
    dek                 text not null,
    session             text,
    token               text,
    cookies             text,
    api_authed          bool generated always as (
        session is not null and
        token is not null and
        cookies is not null
        ) stored,
    last_auth           timestamp,
    constraint user_data_pk primary key (user_uuid)
);
