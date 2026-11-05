CREATE TABLE identity.users (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      text        NOT NULL,
    email         text        NOT NULL,
    password_hash text        NOT NULL DEFAULT '',
    picture       text        NOT NULL DEFAULT '',
    is_oauth      boolean     NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);

-- E-mail único sem diferenciar maiúsculas; e-mails de contas removidas ficam livres.
CREATE UNIQUE INDEX users_email_key ON identity.users (lower(email)) WHERE deleted_at IS NULL;
