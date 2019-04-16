-- текущие клиенты брокера

CREATE TABLE `clients` (
    `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `login_id` int NOT NULL,
    -- `password` varchar(300) NOT NULL,
    `balance` int NOT NULL
);

-- INSERT INTO `clients` (`id`, `login`,  `password`, `balance`) 
--     VALUES (1, 'Vasily', '123456', 200000),
--     VALUES (2, 'Ivan', 'qwerty', 200000),
--     VALUES (3, 'Olga', '1qaz2wsx', 200000);


CREATE TABLE `positions` (
    `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` int NOT NULL,
    `ticker` varchar(300) NOT NULL,
    `vol` int NOT NULL,
    KEY user_id(user_id)
);

-- INSERT INTO `clients` (`user_id`, `ticker`, `amount`) 
--     VALUES (1, 'SiM7', '123456', 200000),
--     VALUES (1, 'RIM7', '123456', 200000),
--     VALUES (2, 'RIM7', 'qwerty', 200000);

--от биржи    
CREATE TABLE `orders_history` (
    `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `time` int NOT NULL,
    `user_id` int,
    `ticker` varchar(300) NOT NULL,
    `vol` int NOT NULL,
    `price` float not null,
    `is_buy` int not null,
    KEY user_id(user_id)
);

--активные запросы сейчас
CREATE TABLE `request` ( -- запросы
    `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` int,
    `ticker` varchar(300) NOT NULL,
    `vol` int NOT NULL,
    `price` float NOT NULL,
    `is_buy` int not null, -- 1 - покупаем, 0 - продаем
    KEY user_id(user_id)
);

--история,обновляется раз в минуту
CREATE TABLE `stat` ( -- запросы
    `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `time` int,
    `interval` int,
    `open` float,
    `high` float,
    `low` float,
    `close` float,
    `volume` int,
    `ticker` varchar(300),
    KEY id(id)
);
