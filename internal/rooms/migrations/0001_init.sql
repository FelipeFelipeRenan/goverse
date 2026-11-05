-- Depende de identity.users (o módulo identity migra antes: ordem de registro no main).
CREATE TABLE rooms.rooms (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name         text        NOT NULL,
    description  text        NOT NULL DEFAULT '',
    is_public    boolean     NOT NULL DEFAULT true,
    max_members  integer     NOT NULL DEFAULT 10 CHECK (max_members > 0),
    owner_id     uuid        NOT NULL REFERENCES identity.users (id),
    member_count integer     NOT NULL DEFAULT 0 CHECK (member_count >= 0),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz,
    CONSTRAINT member_count_within_max CHECK (member_count <= max_members)
);

CREATE TABLE rooms.members (
    room_id   uuid        NOT NULL REFERENCES rooms.rooms (id) ON DELETE CASCADE,
    user_id   uuid        NOT NULL REFERENCES identity.users (id),
    role      text        NOT NULL CHECK (role IN ('owner', 'admin', 'moderator', 'member', 'guest')),
    joined_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX members_user_idx ON rooms.members (user_id);
-- Uma sala tem exatamente um dono.
CREATE UNIQUE INDEX one_owner_per_room ON rooms.members (room_id) WHERE role = 'owner';
CREATE INDEX rooms_public_idx ON rooms.rooms (created_at DESC) WHERE is_public AND deleted_at IS NULL;
