create schema if not exists subscription_service;

create type lab_type as enum ('Defence', 'Performance');
create type lab_topic as enum ('Virtual', 'Electricity', 'Mechanics', 'Optics', 'Rigid Body');
create type subscription_status as enum ('Active', 'Paused', 'Closed');
create type day_of_week as enum ('MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN');

create table if not exists subscription_service.subscriptions
(
    subscription_uuid uuid        not null unique,
    lab_type          lab_type    not null,
    lab_topic         lab_topic   not null,
    lab_number        int         not null,
    lab_auditorium    int,
    status            subscription_status      not null,
    auto_enroll       bool        not null,
    any_date          bool        not null,
    created_at        timestamptz not null,
    closed_at         timestamptz,
    user_uuid         uuid        not null,
    constraint subscriptions_pk primary key (lab_type, lab_topic, lab_number,
                                             user_uuid)
);

CREATE INDEX idx_subscriptions_lookup
    ON subscription_service.subscriptions (lab_type, lab_topic, lab_number, status)
    INCLUDE (lab_auditorium, user_uuid, any_date, auto_enroll);

create table if not exists subscription_service.time_preferences
(
    day_of_week day_of_week not null,
    lessons     int[]       not null,
    user_uuid   uuid        not null,
    constraint time_preferences_pk primary key (day_of_week, user_uuid)
);

CREATE INDEX idx_time_prefs_user_day
    ON subscription_service.time_preferences (user_uuid, day_of_week)
    INCLUDE (lessons);

create table if not exists subscription_service.teacher_preferences
(
    blacklisted_teachers text[] not null,
    user_uuid            uuid   not null,
    constraint teacher_preferences_pk primary key (user_uuid)
);

CREATE INDEX idx_teacher_prefs_user
    ON subscription_service.teacher_preferences (user_uuid)
    INCLUDE (blacklisted_teachers);

create table if not exists subscription_service.details
(
    successful_subscriptions     int  not null,
    last_successful_subscription timestamptz,
    user_uuid                    uuid not null,
    constraint details_pk primary key (user_uuid)
);

CREATE INDEX idx_details_user_stats
    ON subscription_service.details (user_uuid)
    INCLUDE (successful_subscriptions, last_successful_subscription);
