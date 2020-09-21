DROP TABLE IF EXISTS transactions, balance;

CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  requestId VARCHAR(36) UNIQUE,
  transactionId VARCHAR(36) UNIQUE,
  senderid VARCHAR(36),
  receiverid VARCHAR(36),
  message VARCHAR(128),
  amount FLOAT,
  currency VARCHAR(3),
  createdAt timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE balance (
  id SERIAL PRIMARY KEY,
  userId  VARCHAR(36) UNIQUE,
  amount FLOAT,
  lastTransactionId VARCHAR(36),
  updatedAt timestamp NOT NULL DEFAULT NOW()
);


INSERT into balance (userid, amount ) VALUES ('1', '1000');
INSERT into balance (userid, amount ) VALUES ('2', '0');
