create schema if not exists user_service;

create table if not exists user_service.users
(
    uuid         uuid   not null,
    name         text,
    surname      text,
    patronymic   text,
    group_code   text,
    phone_number text,
    telegram_id  bigint not null,
    username     text   not null,
    photo_url    text,
    constraint user_pk primary key (uuid)
);