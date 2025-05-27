/*Insert file */

ALTER DATABASE appointment SET TIMEZONE TO 'Europe/Paris';

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

CREATE FUNCTION SignUser(userEmail VARCHAR) RETURNS TABLE (id BIGINT, shortname TEXT, picture TEXT, employee BIGINT, etablishment BIGINT, salt INT, password TEXT) AS $$
BEGIN
    RETURN QUERY SELECT u.id, LEFT(u.firstname, 1) || LEFT(u.lastname, 1), COALESCE(u.picture, ''), COALESCE((SELECT e.id FROM employee AS e WHERE e.user_id=u.id LIMIT 1), 0), 
    COALESCE((SELECT et.id FROM  etablishment AS et WHERE et.user_id=u.id LIMIT 1), 0), u.salt, u.password FROM users AS u WHERE u.email=userEmail AND u.confirmed=true;
END;
$$ LANGUAGE plpgsql;

INSERT INTO users(firstname, lastname, email, password, confirmed, salt) VALUES('John', 'Doe', 'john@test.com', 'Test12345!', true, 4321), 
('Alice', 'Price', 'alice@test.com', 'Localhost232!', true, 6521), ('Inés', 'Narayanaiken', 'ines@test.com', 'testskdqsd', true, 1243);

INSERT INTO category (name) VALUES('Manicure'), ('Barber'), ('Coiffeur'), ('Spa'), ('Institut Beauté'), ('Autre');

INSERT INTO etablishment(name, adresse, postal, schedule, payment, lat, lon, user_id, category_id) VALUES('Momo nails', '7 Rue Matignon', 75002, 
    '{"from": ["09:00", "09:00", "09:00", "09:00", "09:00", "", ""], "to": ["17:00", "17:00", "17:00", "17:00", "17:00", "", ""]}', '{"Espéce", "CB"}', 48.860826, 2.344729, 1, 1), 
('Gringo style', '39 Rue Martyr', 75018, 
    '{"from": ["09:00", "09:00", "09:00", "09:00", "09:00", "", ""], "to": ["17:00", "17:00", "17:00", "17:00", "17:00", "", ""]}', '{"Espéce", "Paypal"}', 48.864581, 2.352282, 1, 2);

INSERT INTO service (name, price, duration, description, etablishment_id) VALUES('Coupe', '20', 30, 'Hello Descruiption pour le service', 2), 
('Coupe + Barbe', '25', 45, 'Je sais pas quoi dire de la description', 2), ('Massage', '35', 50, 'no se q decir en esta descriptcion pero ahi esta', 1), 
('Manicure', '15', 25, 'Mama Seigneur petite descirption', 1);

INSERT INTO employee(schedule, etablishment_id, user_id) VALUES('{"from": ["09:00", "10:00", "10:00", "", "", "", ""], "to": ["17:00", "18:00", "17:00", "", "", "", ""]}', 2, 2), 
('{"from": ["13:00", "13:00", "13:00", "", "", "", ""], "to": ["20:00", "19:00", "20:00", "", "", "", ""]}', 1, 3), ('{"from": ["09:00", "10:00", "10:00", "", "", "", ""], "to": ["17:00", "18:00", "17:00", "", "", "", ""]}', 1, 2);

INSERT INTO appointment("date", status, user_id, etablishment_id, employee_id) VALUES('[2025-03-10 10:00, 2025-03-10 11:00)', 'Terminé', 1, 2, 1), 
('[2025-05-23 15:00, 2025-05-23 15:30)', 'Confirmé', 1, 2, 1), ('[2025-04-22 11:00, 2025-04-22 12:00)', 'Annulé', 1, 1, 2);

INSERT INTO appointment_service(service_id, appointment_id) VALUES(1, 1), (1,2), (2,2), (4, 3), (3, 3);
