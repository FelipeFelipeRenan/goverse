-- Depende de rooms.rooms (rooms migra antes: ordem de registro no main).
-- id é UUIDv7 gerado pelo servidor: ordena por tempo, então o índice (room_id, id) serve ao histórico.
CREATE TABLE chat.messages (
    id         uuid        PRIMARY KEY,
    room_id    uuid        NOT NULL REFERENCES rooms.rooms (id) ON DELETE CASCADE,
    user_id    uuid        NOT NULL,
    username   text        NOT NULL,
    content    text        NOT NULL CHECK (length(content) > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX messages_room_id_idx ON chat.messages (room_id, id DESC);
