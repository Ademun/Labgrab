create schema if not exists lab_enrollment_service;

CREATE TABLE IF NOT EXISTS lab_enrollment_service.user_data
(
    user_uuid           uuid not null,
    dikidi_phone_number text not null,
    dikidi_password     text not null,
    dek                 text not null,
    session             text,
    token               text,
    cookies             text,
    CONSTRAINT user_data_pk PRIMARY KEY (user_uuid)
);