-- \c postgres; -- Conecta ao banco 'postgres' padrão para rodar os comandos

SELECT 'CREATE DATABASE users'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'users')\gexec

SELECT 'CREATE DATABASE rooms'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'rooms')\gexec

SELECT 'CREATE DATABASE messages'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'messages')\gexec