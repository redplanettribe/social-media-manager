CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "username" varchar(50) NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "password_hash" varchar(255) NOT NULL,
  "salt" varchar(32) NOT NULL,
  "created_at" timestamp DEFAULT (NOW()),
  "updated_at" timestamp DEFAULT (NOW())
);