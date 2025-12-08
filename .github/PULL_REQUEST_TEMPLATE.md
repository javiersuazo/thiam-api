## Summary
Brief description of the changes.

## Related Issue
Fixes #(issue number)

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)

## Checklist

### General
- [ ] I have read the [CONTRIBUTING](CONTRIBUTING.md) guidelines
- [ ] I have performed a self-review of my code
- [ ] New and existing tests pass locally (`make test`)
- [ ] I have run the linter (`make linter-golangci`)

### Code Quality ([Coding Rules](docs/CODING_RULES.md))
- [ ] Errors are wrapped with context using `fmt.Errorf("StructName.Method: %w", err)`
- [ ] Domain errors use sentinel errors (`errors.Is()` for checking)
- [ ] Interfaces are small and defined where they're used
- [ ] Context is passed as first parameter to all functions
- [ ] No sensitive data in logs or error messages
- [ ] Tests are table-driven where applicable

### Architecture ([Architecture](docs/ARCHITECTURE.md))
- [ ] Changes respect layer boundaries (dependencies point inward)
- [ ] Business logic is in use cases, not controllers
- [ ] New features follow existing patterns

## Test Plan
Describe how you tested these changes.

## Screenshots (if applicable)
Add screenshots to help explain your changes.
