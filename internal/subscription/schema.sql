create schema if not exists subscription_service;

create type lab_type as enum ('Defence', 'Performance');
create type lab_topic as enum ('Virtual', 'Electricity', 'Mechanics');

create table if not exists subscription_service.subscriptions
(
    subscription_uuid uuid        not null unique,
    lab_type          lab_type    not null,
    lab_topic         lab_topic   not null,
    lab_number        int         not null,
    lab_auditorium    int         ,
    created_at        timestamptz not null,
    closed_at         timestamptz,
    user_uuid         uuid        not null,
    constraint subscriptions_pk primary key (lab_type, lab_topic, lab_number, lab_auditorium,
                                             user_uuid)
);

create index if not exists subscriptions_search_idx on subscription_service.subscriptions (lab_type,
                                                                                           lab_topic,
                                                                                           lab_number,
                                                                                           lab_auditorium,
                                                                                           closed_at,
                                                                                           user_uuid);



create type day_of_week as enum ('MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN');

create table if not exists subscription_service.time_preferences
(
    day_of_week day_of_week not null,
    lessons     int[]       not null,
    user_uuid   uuid        not null,
    constraint time_preferences_pk primary key (day_of_week, user_uuid)
);

create index if not exists time_preferences_search_idx on subscription_service.time_preferences (day_of_week, user_uuid);

create table if not exists subscription_service.teacher_preferences
(
    blacklisted_teachers text[] not null,
    user_uuid            uuid   not null,
    constraint teacher_preferences_pk primary key (user_uuid)
);

create table if not exists subscription_service.details
(
    successful_subscriptions     int  not null,
    last_successful_subscription timestamptz,
    user_uuid                    uuid not null,
    constraint details_pk primary key (user_uuid)
);

