CREATE TABLE IF NOT EXISTS templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT NOT NULL,
  image TEXT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO templates(slug, image, description) VALUES 
('base', 'alpine:latest', 'Minimal Alpine Linux'),
('python-3.12', 'python:3.12-slim', 'Pythong 3.12 with pip'),
('node-20', 'node:20-slim', 'Node.js 20 with npm');