/*Insert file */

INSERT INTO users(firstname, lastname, email, password, confirmed, salt, ispremium) VALUES('John', 'Doe', 'john@test.com', 'Test12345!', true, 4321, true), 
('Alice', 'Price', 'alice@test.com', 'Localhost232!', true, 6521, false), ('Inés', 'Narayanaiken', 'ines@test.com', 'testskdqsd', true, 1243, false);

INSERT INTO category (name) VALUES('Manicure'), ('Barber'), ('Coiffeur'), ('Spa'), ('Institut Beauté'), ('Autre');

INSERT INTO etablishment(name, adresse, postal, payment, geolocation, user_id, category_id) VALUES
('Momo nails', '7 Rue Matignon', 75002, '{"Espéce", "CB"}', POINT(48.860826, 2.344729), 1, 1), 
('Gringo style', '39 Rue Martyr', 75018, '{"Espéce", "Paypal"}', POINT(48.864581, 2.352282), 1, 2),
('Its just a test', '39 Rue Martyr', 75018, '{"Espéce"}', POINT(48.852849, 2.293166), 2, 3),
('Barberia', '39 Rue Martyr', 75018, '{"Espéce"}', POINT(48.855132, 2.306712), 1, 2);

INSERT INTO schedule (day, open_time, close_time, etablishment_id) VALUES
(0, '10:00', '14:00', 1), (1, '20:00', '23:00', 1), (2, '08:00', '13:00', 1), (3, '14:00', '19:00', 1), (4, '12:00', '17:00', 1), (5, '12:00', NULL, 1), (6, '20:00', NULL, 1),
(0, '16:00', '20:00', 2), (1, '10:00', '20:00', 2), (4, '09:00', '19:00', 2), (3, '09:00', '19:00', 2),  (5, '10:00', '20:00', 2), 
(2, '15:00', '19:00', 3);

INSERT INTO service (name, price, duration, description, discount, etablishment_id) VALUES('Coupe', '20', 30, 'Hello Descruiption pour le service', 0, 2), 
('Coupe + Barbe', '25', 45, 'Je sais pas quoi dire de la description', 0, 2), ('Massage', '35', 50, 'no se q decir en esta descriptcion pero ahi esta', 10, 1), 
('Manicure', '15', 25, 'Mama Seigneur petite descirption', 0, 1);

INSERT INTO employee(schedule, joined, etablishment_id, user_id, is_active) VALUES
('{"from": ["09:00", "10:00", "10:00", "", "", "", ""], "to": ["17:00", "18:00", "17:00", "", "", "", ""]}','2024-04-10', 2, 2, true), 
('{"from": ["13:00", "13:00", "13:00", "13:00", "", "", ""], "to": ["20:00", "19:00", "20:00", "17:00", "", "", ""]}', '2021-01-23', 1, 3, true), 
('{"from": ["09:00", "10:00", "10:00", "", "", "", ""], "to": ["17:00", "18:00", "17:00", "", "", "", ""]}', '2025-03-10', 1, 2, false);

INSERT INTO appointment("date", status, user_id, etablishment_id, employee_id) VALUES('[2025-08-03 10:00, 2025-08-03 11:00)', 'Terminé', 1, 2, 1), 
('[2025-08-03 15:00, 2025-08-03 15:30)', 'Confirmé', 1, 2, 1), 
('[2025-08-03 16:00, 2025-08-03 16:30)', 'Confirmé', 3, 2, 1), ('[2025-04-22 11:00, 2025-04-22 12:00)', 'Annulé', 1, 2, 2),
('[2025-12-25 11:00, 2025-12-25 13:00)', 'Confirmé', 3, 2, 1), ('[2025-12-25 07:00, 2025-12-25 08:00)', 'Terminé', 1, 1, 2);

INSERT INTO appointment_service(service_id, appointment_id) VALUES(1, 1), (1,2), (2,2), (2, 3), (2, 4), (1,5), (2,5), (3,6);

INSERT INTO review(comment, rating, etablishment_id, user_id) VALUES('Test', 3, 2, 1), ('Random Review', 5, 2, 3), ('No Way', 2, 2, 2);

UPDATE appointment SET status='Terminé' WHERE id=2;
