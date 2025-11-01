-- Criar Sala Geral (Pública)
INSERT INTO rooms (id, name, description, is_public, max_members, owner_id, member_count, created_at, updated_at)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380b11',
    'Sala Geral',
    'Sala pública para todos os usuários.',
    true,
    100,
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- ID do Admin
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Adicionar Admin como 'owner' da Sala Geral
INSERT INTO room_members (room_id, user_id, role, joined_at)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', -- ID da Sala Geral
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- ID do Admin
    'owner',
    CURRENT_TIMESTAMP
);

-- Criar Sala da Admin (Privada)
INSERT INTO rooms (id, name, description, is_public, max_members, owner_id, member_count, created_at, updated_at)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380b12',
    'Sala da Admin',
    'Sala privada do Admin.',
    false,
    10,
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- ID do Admin
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Adicionar Admin como 'owner' da Sala Privada
INSERT INTO room_members (room_id, user_id, role, joined_at)
VALUES (
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380b12', -- ID da Sala Privada
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- ID do Admin
    'owner',
    CURRENT_TIMESTAMP
);