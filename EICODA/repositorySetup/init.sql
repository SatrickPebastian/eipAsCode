CREATE TABLE FilterTypes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    artifact VARCHAR(255),
    configs TEXT,
    derivedFrom VARCHAR(255)
);

