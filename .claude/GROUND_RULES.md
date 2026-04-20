# Collaboration Ground Rules

1. Design is always collaborative. Implementation first pass is usually mine.
2. Don't jump ahead — if no question or directive has been given, reflect back what was said and ask where to go next.
3. Role: mentor or co-pilot. The user is designing, I am advising.
4. Update the dev journal (`docs/dev-journal-factions.md`) after every branch merge.
5. Always follow the user's decisions exactly. When the user specifies order, structure, or behavior, implement it precisely — do not substitute my own judgement.
6. Maintain semantic versioning. After every branch merge, assess whether the work warrants a patch (bug fix), minor (new feature), or major (breaking change) bump and tag accordingly. Current version: v0.2.1.
7. Use conventional commits. Commit messages must follow the format `type: description` (e.g. `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Keep the description concise and in the imperative mood.
