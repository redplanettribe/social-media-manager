
CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "password_hash" varchar(255) NOT NULL,
  "created_at" timestamp DEFAULT (NOW()),
  "updated_at" timestamp DEFAULT (NOW())
);

CREATE TABLE "teams" (
  "id" serial PRIMARY KEY,
  "name" varchar(100) NOT NULL,
  "created_by" integer NOT NULL,
  "created_at" timestamp DEFAULT (NOW()),
  "updated_at" timestamp DEFAULT (NOW())
);

CREATE TABLE "team_members" (
  "id" serial PRIMARY KEY,
  "team_id" integer NOT NULL,
  "user_id" integer NOT NULL,
  "role" varchar(20) NOT NULL,
  "added_at" timestamp DEFAULT (NOW())
);

CREATE TABLE "posts" (
  "id" serial PRIMARY KEY,
  "team_id" integer NOT NULL,
  "title" text NOT NULL,
  "content" text,
  "is_idea" boolean NOT NULL DEFAULT false,
  "status" varchar(20) NOT NULL DEFAULT 'draft',
  "scheduled_publish_date" timestamp,
  "created_by" integer NOT NULL,
  "created_at" timestamp DEFAULT (NOW()),
  "updated_at" timestamp DEFAULT (NOW()),
  "priority" integer NOT NULL DEFAULT 5
);

CREATE TABLE "comments" (
  "id" serial PRIMARY KEY,
  "post_id" integer NOT NULL,
  "user_id" integer NOT NULL,
  "content" text NOT NULL,
  "created_at" timestamp DEFAULT (NOW()),
  "updated_at" timestamp DEFAULT (NOW())
);

CREATE TABLE "social_networks" (
  "id" serial PRIMARY KEY,
  "name" varchar(50) UNIQUE NOT NULL
);

CREATE TABLE "post_social_networks" (
  "id" serial PRIMARY KEY,
  "post_id" integer NOT NULL,
  "social_network_id" integer NOT NULL,
  "scheduled_publish_date" timestamp,
  "status" varchar(20) NOT NULL DEFAULT 'scheduled'
);

CREATE TABLE "settings" (
  "id" serial PRIMARY KEY,
  "team_id" integer NOT NULL,
  "key" varchar(50) NOT NULL,
  "value" text NOT NULL
);

CREATE UNIQUE INDEX ON "team_members" ("team_id", "user_id");

CREATE UNIQUE INDEX ON "post_social_networks" ("post_id", "social_network_id");

CREATE UNIQUE INDEX ON "settings" ("team_id", "key");

ALTER TABLE "teams" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "team_members" ADD FOREIGN KEY ("team_id") REFERENCES "teams" ("id");

ALTER TABLE "team_members" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "posts" ADD FOREIGN KEY ("team_id") REFERENCES "teams" ("id");

ALTER TABLE "posts" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id");

ALTER TABLE "comments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "post_social_networks" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id");

ALTER TABLE "post_social_networks" ADD FOREIGN KEY ("social_network_id") REFERENCES "social_networks" ("id");

ALTER TABLE "settings" ADD FOREIGN KEY ("team_id") REFERENCES "teams" ("id");