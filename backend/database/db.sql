-- -------------------------
-- Table: assets
-- -------------------------
CREATE TABLE IF NOT EXISTS assets (
    asset_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    CONSTRAINT assets_name_unique UNIQUE (name),
    CONSTRAINT assets_type_unique UNIQUE (type)
);

-- -------------------------
-- Table: users
-- -------------------------
CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL,
    auth_host VARCHAR(100) NOT NULL,
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_auth_host_unique UNIQUE (auth_host)
);

-- -------------------------
-- Table: transition
-- -------------------------
CREATE TABLE IF NOT EXISTS transition (
    trade_id SERIAL PRIMARY KEY,
    trade_type VARCHAR(10) NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    user_id INT NOT NULL,
    asset_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- ENUM replacement
    CONSTRAINT trade_type_enum CHECK (trade_type IN ('Buy', 'Sell')),

    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (asset_id) REFERENCES assets(asset_id)
);

-- -------------------------
-- Table: user_assets
-- -------------------------
CREATE TABLE IF NOT EXISTS user_assets (
    user_assets_id SERIAL PRIMARY KEY,
    asset_id INT NOT NULL,
    user_id INT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    price DOUBLE PRECISION NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (asset_id) REFERENCES assets(asset_id)
);
