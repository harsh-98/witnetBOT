DROP TABLE IF EXISTS tblUsers;
DROP TABLE IF EXISTS userNodeMap;
DROP TABLE IF EXISTS reputation;
DROP TABLE IF EXISTS tblNodes;
DROP TABLE IF EXISTS blockchain;

CREATE TABLE tblUsers (
    UserID int,
	UserName varchar(255),
    FirstName varchar(255),
	LastName varchar(255),
	IsAdmin bool,
	PRIMARY KEY (UserID)
);
-- https://stackoverflow.com/questions/1049728/how-do-i-see-what-character-set-a-mysql-database-table-column-is
ALTER TABLE witnet.tblUsers MODIFY COLUMN UserName VARCHAR(255)
    CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE witnet.tblUsers MODIFY COLUMN FirstName VARCHAR(255)
    CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
ALTER TABLE witnet.tblUsers MODIFY COLUMN LastName VARCHAR(255)
    CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE tblNodes (
	NodeID  varchar(50),
	Active bool,
	Reputation float,
	PRIMARY KEY (NodeID)
);
CREATE TABLE reputation (
	ID INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
	NodeID  varchar(50),
	Reputation float,
	CreateAt TIMESTAMP DEFAULT now()
	-- FOREIGN KEY (NodeID) REFERENCES tblNodes(NodeID)
);

CREATE TABLE blockchain (
	Epoch INTEGER NOT NULL PRIMARY KEY,
	Miner  varchar(50),
	reward INTEGER,
	hash varchar(130)
);

CREATE TABLE lightBlockchain (
	latestEpoch INTEGER,
	Miner  varchar(50) NOT NULL PRIMARY KEY,
	reward float,
	blockCount INTEGER,
	lastXEpochs varchar(200)
);


CREATE TABLE userNodeMap (
	UserID int,
	NodeID  varchar(257),
	NodeName  varchar(257),
	CONSTRAINT Mapping PRIMARY KEY (UserID,NodeID),
	-- FOREIGN KEY (NodeID) REFERENCES tblNodes(NodeID),
	FOREIGN KEY (UserID) REFERENCES tblUsers(UserID)
);
ALTER TABLE witnet.userNodeMap MODIFY COLUMN NodeName VARCHAR(255)
    CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

-- ALTER TABLE blockchain ADD reward INTEGER;
-- table name is case-sensitive use camel case.

-- finding the foreign key
-- select COLUMN_NAME, CONSTRAINT_NAME, REFERENCED_COLUMN_NAME, REFERENCED_TABLE_NAME
--    from information_schema.KEY_COLUMN_USAGE
--    where TABLE_NAME = 'reputation';

-- drop the foreign key
--    ALTER TABLE reputation DROP FOREIGN KEY reputation_ibfk_1;
-- delete from reputation where CreateAt < DATE_SUB(NOW(), INTERVAL 24 HOUR);

-- insert into lightBlockchain(lastestEpoch, Miner, reward, blockCount, fiveEpochs) VALUES(1, "12", 10, 1, "1,2,3,4,5");
-- update lightBlockchain set  fiveEpochs = SUBSTRING_INDEX(fiveEpochs,',', -2);

-- select SUBSTRING_INDEX(lastXEpochs,',', -2) from lightBlockchain;

-- INSERT INTO lightBlockchain (Miner, latestEpoch, reward, blockCount, lastXEpochs) VALUES(?, ?, ?, ?, ?) 
-- ON DUPLICATE KEY UPDATE latestEpoch=?, reward=reward+?, blockCount =  blockCount + ?, lastXEpochs = CONCAT_WS(SUBSTRING_INDEX(lastXEpochs, ',', ?), ?, ',');

-- select * from
-- 	(select count(epoch), sum(reward) as blockCount from blockchain where Miner=?) as T1
-- 	inner join
-- 	(select group_concat(epoch) as epochs  from
-- 		(select * from blockchain where Miner=? order by   Epoch desc limit 5) as T) as T2 on true ;

-- leaderboard with most mined blocks per user
-- select sum(blockCount) as _count , u.UserID, u.UserName, u.firstname, (sum(blockCount)/(select sum(blockCount) from lightBlockchain)) as Percentage, count(*) as NodeNumber from tblUsers u  inner join userNodeMap n inner join lightBlockchain b where n.UserID = u.UserID and n.NodeID = b.Miner group by u.UserID  order by _count desc;
-- select sum(_count ) from U;

-- ALTER USER 'root'@'localhost' IDENTIFIED WITH mysql_native_password BY '';
-- -- https://stackoverflow.com/questions/43094726/your-password-does-not-satisfy-the-current-policy-requirements
-- SET GLOBAL validate_password_policy=LOW;

--  CREATE USER 'network'@'localhost' IDENTIFIED BY '';
-- GRANT ALL PRIVILEGES ON witnet.* TO 'network'@'localhost' WITH GRANT OPTION;

-- checking character set of the table column
-- SELECT character_set_name FROM information_schema.`COLUMNS` 
-- WHERE table_schema = "witnet"
--   AND table_name = "tblUsers"
--   AND column_name = "FirstName";