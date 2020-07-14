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

CREATE TABLE tblNodes (
	NodeID  varchar(50),
	Active bool,
	Reputation float,
	Blocks int,
	PRIMARY KEY (NodeID)
);
CREATE TABLE reputation (
	ID INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
	NodeID  varchar(50),
	Reputation float,
	CreateAt TIMESTAMP DEFAULT now(),
	FOREIGN KEY (NodeID) REFERENCES tblNodes(NodeID)
);

CREATE TABLE blockchain (
	Epoch INTEGER NOT NULL PRIMARY KEY,
	Miner  varchar(50),
	reward INTEGER,
	hash varchar(130)
);

CREATE TABLE userNodeMap (
	UserID int,
	NodeID  varchar(257),
	NodeName  varchar(257),
	CONSTRAINT Mapping PRIMARY KEY (UserID,NodeID),
	-- FOREIGN KEY (NodeID) REFERENCES tblNodes(NodeID),
	FOREIGN KEY (UserID) REFERENCES tblUsers(UserID)
);

-- ALTER TABLE blockchain ADD reward INTEGER;