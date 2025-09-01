CREATE EXTENSION btree_gist;

ALTER DATABASE appointment SET TIMEZONE TO 'Europe/Paris';

CREATE FUNCTION max_etablishment() RETURNS TRIGGER LANGUAGE PLPGSQL
AS
$$
DECLARE
  etablishment_count int;
  is_premium BOOL;
BEGIN
  SELECT COUNT(e.id), ispremium FROM etablishment AS e LEFT JOIN users AS u ON u.id=NEW.user_id WHERE e.user_id=NEW.user_id GROUP BY u.ispremium INTO etablishment_count, is_premium;
  IF NOT is_premium AND etablishment_count > 0 THEN
    RAISE 'Max free tier reached: %', etablishment_count;
  ELSEIF is_premium AND etablishment_count > 2 THEN
    RAISE 'Max premium tier reached: %', etablishment_count;
  END IF;
  RETURN NEW;
END;
$$;

CREATE FUNCTION send_review_request() RETURNS TRIGGER LANGUAGE PLPGSQL AS $$
BEGIN
  IF NEW.status = 'Terminé' AND NEW.status != OLD.status AND NEW.user_id IS NOT NULL THEN
    INSERT INTO review (appointment_id, etablishment_id, user_id, employee_id) VALUES(NEW.id, NEW.etablishment_id, NEW.user_id, NEW.employee_id);
  END IF;
  RETURN NEW;
END;
$$;

CREATE FUNCTION service_exist_within_etablishment() RETURNS TRIGGER LANGUAGE PLPGSQL AS $$
DECLARE
    exist BOOL;
    e_id BIGINT;
BEGIN
    SELECT etablishment_id FROM appointment AS a WHERE a.id=NEW.appointment_id INTO e_id;
    SELECT EXISTS(SELECT 1 FROM service AS s WHERE s.id=NEW.service_id AND e_id=s.etablishment_id) INTO exist;
    IF NOT exist THEN
        RAISE 'Le service ne pas dans letablishment: %', NEW.service_id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION GeolocationDistance (geolocation1 POINT, geolocation2 POINT) 
RETURNS FLOAT language plpgsql AS $$
  DECLARE
    result FLOAT;
  BEGIN
    result = 60 * 1.1515 * (180/PI()) * ACOS(
      SIN(geolocation1[0] * (PI()/180)) * SIN(geolocation2[0] * (PI()/180)) + 
      COS(geolocation1[0] * (PI()/180)) * COS(geolocation2[0] * (PI()/180)) *
      COS((geolocation1[1] - geolocation2[1]) * (PI()/180))
    );
    RETURN result * 1.609344;
END;
$$;

CREATE FUNCTION GetAvaileblesDates(employeeId BIGINT, fromDate date) RETURNS time[] AS $$
DECLARE
    emp record;
    timer timestamp;
    times time[] = '{}';
    started time;
    ended time;
BEGIN
    SELECT CAST(e.schedule->'from'->>EXTRACT(ISODOW FROM DATE (fromDate))::integer - 1 AS time), CAST(e.schedule->'to'->>EXTRACT(ISODOW FROM DATE (fromDate))::integer - 1 AS time) - '30 minute'::INTERVAL 
    INTO started, ended FROM employee AS e WHERE id=employeeId;
    FOR timer IN SELECT * FROM generate_series(CONCAT(fromDate, ' ', started)::timestamp, CONCAT(fromDate, ' ',ended)::timestamp, '30 minute'::INTERVAL)
    LOOP
        IF TIMER > NOW() THEN
            times := array_append(times, timer::time);
        END IF;
        FOR emp IN SELECT "date" FROM appointment WHERE employee_id=employeeId AND LOWER("date")::DATE = fromDate
        LOOP
          IF emp.date @> timer THEN 
            times := array_remove(times, timer::time);
          END IF;
        END LOOP; 
    END LOOP;
    RETURN times;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION CreateGoogleAccountOrSignin(u_email TEXT, u_name TEXT) RETURNS 
