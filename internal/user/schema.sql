create schema if not exists user_service;

create table if not exists user_service.users
(
    uuid uuid not null,
    constraint user_pk primary key (uuid)
);

create table if not exists user_service.users_details
(
    name       text not null,
    surname    text not null,
    patronymic text,
    group_code text not null,
    user_uuid  uuid not null,
    constraint users_details_pk primary key (user_uuid),
    constraint users_details_fk foreign key (user_uuid) references user_service.users (uuid) match simple on delete cascade on update cascade
);

create table if not exists user_service.users_contacts
(
    phone_number text not null,
    telegram_id  bigint not null ,
    user_uuid    uuid not null,
    constraint users_contacts_pk primary key (user_uuid),
    constraint users_contacts_fk foreign key (user_uuid) references user_service.users (uuid) match simple on delete cascade on update cascade
)