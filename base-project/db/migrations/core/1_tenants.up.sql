CREATE TABLE tenants (
    id serial PRIMARY KEY,
    name citext NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX tenants_name_uindex ON tenants USING btree (name);
