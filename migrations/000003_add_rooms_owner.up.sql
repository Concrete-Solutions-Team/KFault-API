ALTER TABLE rooms
ADD owner_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL;

CREATE INDEX idx_rooms_owner_id ON rooms(owner_id);