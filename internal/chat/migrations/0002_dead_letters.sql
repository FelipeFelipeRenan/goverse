-- Mensagens que não puderam ser gravadas por motivo permanente (ex.: sala apagada enquanto estavam no spool).
-- Sem FK de propósito: a sala pode não existir mais. Reprocesse/descarte manualmente.
CREATE TABLE chat.dead_letters (
    id          uuid        PRIMARY KEY,
    room_id     uuid        NOT NULL,
    user_id     uuid        NOT NULL,
    username    text        NOT NULL,
    content     text        NOT NULL,
    sent_at     timestamptz NOT NULL,
    reason      text        NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now()
);
