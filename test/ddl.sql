create table test.demo_student
(
    id          bigint unsigned auto_increment
        primary key,
    create_time datetime    default CURRENT_TIMESTAMP null,
    update_time datetime    default CURRENT_TIMESTAMP null,
    name        varchar(10) default ''                not null,
    sex         char        default '1'               not null,
    age         int         default 0                 not null,
    teacher_id  bigint                                null
)
    engine = InnoDB
    charset = utf8mb4;

create table test.demo_teacher
(
    id          bigint unsigned auto_increment
        primary key,
    create_time datetime    default CURRENT_TIMESTAMP null,
    update_time datetime    default CURRENT_TIMESTAMP null,
    name        varchar(10) default ''                not null,
    sex         char        default '1'               not null,
    age         int         default 0                 not null
)
    engine = InnoDB
    charset = utf8mb4;

