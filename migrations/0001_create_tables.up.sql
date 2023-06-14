CREATE TABLE states (
    id CHAR(2) PRIMARY KEY,
    total_zones INTEGER,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE state_zones (
    id SERIAL PRIMARY KEY,
    uri TEXT NOT NULL,
    code CHAR(6) NOT NULL,
    type VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    effective_date TIMESTAMPTZ NOT NULL,
    state CHAR(2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY(state) REFERENCES states(id) ON DELETE CASCADE
);

CREATE TABLE state_zone_perimeters (
    id SERIAL PRIMARY KEY,
    sz_id INTEGER NOT NULL,
    boundary POLYGON NOT NULL,
    FOREIGN KEY(sz_id) REFERENCES state_zones(id) ON DELETE CASCADE
);

CREATE TABLE state_zone_holes (
    id SERIAL PRIMARY KEY,
    zp_id INTEGER NOT NULL,
    boundary POLYGON NOT NULL,
    FOREIGN KEY(zp_id) REFERENCES state_zone_perimeters(id) ON DELETE CASCADE
);

CREATE TABLE alerts (
    id TEXT PRIMARY KEY,
    area_desc TEXT NOT NULL,
    onset TIMESTAMPTZ,
    expires TIMESTAMPTZ NOT NULL,
    ends TIMESTAMPTZ,
    message_type TEXT NOT NULL,
    category TEXT NOT NULL,
    severity TEXT NOT NULL,
    certainty TEXT NOT NULL,
    urgency TEXT NOT NULL,
    event TEXT NOT NULL,
    headline TEXT,
    description TEXT NOT NULL,
    instruction TEXT,
    response TEXT NOT NULL,
    boundary POLYGON,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE alert_zones (
    alert_id TEXT NOT NULL,
    sz_id INTEGER NOT NULL,
    FOREIGN key(alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    FOREIGN key(sz_id) REFERENCES state_zones(id) ON DELETE CASCADE,
    PRIMARY key(alert_id, sz_id)
);

CREATE TABLE lonely_alerts (
    alert_id TEXT NOT NULL,
    sz_uri TEXT NOT NULL,
    FOREIGN KEY(alert_id) REFERENCES alerts(id) ON DELETE CASCADE,
    PRIMARY KEY(alert_id, sz_uri)
);