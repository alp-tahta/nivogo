CREATE TABLE IF NOT EXISTS inventory (
    product_id integer NOT NULL,
    quantity integer NOT NULL
);

CREATE INDEX ON inventory (product_id);

INSERT INTO inventory (product_id, quantity) VALUES 
    (1,10),
    (2,20);

