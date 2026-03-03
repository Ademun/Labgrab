create schema if not exists booking_service;

create type booking_service.booking_status as enum ('Open', 'Closed');

create table if not exists booking_service.bookings
(
    booking_id int                            not null unique,
    type       lab_type                       not null,
    topic      lab_topic                      not null,
    auditorium int                            not null,
    spot       int,
    lesson     int                            not null,
    start_time timestamp                      not null,
    end_time   timestamp                      not null,
    status     booking_service.booking_status not null default 'Open',
    user_uuid  uuid                           not null,
    constraint bookings_pk primary key (booking_id)
)