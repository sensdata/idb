# Contributing to iDB

First off, thank you for considering contributing to iDB! It's people like you that make iDB such a great tool.

## Table of Contents

- [Contributing to iDB](#contributing-to-idb)
  - [Table of Contents](#table-of-contents)
  - [Code of Conduct](#code-of-conduct)
  - [How Can I Contribute?](#how-can-i-contribute)
    - [Reporting Bugs](#reporting-bugs)
    - [Suggesting Enhancements](#suggesting-enhancements)
    - [Your First Code Contribution](#your-first-code-contribution)
    - [Pull Requests](#pull-requests)
  - [Development Setup](#development-setup)
    - [Prerequisites](#prerequisites)
    - [Backend Development](#backend-development)
    - [Frontend Development](#frontend-development)
  - [Coding Guidelines](#coding-guidelines)
    - [Go Code Guidelines](#go-code-guidelines)
    - [JavaScript/TypeScript Guidelines](#javascripttypescript-guidelines)
  - [Testing](#testing)
  - [Documentation](#documentation)
  - [Submitting a Pull Request](#submitting-a-pull-request)
  - [Styleguides](#styleguides)
    - [Git Commit Messages](#git-commit-messages)
    - [Go Code Style](#go-code-style)
    - [JavaScript/TypeScript Code Style](#javascripttypescript-code-style)
    - [Documentation Style](#documentation-style)

## Code of Conduct

This project and everyone participating in it is governed by the [iDB Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [support@sensdata.com].

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for iDB. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

#### Before Submitting A Bug Report

- **Check the [Issues](https://github.com/sensdata/idb/issues)** to see if the bug has already been reported. If it has and the issue is still open, add a comment to the existing issue instead of opening a new one.
- **Check if the issue has been fixed** by trying to reproduce it using the latest version of the repository.
- **Perform a [cursory search](https://github.com/search?q=+is%3Aissue+repo%3Asensdata%2Fidb)** to see if the problem has been reported.

#### How Do I Submit A Good Bug Report?

Bugs are tracked as [GitHub issues](https://guides.github.com/features/issues/). Create an issue on that repository and provide the following information by filling in the [template](ISSUE_TEMPLATE/bug_report.md).

Explain the problem and include additional details to help maintainers reproduce the problem:

- **Use a clear and descriptive title** for the issue to identify the problem.
- **Describe the exact steps which reproduce the problem** in as many details as possible.
- **Provide specific examples to demonstrate the steps**. Include links to files or GitHub projects, or copy/pasteable snippets, which you use in those examples.
- **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
- **Explain which behavior you expected to see instead and why.**
- **Include screenshots and animated GIFs** which show you following the described steps and clearly demonstrate the problem.
- **If you're reporting that iDB crashed**, include a crash report with a stack trace from the operating system.
- **If the problem is related to performance or memory**, include a CPU profile capture with your report.
- **If the problem wasn't triggered by a specific action**, describe what you were doing before the problem happened and share more information using the guidelines below.

Provide more context by answering these questions:

- **Did the problem start happening recently** (e.g. after updating to a new version of iDB) or was this always a problem?
- If the problem started happening recently, **can you reproduce the problem in an older version of iDB?** What's the most recent version in which the problem doesn't happen?
- **Can you reliably reproduce the issue?** If not, provide details about how often the problem happens and under which conditions it normally happens.

Include details about your configuration and environment:

- **Which version of iDB are you using?**
- **What's the name and version of the OS you're using?**
- **Are you running iDB in a virtual machine?** If so, which VM software are you using and which operating systems and versions are used for the host and the guest?
- **Which Docker version are you using?**

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for iDB, including completely new features and minor improvements to existing functionality. Following these guidelines helps maintainers and the community understand your suggestion and find related suggestions.

#### Before Submitting An Enhancement Suggestion

- **Check the [Issues](https://github.com/sensdata/idb/issues)** to see if the enhancement has already been suggested. If it has and the issue is still open, add a comment to the existing issue instead of opening a new one.
- **Check if the enhancement has been implemented** by trying the latest version of the repository.
- **Perform a [cursory search](https://github.com/search?q=+is%3Aissue+repo%3Asensdata%2Fidb)** to see if the enhancement has been suggested before.

#### How Do I Submit A Good Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://guides.github.com/features/issues/). Create an issue on that repository and provide the following information by filling in the [template](ISSUE_TEMPLATE/feature_request.md).

- **Use a clear and descriptive title** for the issue to identify the suggestion.
- **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
- **Provide specific examples to demonstrate the steps**. Include copy/pasteable snippets which you use in those examples, as Markdown code blocks.
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
- **Include screenshots and animated GIFs** which help you demonstrate the steps or point out the part of iDB which the suggestion is related to.
- **Explain why this enhancement would be useful** to most iDB users.
- **List some other tools or applications where this enhancement exists.**
- **Specify which version of iDB you're using.**
- **Specify the name and version of the OS you're using.**

### Your First Code Contribution

Unsure where to begin contributing to iDB? You can start by looking through these `good first issue` and `help wanted` issues:

- **[good first issue](https://github.com/sensdata/idb/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)** - issues which should only require a few lines of code, and a test or two.
- **[help wanted](https://github.com/sensdata/idb/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)** - issues which should be a bit more involved than `good first issue` issues.

#### Local Development

1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/sensdata/idb.git
   cd idb
   ```
3. Create a branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. Make your changes following the coding guidelines.
5. Test your changes.
6. Commit your changes following the commit message guidelines.
7. Push your changes to GitHub:
   ```bash
   git push origin feature/your-feature-name
   ```
8. Create a pull request from your fork to the main repository.

### Pull Requests

The process described here has several goals:

- Maintain iDB's quality
- Fix problems that are important to users
- Engage the community in working toward the best possible iDB
- Enable a sustainable system for iDB's maintainers to review contributions

Please follow these steps to have your contribution considered by the maintainers:

1. Follow all instructions in [the template](PULL_REQUEST_TEMPLATE.md)
2. Follow the [styleguides](#styleguides)
3. After you submit your pull request, verify that all [status checks](https://help.github.com/articles/about-status-checks/) are passing

While the prerequisites above must be satisfied prior to having your pull request reviewed, the reviewer(s) may ask you to complete additional design work, tests, or other changes before your pull request can be ultimately accepted.

## Development Setup

### Prerequisites

- **Go** (version 1.23+)
- **Node.js** (version 18+)
- **npm** or **yarn**
- **Docker** (for container-related features)

### Backend Development

1. Navigate to the backend directory:
   ```bash
   cd center
   ```

2. Build the backend:
   ```bash
   make build
   ```

3. Run the backend:
   ```bash
   ./idb start
   ```

### Frontend Development

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Run the development server:
   ```bash
   npm run dev
   ```

4. Build for production:
   ```bash
   npm run build
   ```

## Coding Guidelines

### Go Code Guidelines

1. **Follow Go conventions** - Use `go fmt` to format your code.
2. **Use meaningful variable names** - Avoid single-letter variables except for loop counters.
3. **Write clear comments** - Explain why the code does something, not just what it does.
4. **Use error handling** - Don't ignore errors.
5. **Write small functions** - Each function should do one thing well.
6. **Use interfaces** - For better testability and flexibility.
7. **Avoid global variables** - Where possible, pass dependencies as parameters.

### JavaScript/TypeScript Guidelines

1. **Follow TypeScript best practices** - Use strict type checking.
2. **Use meaningful variable names** - Avoid abbreviations where possible.
3. **Write clear comments** - Explain complex logic.
4. **Use async/await** - Avoid callback hell.
5. **Follow Vue.js best practices** - Use composition API where possible.
6. **Keep components small** - Each component should have a single responsibility.

## Testing

1. **Write unit tests** for all new functions.
2. **Write integration tests** for complex features.
3. **Run the test suite** before submitting a pull request:
   ```bash
   cd center && go test ./...
   cd ../agent && go test ./...
   cd ../core && go test ./...
   ```
4. **Ensure 100% test coverage** for critical components.

## Documentation

1. **Update documentation** for all changes.
2. **Write clear API documentation** using GoDoc comments.
3. **Update README.md** if you change the installation or usage instructions.
4. **Update the changelog** for all user-visible changes.

## Branch Policy

- **All pull requests must target the `develop` branch**
- **`main` is reserved for releases** and is updated only by maintainers
- **PRs opened against `main` will be closed**

## Submitting a Pull Request

1. **Fork the repository** and create your branch from `develop`.
2. **Test your changes** - Ensure all tests pass.
3. **Update documentation** - Ensure your changes are properly documented.
4. **Follow the commit message guidelines**.
5. **Submit your pull request** to the `develop` branch.
6. **Respond to feedback** - Be prepared to make changes based on reviewer feedback.

## Styleguides

### Git Commit Messages

1. **Use the present tense** ("Add feature" not "Added feature").
2. **Use the imperative mood** ("Move cursor to..." not "Moves cursor to...").
3. **Limit the first line to 72 characters or less**.
4. **Reference issues and pull requests** liberally after the first line.
5. **Add a description** - Explain what the commit does and why it's needed.
6. **Use prefixes** to categorize your commits:
   - `feat`: A new feature
   - `fix`: A bug fix
   - `docs`: Documentation only changes
   - `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
   - `refactor`: A code change that neither fixes a bug nor adds a feature
   - `perf`: A code change that improves performance
   - `test`: Adding missing tests or correcting existing tests
   - `chore`: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Go Code Style

1. **Use `go fmt`** to format your code.
2. **Use `golangci-lint`** to check for linting errors.
3. **Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)**.
4. **Use meaningful package names** - Package names should be short and descriptive.
5. **Use dependency injection** for better testability.

### JavaScript/TypeScript Code Style

1. **Use `prettier`** to format your code.
2. **Use `eslint`** to check for linting errors.
3. **Follow the [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)**.
4. **Use TypeScript strict mode**.

### Documentation Style

1. **Use Markdown** for all documentation.
2. **Use clear headings** to organize content.
3. **Use code blocks** for code examples.
4. **Use screenshots** to illustrate complex concepts.
5. **Keep documentation up to date**.

Thank you for contributing to iDB! 🎉
