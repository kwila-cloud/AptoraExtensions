# Agent Guidelines

Be sure to refer to CONTRIBUTING.md for detailed guidelines on contributing to this project.

## Project Context

This is a small web server with extra features based on data from Aptora MS SQL Server database.

### Contribution Process

- **Specs**: Feature requests → `specs/` directory (see CONTRIBUTING.md)
- **Issues**: Bug reports only in GitHub
- **Code style**: See CONTRIBUTING.md for full guidelines

## Documentation Synchronization

**CRITICAL**: When making architectural decisions or updating documentation, you MUST keep related documents in sync.

### Rules
1. **Architecture changes** → Update BOTH `ARCHITECTURE.md` AND relevant **pending/active** spec files in `specs/`
   - Only update specs that are currently being worked on or planned
   - Do NOT update specs that are completed/implemented (they are historical records)
2. **Spec changes that affect architecture** → Update BOTH the spec AND `ARCHITECTURE.md`
3. **Always verify** you've updated all affected documents before moving to next topic

### Examples

<example>
<situation>User decides to use Go's embed package for serving frontend</situation>
<correct>
1. Update ARCHITECTURE.md with the decision and rationale
2. Update specs/000-proof-of-concept.md to reflect:
   - Build process changes (embed into binary)
   - Deployment changes (single binary instead of tarball)
   - Frontend serving implementation details
   - Development mode requirements (hot-reload flag)
</correct>
<incorrect>
1. Update ARCHITECTURE.md only
2. Move to next topic without updating POC spec
</incorrect>
</example>

<example>
<situation>User adds new high-level principle to ARCHITECTURE.md</situation>
<correct>
1. Update ARCHITECTURE.md with new principle
2. Check if any specs need updates to align with new principle
3. Ask user if this principle affects POC or other specs
</correct>
<incorrect>
1. Update ARCHITECTURE.md only
2. Assume specs don't need changes
</incorrect>
</example>

## Pull Request Workflow

**CRITICAL**: Spec Task List completion is a multi-step process:

### /start-pr Phase (Planning and Research)
- **ONLY** do planning, research, and asking clarifying questions
- **DO NOT** make any code changes or implementation
- **DO NOT** mark spec items as complete or check things off
- **MAY** update spec files for adding/revising the spec content (not completion status)
- **DO** explain the planned approach and wait for user approval

### /do-pr Phase (Implementation)
- User runs `/do-pr` command to trigger implementation
- Begin actual implementation work
- Mark spec items as completed only when work is actually finished
- Make code changes, create files, update documentation as needed

### Common Mistakes to Avoid

<example>
<situation>Agent does implementation work during /start-pr</situation>
<incorrect>
1. Research Task List section
2. Ask clarifying questions
3. Make code changes (create files, edit files)
4. Mark spec items as completed ✅
5. Wait for approval
</incorrect>
<correct>
1. Research Task List section
2. Ask clarifying questions  
3. Explain planned approach and wait for approval
4. Only make code changes after `/do-pr` is triggered
5. Only mark spec items as completed after actual implementation
</correct>
</example>

<example>
<situation>Agent updates spec file completion status during /start-pr</situation>
<incorrect>
1. Read spec file
2. Ask clarifying questions and get answers
3. Update spec file with completed checkboxes
4. Explain what was "completed"
</incorrect>
<correct>
1. Read spec file
2. Ask clarifying questions and get answers
3. Explain planned approach and wait for approval
4. Only update spec file completion status after `/do-pr` when actual work is implemented
</correct>
</example>

### Rule Summary
**NEVER make code changes or mark spec items as complete during /start-pr. Only do planning and research.**

<example>
<situation>User adds new high-level principle to ARCHITECTURE.md</situation>
<correct>
1. Update ARCHITECTURE.md with new principle
2. Check if any specs need updates to align with new principle
3. Ask user if this principle affects POC or other specs
</correct>
<incorrect>
1. Update ARCHITECTURE.md only
2. Assume specs don't need changes
</incorrect>
</example>

## Where to Find More Info

NOTE: we need to add these files

- **Architecture details**: [ARCHITECTURE.md](./ARCHITECTURE.md)
- **Contributing guidelines**: [CONTRIBUTING.md](./CONTRIBUTING.md)
- **Project overview**: [README.md](./README.md)
- **Current specs**: [`specs/`](./specs/) directory

