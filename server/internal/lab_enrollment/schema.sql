create schema if not exists lab_enrollment_service;

create type job_status as enum ('Queued', 'Processing', 'Completed');
create type job_result as enum ('Success', 'Failed');

create table if not exists lab_enrollment_service.jobs
(
    job_uuid          uuid       not null unique,
    user_uuid         uuid       not null,
    subscription_uuid uuid       not null,
    status            job_status not null default 'Queued',
    available_dates   jsonb      not null,
    started_at        timestamp,
    completed_at      timestamp,
    constraint jobs_pk primary key (job_uuid)
);

create table if not exists lab_enrollment_service.job_results
(
    job_uuid        uuid       not null unique,
    result          job_result not null,
    error_message   text,
    enrollment_uuid uuid,
    constraint job_results_pk primary key (job_uuid),
    constraint job_results_job_fk foreign key (job_uuid) references lab_enrollment_service.jobs (job_uuid)
);

create table if not exists lab_enrollment_service.enrollments
(
    enrollment_uuid      uuid        not null unique,
    user_uuid            uuid        not null,
    dikidi_enrollment_id int         not null,
    visit_time           timestamptz not null,
    enrolled_at          timestamp   not null,
    constraint enrollments_pk primary key (enrollment_uuid)
);