-- Init SQL for Postgres container

-- Enable pgcrypto for gen_random_uuid() and crypt functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create schema and users table
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'user',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Insert example admin user (password: admin123) - hashed using crypt()
INSERT INTO users (username, email, password, role)
VALUES (
  'admin',
  'admin@example.com',
  crypt('admin123', gen_salt('bf')),
  'admin'
)
ON CONFLICT (username) DO NOTHING;

-- Example: create a sample database table for quizzes
CREATE TABLE IF NOT EXISTS quizzes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  description TEXT,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Example: create questions table
CREATE TABLE IF NOT EXISTS questions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quiz_id UUID REFERENCES quizzes(id) ON DELETE CASCADE,
  question TEXT NOT NULL,
  choices JSONB NOT NULL,
  answer INT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Commit
COMMIT;