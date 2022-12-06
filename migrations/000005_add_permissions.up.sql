--Filename: migrations/000005_add_permissions.up.sql

CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);

--create a linking table that links users to permissions
--this is an example of a many to many relationships.