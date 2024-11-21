USE tacacsdb;

CREATE TABLE logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    ip VARCHAR(15) NOT NULL,
    username VARCHAR(255) NOT NULL,
    interface VARCHAR(255) NOT NULL,
    client_ip VARCHAR(15) NOT NULL,
    action VARCHAR(255) NOT NULL
);

CREATE USER 'tacacsdb'@'%' IDENTIFIED BY 'tacacsdb#123db';

GRANT ALL PRIVILEGES ON tacacsdb.* TO 'tacacsdb'@'%';

FLUSH PRIVILEGES;

