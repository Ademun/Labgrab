create schema if not exists record_service;

create type record_service.record_status as enum ('Open', 'Closed');

create table if not exists record_service.records
(
    record_id      int                          not null unique,
    lab_type       lab_type                     not null,
    lab_topic      lab_topic                    not null,
    lab_auditorium int                          not null,
    lesson         int                          not null,
    start_time     timestamptz                  not null,
    end_time       timestamptz                  not null,
    status         record_service.record_status not null default 'Open',
    user_uuid      uuid                         not null,
    constraint records_pk primary key (record_id)
)