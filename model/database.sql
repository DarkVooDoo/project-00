CREATE EXTENSION btree_gist;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    firstname VARCHAR(40) NOT NULL,
    lastname VARCHAR(40) NOT NULL,
    password TEXT NOT NULL,
    phone VARCHAR,
    town VARCHAR,
    postal VARCHAR(6),
    geolocation POINT,
    email VARCHAR(50) UNIQUE NOT NULL,
    picture TEXT,
    salt INT NOT NULL,
    confirmed BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30),
    category_vector TSVECTOR GENERATED ALWAYS AS(to_tsvector('french', name)) STORED
);

CREATE TYPE payment_type AS ENUM('Espéce', 'CB', 'Chéque', 'Paypal');

CREATE TABLE etablishment (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    search_vector TSVECTOR GENERATED ALWAYS AS(to_tsvector('french',name)) STORED,
    phone VARCHAR(13),
    adresse VARCHAR(150) NOT NULL,
    postal INT NOT NULL,
    payment payment_type[],
    geolocation POINT,
    instagram VARCHAR(30),
    created_at date DEFAULT NOW(),
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    category_id INT REFERENCES category(id)
);

CREATE TABLE schedule(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    day INT NOT NULL CHECK(day BETWEEN 0 AND 6),
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    etablishment_id BIGINT REFERENCES etablishment(id)
);

CREATE INDEX idx_name ON etablishment(name);

CREATE TABLE service (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(150) NOT NULL,
    duration INT NOT NULL,
    price MONEY NOT NULL,
    discount INT DEFAULT 0,
    etablishment_id BIGINT REFERENCES etablishment(id)
);

CREATE TYPE employee_role AS ENUM ('Admin', 'Employee');

CREATE TABLE employee (
    id BIGSERIAL PRIMARY KEY,
    schedule JSONB,
    role employee_role,
    joined DATE DEFAULT NOW(),
    etablishment_id BIGINT REFERENCES etablishment(id),
    user_id BIGINT REFERENCES users(id),
    CONSTRAINT unique_etablishment_employee UNIQUE(etablishment_id, user_id)
);

CREATE TYPE appointment_status AS ENUM ('Confirmé', 'Terminé', 'Annulé');

CREATE TABLE appointment (
    id BIGSERIAL PRIMARY KEY,
    "date" TSRANGE,
    total MONEY,
    status appointment_status,
    user_id BIGINT REFERENCES users(id),
    etablishment_id BIGINT REFERENCES etablishment(id),
    employee_id BIGINT REFERENCES employee(id),
    EXCLUDE USING GIST (date WITH &&, employee_id WITH =) WHERE (status != 'Confirmé')
);

CREATE TABLE appointment_service(
    service_id BIGINT REFERENCES service(id) ON DELETE CASCADE,
    appointment_id BIGINT REFERENCES appointment(id) ON DELETE CASCADE,
    PRIMARY KEY(appointment_id, service_id) 
);

CREATE TABLE review (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    comment TEXT,
    rating INT CHECK(rating <= 5 AND rating >= 0),
    created_at DATE DEFAULT NOW(),
    etablishment_id BIGINT REFERENCES etablishment(id),
    user_id BIGINT REFERENCES users(id),
    employee_id BIGINT REFERENCES employee(id)
);

/*CREATE TABLE message (
  id BIGSERIAL PRIMARY KEY,
  msg text,
  created_at date,
  etablishment_id bigint,
  from_id bigint,
  to_id bigint
);

*/
