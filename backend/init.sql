-- backend/init.sql

-- 1. ตาราง Users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. ตาราง Games
CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    player1_id INT REFERENCES users(id),
    player2_id INT REFERENCES users(id),
    current_turn_id INT REFERENCES users(id),
    board VARCHAR(9) DEFAULT '---------', -- เก็บเป็น 'XOXO----'
    status VARCHAR(20) DEFAULT 'WAITING', -- WAITING, IN_PROGRESS, FINISHED, DRAW
    winner_id INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. ตาราง Moves (สำคัญมากสำหรับการทำ Replay และกัน Race Condition)
CREATE TABLE IF NOT EXISTS moves (
    id SERIAL PRIMARY KEY,
    game_id INT REFERENCES games(id) ON DELETE CASCADE,
    player_id INT REFERENCES users(id),
    x INT NOT NULL CHECK (x >= 0 AND x <= 2),
    y INT NOT NULL CHECK (y >= 0 AND y <= 2),
    move_order INT NOT NULL, -- ลำดับการเดิน
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- ห้ามลงซ้ำช่องเดิมในเกมเดียวกัน (Database Level Protection)
    CONSTRAINT unique_move_per_cell UNIQUE (game_id, x, y)
);