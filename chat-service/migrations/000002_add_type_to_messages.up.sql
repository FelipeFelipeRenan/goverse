ALTER TABLE messages ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'CHAT';
CREATE INDEX idx_messages_type ON messages(type);