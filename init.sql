-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    uuid VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    telegram_id INTEGER UNIQUE,
    balance INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы столов
CREATE TABLE IF NOT EXISTS tables (
    id SERIAL PRIMARY KEY,
    category VARCHAR(10) NOT NULL CHECK (category IN ('LOW', 'MID', 'VIP')),
    blinds VARCHAR(20) NOT NULL,
    buy_in INTEGER NOT NULL,
    players INTEGER DEFAULT 0,
    max_seats INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы игроков за столами
CREATE TABLE IF NOT EXISTS table_players (
    id SERIAL PRIMARY KEY,
    table_id INTEGER REFERENCES tables(id) ON DELETE CASCADE,
    user_uuid VARCHAR(36) REFERENCES users(uuid) ON DELETE CASCADE,
    seat_number INTEGER,
    chips INTEGER DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_id, seat_number),
    UNIQUE(table_id, user_uuid)
);

-- Создание таблицы игр
CREATE TABLE IF NOT EXISTS games (
    id VARCHAR(36) PRIMARY KEY,
    table_id INTEGER REFERENCES tables(id) ON DELETE CASCADE,
    state VARCHAR(20) DEFAULT 'waiting',
    deck JSONB,
    community_cards JSONB DEFAULT '[]',
    pot INTEGER DEFAULT 0,
    current_bet INTEGER DEFAULT 0,
    dealer_position INTEGER DEFAULT 0,
    current_player INTEGER DEFAULT 0,
    small_blind INTEGER,
    big_blind INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы игроков в игре
CREATE TABLE IF NOT EXISTS game_players (
    id SERIAL PRIMARY KEY,
    game_id VARCHAR(36) REFERENCES games(id) ON DELETE CASCADE,
    user_uuid VARCHAR(36) REFERENCES users(uuid) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    cards JSONB DEFAULT '[]',
    chips INTEGER DEFAULT 0,
    bet INTEGER DEFAULT 0,
    is_folded BOOLEAN DEFAULT FALSE,
    is_all_in BOOLEAN DEFAULT FALSE,
    last_action VARCHAR(10),
    UNIQUE(game_id, user_uuid),
    UNIQUE(game_id, position)
);

-- Создание таблицы действий игроков
CREATE TABLE IF NOT EXISTS game_actions (
    id SERIAL PRIMARY KEY,
    game_id VARCHAR(36) REFERENCES games(id) ON DELETE CASCADE,
    user_uuid VARCHAR(36) REFERENCES users(uuid) ON DELETE CASCADE,
    action VARCHAR(10) NOT NULL,
    amount INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Вставка тестовых столов
INSERT INTO tables (category, blinds, buy_in, players, max_seats) VALUES
('LOW', '1/2', 50, 0, 6),
('MID', '5/10', 200, 0, 9),
('VIP', '25/50', 1000, 0, 6)
ON CONFLICT DO NOTHING;

-- Обновляем счетчик игроков в существующих столах
UPDATE tables SET players = (
    SELECT COUNT(*) FROM table_players WHERE table_players.table_id = tables.id
);

-- Создание индексов для оптимизации
CREATE INDEX IF NOT EXISTS idx_tables_category ON tables(category);
CREATE INDEX IF NOT EXISTS idx_table_players_table_id ON table_players(table_id);
CREATE INDEX IF NOT EXISTS idx_table_players_user_uuid ON table_players(user_uuid);
CREATE INDEX IF NOT EXISTS idx_games_table_id ON games(table_id);
CREATE INDEX IF NOT EXISTS idx_games_state ON games(state);
CREATE INDEX IF NOT EXISTS idx_game_players_game_id ON game_players(game_id);
CREATE INDEX IF NOT EXISTS idx_game_players_user_uuid ON game_players(user_uuid);
CREATE INDEX IF NOT EXISTS idx_game_actions_game_id ON game_actions(game_id);

-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tables_updated_at BEFORE UPDATE ON tables
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_games_updated_at BEFORE UPDATE ON games
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();