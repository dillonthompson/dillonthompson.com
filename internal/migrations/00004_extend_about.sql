-- +goose Up
-- Extend the profile.about JSONB row with long-form bio fields used by the
-- /about page. Renders gracefully if any individual field is empty/missing.
--
-- Fields:
--   bio        — array of paragraphs for the professional narrative
--   personal   — single paragraph for the "outside of code" section
--   interests  — array of short strings (hobbies, shows, games) shown as chips
UPDATE profile
SET content = content || '{
    "bio": [
        "I''m a senior full stack engineer with over 7 years of experience building secure, high-scale systems. Most of that time has been spent in Go and TypeScript, with a focus on the boring-but-critical layer where APIs, databases, and infrastructure meet.",
        "These days I''m at Virtru, where I architect Delivery Decisioning — a webhook-driven engine that lets enterprise customers run custom DLP scanners against E2E-encrypted data before delivery. Before that I introduced Go as an officially supported language at Virtru and led a core product as both engineer and PM for two years.",
        "I take products from zero to one. I work best on small teams where I can own architecture end-to-end, ship fast, and build the kind of infrastructure stories that hold up under interview-level scrutiny — like the one this site runs on (you''re welcome to poke around the deploy logs)."
    ],
    "personal": "Outside of code I''m a husband, a gamer, and a pop-culture sponge. Birmingham, AL is home. I lean into projects that have a personality — which is why this portfolio has a deploy-themed intro animation, easter-egg quotes, and a Cmd+K terminal instead of just being yet another React resume page.",
    "interests": [
        "Halo",
        "Valheim",
        "Adventure Time",
        "Rick and Morty",
        "Community",
        "Midnight Gospel",
        "Cooking",
        "AI-augmented dev workflows"
    ]
}'::jsonb
WHERE key = 'about';

-- +goose Down
UPDATE profile
SET content = content - 'bio' - 'personal' - 'interests'
WHERE key = 'about';
