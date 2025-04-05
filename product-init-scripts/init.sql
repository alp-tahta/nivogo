CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(100) NOT NULL,
    price VARCHAR(100) NOT NULL
);

CREATE INDEX ON products (id);

INSERT INTO products (name, description, price) VALUES 
    ('Bird s Nest Fern', 'The Bird s Nest Fern is a tropical plant known for its vibrant green, wavy fronds...',22),
    ('Ctenanthe', 'The Ctenanthe, also known as the Prayer Plant, is a stunning tropical plant with bold...',45);

