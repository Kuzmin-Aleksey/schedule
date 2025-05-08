SET @minute = 60000000000;

INSERT INTO schedule (id, user_id, name, end_at, period) VALUES (1, 1000000000000000, 'Test get_schedule name',   '2025-01-01', @minute * 120);
