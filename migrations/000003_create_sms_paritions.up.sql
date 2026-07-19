DO
$$
DECLARE
start_date date := CURRENT_DATE;
    end_date   date := CURRENT_DATE + INTERVAL '3 months';
    d          date;
BEGIN
    d := start_date;

    WHILE d < end_date LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS sms_%s PARTITION OF sms
             FOR VALUES FROM (%L) TO (%L)',
            to_char(d, 'YYYY_MM_DD'),
            d,
            d + INTERVAL '1 day'
        );

        d := d + INTERVAL '1 day';
END LOOP;
END;
$$;