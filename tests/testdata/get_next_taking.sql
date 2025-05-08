SET @minute = 60000000000;

INSERT INTO schedule (id, user_id, name, end_at, period) VALUES (1, 1000000000000000, 'Test get_next_taking name1',   '2025-01-01', @minute * 60);
INSERT INTO schedule (id, user_id, name,         period) VALUES (2, 1000000000000000, 'Test get_next_taking name2' ,                @minute * 70);
INSERT INTO schedule (id, user_id, name, end_at, period) VALUES (3, 1000000000000000, 'Test get_next_taking name3',   '2025-01-02', @minute * 60 * 5);
INSERT INTO schedule (id, user_id, name, end_at, period) VALUES (4, 1000000000000000, 'Test get_next_taking expired', '2024-12-31', @minute * 60);
