create schema if not exists auth_service;

create table if not exists auth_service.user_data
(
    user_uuid           uuid not null,
    dikidi_phone_number text not null,
    dikidi_password     text not null,
    dek                 text not null,
    session             text,
    token               text,
    cookies             text,
    constraint user_data_pk primary key (user_uuid)
);
