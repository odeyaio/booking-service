CREATE TABLE IF NOT EXISTS role (
  id SMALLINT PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE
);
INSERT INTO role
VALUES (1, 'admin'),
  (2, 'user') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS booking_status (
  id SMALLINT PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE
);
INSERT INTO booking_status
VALUES (1, 'active'),
  (2, 'cancelled') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS "user" (
  id UUID PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role_id SMALLINT NOT NULL REFERENCES role(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO "user" (id, email, password_hash, role_id)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@example.com',
    '$2a$10$placeholder_admin_hash',
    1
  ),
  (
    '00000000-0000-0000-0000-000000000002',
    'user@example.com',
    '$2a$10$placeholder_user_hash',
    2
  ) ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS room (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  capacity INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS schedule (
  id UUID PRIMARY KEY,
  room_id UUID NOT NULL UNIQUE REFERENCES room(id) ON DELETE CASCADE,
  days_of_week SMALLINT [] NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  CONSTRAINT uq_schedule_id_room UNIQUE (id, room_id)
);

CREATE TABLE IF NOT EXISTS slot (
  id UUID PRIMARY KEY,
  room_id UUID NOT NULL,
  schedule_id UUID NOT NULL,
  start TIMESTAMPTZ NOT NULL,
  "end" TIMESTAMPTZ NOT NULL,
  CONSTRAINT fk_slot_schedule_room
    FOREIGN KEY (schedule_id, room_id)
    REFERENCES schedule(id, room_id)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS booking (
  id UUID PRIMARY KEY,
  slot_id UUID NOT NULL REFERENCES slot(id),
  user_id UUID NOT NULL REFERENCES "user"(id),
  status_id SMALLINT NOT NULL REFERENCES booking_status(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_slot_room_start ON slot(room_id, start);

CREATE INDEX IF NOT EXISTS idx_booking_user ON booking(user_id);

CREATE INDEX IF NOT EXISTS idx_booking_slot ON booking(slot_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_active_booking_slot ON booking(slot_id) WHERE status_id = 1;
