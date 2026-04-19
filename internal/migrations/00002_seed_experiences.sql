-- +goose Up
INSERT INTO experiences (company, role, start_date, end_date, description, tech_stack, security_highlights, metadata, is_published, sort_order)
VALUES
(
    'Virtru',
    'Senior Full Stack Engineer',
    '2024-08-01',
    NULL,
    '[
        "Architected Delivery Decisioning — a webhook-driven engine enabling enterprise customers to run custom DLP scanners against E2E-encrypted data before delivery",
        "Built gRPC services on top of the Data Security Platform for automated key rotation and policy bootstrapping across multi-tenant environments",
        "Shipped hybrid deployment support — Helm charts for Kubernetes and standalone Linux provisioning scripts — to serve both cloud and on-prem customers",
        "Integrated the modern DSP into legacy on-prem Customer Key Servers using Caddy as a secure reverse proxy, enabling incremental migration without downtime",
        "Adopted a multi-agent AI workflow for development and security review, cutting feature delivery time while catching insecure patterns before code review"
    ]'::jsonb,
    '["Go", "gRPC", "Kubernetes", "Helm", "Caddy", "Docker", "Terraform"]'::jsonb,
    '["E2EE delivery decisioning with custom DLP scanning", "Automated security auditing via AI agents", "On-prem key server hardening"]'::jsonb,
    '{"location": "Remote"}'::jsonb,
    true,
    1
),
(
    'Adapter',
    'DevOps & Full Stack Engineer',
    '2024-04-01',
    '2024-08-01',
    '[
        "Stood up production observability from scratch — Prometheus metrics, Grafana dashboards, and PagerDuty alerting across GCP-hosted services",
        "Built CloudBuild CI/CD pipelines with automated container scanning and image signing, eliminating manual release steps",
        "Designed and deployed client-to-site VPN infrastructure to enforce zero-trust access for developers working with private cloud resources"
    ]'::jsonb,
    '["Go", "GCP", "Prometheus", "PagerDuty", "CloudBuild", "Docker", "Terraform"]'::jsonb,
    '["VPN infrastructure for secure developer access", "Container image scanning and signing", "Private cloud resource segmentation"]'::jsonb,
    '{"location": "Remote"}'::jsonb,
    true,
    2
),
(
    'Virtru',
    'Senior Full Stack Engineer',
    '2020-11-01',
    '2024-04-01',
    '[
        "Championed and introduced Go as an officially supported language, architecting the company''s first Go services and establishing patterns adopted across teams",
        "Built a high-throughput RESTful API in Go serving 600K+ requests/day for the core encryption platform",
        "Led a core product as both engineer and project manager for 2 years — owned roadmap, architecture decisions, and shipped on schedule",
        "Redesigned ETL pipelines from 7 integration points down to 3, processing 2M+ records/day with improved reliability"
    ]'::jsonb,
    '["Go", "Node.js", "PostgreSQL", "Redis", "Kafka", "AWS", "Docker", "Kubernetes"]'::jsonb,
    '["Data encryption at rest and in transit", "SOC 2 compliance automation", "Key management infrastructure"]'::jsonb,
    '{"location": "Remote"}'::jsonb,
    true,
    3
),
(
    'Bell Media, LLC',
    'Web Developer',
    '2018-06-01',
    '2020-11-01',
    '[
        "Built and shipped the company''s first mobile app from scratch using Flutter — sole developer from design through App Store release",
        "Automated product and inventory management by integrating Python scripts with WooCommerce REST APIs, replacing manual workflows",
        "Designed authentication flows and resolved critical bugs across a legacy PHP/jQuery codebase, stabilizing the platform for continued use"
    ]'::jsonb,
    '["Flutter", "Dart", "Python", "PHP", "jQuery", "WooCommerce", "REST APIs"]'::jsonb,
    '["Authentication workflow design", "Secure API integrations"]'::jsonb,
    '{"location": "Remote"}'::jsonb,
    true,
    4
);

-- +goose Down
DELETE FROM experiences WHERE company IN ('Virtru', 'Adapter', 'Bell Media, LLC');
