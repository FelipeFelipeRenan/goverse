-- Senha para todos: "senha123"
-- Hash Bcrypt: $2a$12$6hv4SoYmUps7kt0CL2Y0J.I2KUmywzJMdcz3zA5Zs8FE92A4ykWDi

INSERT INTO users (id, username, email, password, picture,is_oauth)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Admin Goverse',
    'admin@goverse.com',
    '$2a$12$6hv4SoYmUps7kt0CL2Y0J.I2KUmywzJMdcz3zA5Zs8FE92A4ykWDi',
    'picture2.com',
    false
);

INSERT INTO users (id, username, email, password,picture, is_oauth)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12',
    'Usuario Comum',
    'usuario@goverse.com',
    '$2a$12$6hv4SoYmUps7kt0CL2Y0J.I2KUmywzJMdcz3zA5Zs8FE92A4ykWDi',
    'picture.com',

    false
);