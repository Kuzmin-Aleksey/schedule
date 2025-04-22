DO $$
DECLARE
	names TEXT[] = ARRAY['James', 'Mary', 'John', 'Patricia', 'Robert'];
    last_names TEXT[] = ARRAY['Smith', 'Johnson', 'Williams', 'Brown', 'Jones'];
    departments TEXT[] = ARRAY['backend', 'frontend', 'ios', 'android'];
BEGIN
    FOR i IN 1..10000 LOOP
        INSERT INTO developers (name, department, geolocation, last_known_ip, is_available)
        VALUES (
                   names[1 + floor((random() * array_length(names, 1)))::int] || ' ' ||
                   last_names[1 + floor((random() * array_length(last_names, 1)))::int],
                   departments[1 + floor((random() * array_length(departments, 1)))::int],
                   POINT(round((random()*180-90)::numeric, 6), round((random()*360-180)::numeric, 6)),
                   CONCAT(
                           FLOOR(RANDOM() * 256), '.' ,
                           FLOOR(RANDOM() * 256), '.',
                           FLOOR(RANDOM() * 256), '.',
                           FLOOR(RANDOM() * 256))::inet,
                   random()>0.5
               );
    END LOOP;
END $$;