TABLE (userId BIGINT, userEmail VARCHAR(70), shortName TEXT, employeeId BIGINT, etablishmentId BIGINT) LANGUAGE PLPGSQL AS $$
DECLARE
BEGIN
    IF NOT EXISTS(SELECT 1 FROM users WHERE email=u_email) THEN
        RETURN QUERY INSERT INTO users(email, firstname, is_google) VALUES(u_email, u_name, true) RETURNING id, email, CONCAT(LEFT(firstname, 1), LEFT(lastname, 1)), 0::BIGINT, 0::BIGINT;
    ELSE
        RETURN QUERY SELECT u.id, u.email, CONCAT(LEFT(u.firstname, 1), LEFT(u.lastname, 1)), COALESCE((SELECT e.id FROM employee AS e WHERE e.user_id=u.id LIMIT 1), 0),
        COALESCE((SELECT et.id FROM  etablishment AS et WHERE et.user_id=u.id LIMIT 1), 0) FROM users AS u WHERE u.email=u_email;
    END IF;
END;
$$;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    firstname VARCHAR(40) NOT NULL,
    lastname VARCHAR(40),
    password TEXT,
    phone VARCHAR,
    town VARCHAR,
    postal VARCHAR(6),
    geolocation POINT,
    email VARCHAR(70) UNIQUE NOT NULL,
    picture TEXT,
    salt INT,
    refresh_token TEXT,
    is_google BOOL DEFAULT false,
    confirmed BOOL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    ispremium BOOL DEFAULT false
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

CREATE TRIGGER user_max_etablishment BEFORE INSERT ON etablishment FOR EACH ROW EXECUTE FUNCTION max_etablishment();

CREATE TABLE schedule(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    day INT NOT NULL CHECK(day BETWEEN 0 AND 6),
    open_time TIME NOT NULL,
    close_time TIME,
    etablishment_id BIGINT REFERENCES etablishment(id),
    EXCLUDE USING BTREE(day WITH =, etablishment_id WITH =)
);

CREATE TABLE service (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(150) NOT NULL,
    duration INT NOT NULL,
    price MONEY NOT NULL,
    discount INT DEFAULT 0 CHECK(discount BETWEEN 0 AND 100),
    etablishment_id BIGINT REFERENCES etablishment(id)
);

CREATE TYPE employee_role AS ENUM ('Admin', 'Employee');

CREATE TABLE employee (
    id BIGSERIAL PRIMARY KEY,
    schedule JSONB,
    role employee_role,
    joined DATE DEFAULT NOW(),
    is_active BOOL DEFAULT false,
    etablishment_id BIGINT REFERENCES etablishment(id),
    user_id BIGINT REFERENCES users(id),
    CONSTRAINT unique_etablishment_employee UNIQUE(etablishment_id, user_id)
);

CREATE TYPE appointment_status AS ENUM ('Attente', 'Confirmé', 'Terminé', 'Annulé');

CREATE TABLE appointment (
    id BIGSERIAL PRIMARY KEY,
    "date" TSRANGE,
    name VARCHAR(70),
    phone VARCHAR(10),
    status appointment_status,
    user_id BIGINT REFERENCES users(id),
    etablishment_id BIGINT REFERENCES etablishment(id),
    employee_id BIGINT REFERENCES employee(id),
    EXCLUDE USING GIST (date WITH &&, employee_id WITH =) WHERE (status = 'Confirmé')
);

CREATE TRIGGER send_request AFTER UPDATE ON appointment FOR EACH ROW EXECUTE FUNCTION send_review_request();

CREATE TABLE appointment_service(
    service_id BIGINT REFERENCES service(id) ON DELETE CASCADE,
    appointment_id BIGINT REFERENCES appointment(id) ON DELETE CASCADE,
    PRIMARY KEY(appointment_id, service_id) 
);

CREATE TRIGGER service_exist_in_etablisment BEFORE INSERT ON appointment_service FOR EACH ROW EXECUTE FUNCTION service_exist_within_etablishment();

CREATE TABLE review (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    comment TEXT,
    review_key UUID DEFAULT gen_random_uuid(),
    rating INT CHECK(rating <= 5 AND rating >= 0),
    created_at DATE DEFAULT NOW(),
    expire_in DATE DEFAULT NOW() + INTERVAL '30 days',
    appointment_id BIGINT REFERENCES appointment(id),
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
