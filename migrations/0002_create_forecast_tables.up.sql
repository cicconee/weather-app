CREATE TABLE gridpoints(
    id SERIAL PRIMARY KEY,
    grid_id CHAR(3) NOT NULL,
    grid_x INTEGER NOT NULL,
    grid_y INTEGER NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    timezone TEXT NOT NULL,
    boundary POLYGON NOT NULL 
);

CREATE TABLE periods(
    num INTEGER NOT NULL,
    starts TIMESTAMPTZ NOT NULL,
    ends TIMESTAMPTZ NOT NULL,
    is_day_time BOOLEAN NOT NULL,
    temp INTEGER,
    temp_unit VARCHAR(255),
    wind_speed VARCHAR(255),
    wind_direction VARCHAR(255),
    short_forecast TEXT NOT NULL,
    gp_id INTEGER NOT NULL,
    PRIMARY KEY(num, gp_id),
    FOREIGN KEY(gp_id) REFERENCES gridpoints(id) ON DELETE CASCADE
);