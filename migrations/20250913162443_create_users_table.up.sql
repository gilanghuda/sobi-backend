CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    uid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    phone_number VARCHAR(20) UNIQUE,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female')),
    avatar INT CHECK (avatar BETWEEN 1 AND 6), 
    password_hash TEXT NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    user_role VARCHAR(20) DEFAULT 'user',
    otp CHAR(4) , 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
