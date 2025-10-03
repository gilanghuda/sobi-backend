CREATE TABLE user_ahli (
    uid UUID PRIMARY KEY,
    price DECIMAL(10,2),
    category VARCHAR(50) DEFAULT 'ahli agama',
    open_time TIME ,
    rating DECIMAL(2,1) DEFAULT 0.0,
    CONSTRAINT fk_users FOREIGN KEY (uid) REFERENCES users(uid) ON DELETE CASCADE
);
