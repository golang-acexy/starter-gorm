drop table if exists employee;
CREATE TABLE employee
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name       VARCHAR(10),
    sex        CHAR(1)            default '0',
    age        INTEGER,
    leader_id  integer[]
);