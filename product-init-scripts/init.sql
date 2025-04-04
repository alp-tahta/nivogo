CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(50) NOT NULL,
    price VARCHAR(100) NOT NULL
);

CREATE INDEX ON products (id);

INSERT INTO products (name, description, price) VALUES 
    ('test1', 'description1',10),
    ('test2', 'description2',20);

