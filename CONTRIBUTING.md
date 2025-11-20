# Contributing Guidelines

## Architecture

Refer to the [architecture document](/ARCHITECTURE.md)

## GitHub Issues

GitHub issues should be used for bug reports only. Feature requests and refactor requests should be contributed by adding a new file to the `specs/` directory.

## Specs

Specs are stored in the `specs/` directory.

Each spec should be a markdown file with a numeric prefix - for example, `000-mvp.md`.

Each spec file should contain the following:

- Title
- Status
  - pending, in-progress, or complete
- Description
  - A few sentences describing why this change will be useful.
- Design decisions
  - An optional list of design decisions that were made, with pros and cons for the different options considered.
- Task List
  - A checklist of tasks for the implementing the change.

## Commit Messages and PR Titles

### Commit Messages

Use conventional commit format:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

**Examples:**

```
feat: add astro project with react islands
fix: resolve build error in home page
docs: update api documentation
```

### Pull Request Titles

Keep PR titles **short and descriptive**, typically 3-7 words:

**Good examples:**

- `feat: basic infrastructure for home page`
- `fix: tailwind css styling issues`
- `docs: update contributing guidelines`

**Bad examples (avoid):**

- `feat(0): infrastructure setup for basic home page with astro 5.x and react 19 islands including typescript strict mode and tailwind css 4.x integration and environment variable configuration`
- `update the home page` (too vague, do not include number in scope)
