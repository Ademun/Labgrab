create schema if not exists lab_enrollment

CREATE TABLE IF NOT EXISTS lab_enrollment.user_data (
    user_uuid uuid not null,
    dikidi_phone_number text not null,
    dikidi_password text not null,
    password_dek text not null,
    session_cookie text,
    noise_cookies text,
    CONSTRAINT user_data_pk PRIMARY KEY (user_uuid)
);