CREATE SEQUENCE IF NOT EXISTS weatherID;

CREATE TABLE IF NOT EXISTS weatherIDs (
  weather TEXT,
  id NUMERIC
);
INSERT INTO weatherIDs(id, weather) VALUES(nextval('weatherid'), 'UNKNOWN');

ALTER TABLE solar ADD COLUMN IF NOT EXISTS weatherID int;
UPDATE solar SET weatherid = 1 WHERE weatherid IS NULL;
