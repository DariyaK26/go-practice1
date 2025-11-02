
CREATE TABLE IF NOT EXISTS movies (
                                      id SERIAL PRIMARY KEY,
                                      title TEXT NOT NULL,
                                      year INT NOT NULL
);

CREATE TABLE IF NOT EXISTS actors (
                                      id SERIAL PRIMARY KEY,
                                      movie_id INT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
                                      name TEXT NOT NULL
);


INSERT INTO movies (title, year) VALUES
                                     ('Inception', 2010),
                                     ('The Dark Knight', 2008),
                                     ('Interstellar', 2014),
                                     ('Dune', 2021);

INSERT INTO actors (movie_id, name) VALUES
                                        (1, 'Leonardo DiCaprio'),
                                        (1, 'Joseph Gordon-Levitt'),
                                        (1, 'Elliot Page'),
                                        (1, 'Tom Hardy'),
                                        (1, 'Ken Watanabe'),
                                        (2, 'Christian Bale'),
                                        (2, 'Heath Ledger'),
                                        (2, 'Aaron Eckhart'),
                                        (3, 'Matthew McConaughey'),
                                        (3, 'Anne Hathaway'),
                                        (3, 'Jessica Chastain'),
                                        (4, 'Timothée Chalamet'),
                                        (4, 'Rebecca Ferguson');
