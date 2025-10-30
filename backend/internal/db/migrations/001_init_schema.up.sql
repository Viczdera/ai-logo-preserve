CREATE TABLE "jobs" (
  "id" bigserial UNIQUE PRIMARY KEY NOT NULL,
  "status" varchar NOT NULL,
  "s3_key" varchar UNIQUE NOT NULL,
  "upload_url" varchar NOT NULL,
  "result_url" varchar ,
  "logos_found" varchar,
  "error_message" varchar,
  "created_at" timestamptz NOT NULL DEFAULT 'now()',
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "completed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "logos" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "job_id" bigserial NOT NULL,
  "bounding_box" varchar NOT NULL,
  "confidence" int8 NOT NULL,
  "logo_type" varchar NOT NULL,
  "s3_key" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT 'now()'
);

ALTER TABLE "logos" ADD FOREIGN KEY ("job_id") REFERENCES "jobs" ("id");
ALTER TABLE "logos" ADD FOREIGN KEY ("s3_key") REFERENCES "jobs" ("s3_key");

-- Indexes for jobs table
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_updated_at ON jobs(updated_at);
CREATE INDEX idx_jobs_completed_at ON jobs(completed_at);

-- Indexes for logos table
CREATE INDEX idx_logos_job_id ON logos(job_id);
CREATE INDEX idx_logos_confidence ON logos(confidence);
CREATE INDEX idx_logos_logo_type ON logos(logo_type);
CREATE INDEX idx_logos_created_at ON logos(created_at);
CREATE INDEX idx_logos_job_id_confidence ON logos(job_id, confidence);
