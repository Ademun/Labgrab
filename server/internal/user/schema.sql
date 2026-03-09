create schema if not exists user_service;

create table if not exists user_service.users
(
    uuid               uuid   not null,
    name               text,
    surname            text,
    patronymic         text,
    group_code         text,
    phone_number       text,
    telegram_id        bigint not null,
    telegram_username  text   not null,
    telegram_photo_url text,
    api_ready          bool generated always as (
        name is not null and
        surname is not null and
        patronymic is not null and
        group_code is not null and
        phone_number is not null
        ) stored,
    constraint user_pk primary key (uuid)
);