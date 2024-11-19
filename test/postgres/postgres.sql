CREATE TABLE employee
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name       varchar(10),
    sex        char,
    age        int                default 0,
    leader_id  bigint    null
);