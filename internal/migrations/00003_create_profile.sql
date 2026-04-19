-- +goose Up
CREATE TABLE IF NOT EXISTS profile (
    key     TEXT PRIMARY KEY,
    content JSONB NOT NULL DEFAULT '{}'::jsonb
);

INSERT INTO profile (key, content) VALUES
(
    'about',
    '{
        "name": "Dillon Thompson",
        "title": "Senior Full Stack Engineer & Technical Consultant",
        "location": "Birmingham, AL",
        "phone": "(205) 305-5943",
        "email": "dj.thompson715@gmail.com",
        "linkedin": "linkedin.com/in/dillonthompson5",
        "github": "github.com/dillonthompson",
        "summary": "Senior full stack engineer with over 7 years of experience specializing in Go, Node.js, and React. Expert in architecting secure, high-scale systems and taking products from 0 to 1. I utilize a modern, multi-agent AI orchestration workflow to accelerate development cycles and perform automated security audits, ensuring production-ready code at high velocity."
    }'::jsonb
),
(
    'skills',
    '{
        "languages": ["Go", "JavaScript", "TypeScript", "Python", "C++", "Java", "Dart"],
        "frameworks": ["Node", "React", "Next.js", "Flutter", "Gin/Echo (Go)", "Selenium", "Cypress"],
        "tools_devops": ["Docker", "Kubernetes", "Terraform", "AWS", "GCP", "GitHub Actions", "Helm", "gRPC", "Caddy"]
    }'::jsonb
),
(
    'education',
    '{
        "degree": "Bachelor of Software Engineering",
        "school": "Auburn University",
        "graduated": "May 2020",
        "gpa": "3.71",
        "honors": ["Dean''s List"],
        "certifications": [
            {"name": "Cloud-Native Development with OpenShift and Kubernetes", "issuer": "Red Hat", "year": "2024"},
            {"name": "Google Cloud Infrastructure", "issuer": "Google Cloud", "year": "2024"}
        ]
    }'::jsonb
);

-- +goose Down
DROP TABLE IF EXISTS profile;
