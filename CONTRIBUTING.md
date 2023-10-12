# Contributing to kubectl-sql

Thank you for considering contributing to kubectl-sql! We welcome contributions from everyone. Here are some guidelines to help you get started.

## Getting Started

### Setting up the Development Environment

Participating in the development of our project involves forking the [repository](https://github.com/yaacov/kubectl-sql), setting up your local development environment, making changes, and then proposing those changes via a pull request. Below, we walk through the general steps to get you started!

#### 1. Forking the Repository and Setting Up Local Development

**Fork and Clone:** Begin by forking the repository and then clone your fork locally. For step-by-step instructions, check GitHub's guide on [forking repositories](https://docs.github.com/en/get-started/quickstart/fork-a-repo) and [cloning repositories](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository).

```bash
Copy code
git clone https://github.com/[YourUsername]/kubectl-sql.git
cd kubectl-sql
```

  Remember to replace `[YourUsername]` with your GitHub username.

#### 2. Installing Dependencies

**Install Go:** Ensure Go is installed on your machine. If not, download it from the [official Go site](https://golang.org/dl/) and refer to the [installation guide](https://golang.org/doc/install).

**Manage Project Dependencies:** Navigate to the project directory and manage the dependencies using Go Modules:

```bash
go mod tidy
go mod download
```

#### 3. Building and Running the Project
Build the project using Go, or make if a Makefile is available, and verify that it runs locally.

```bash
make
```

Now you should be able to execute the binary or utilize the project as per its functionality and options.

#### 4. Making Changes and Contributing

Once your environment is set up and running, you’re ready to code!

When you're ready to contribute your changes back to the project:

  - Ensure to adhere to the project’s coding standards and guidelines.
  - Refer to the GitHub guide for creating a pull request from your fork.

Congratulations, you’re set up for contributing to the project! Always check any additional CONTRIBUTING guidelines provided by the project repository and engage respectfully with the existing community. Happy coding!

## Understanding the Project Structure

Navigating through a project can be quite daunting if you are unfamiliar with its architecture. Here's a brief overview of our Go project structure to get you started:

`cmd/`
The `cmd/` directory contains the application's entry points, essentially harboring the command-line interfaces or executables of the project. Each subdirectory within `cmd/` is dedicated to an actionable command that the application can perform.

`cmd/kubectl-sql`: This subdirectory holds the source code of the specific command-line user interface. The main function within this directory acts as the entry point to the command.

`pkg/`
The `pkg/` directory encompasses helper modules and libraries that are utilized by the main application and can potentially be shared with other projects. The `pkg/` directory is meant to provide a clear distinction between the application code and the auxiliary code that supports it.

It's crucial to recognize that the code within pkg/ should be designed with reusability in mind, avoiding dependencies from your cmd/ directory, ensuring clean and modular code.

## How to Contribute

### Reporting Bugs
Ensure the bug was not already reported by searching on GitHub under Issues.
If you're unable to find an open issue addressing the problem, open a new one.

### Suggesting Enhancements
Open a new issue with a detailed explanation of your suggestion.

## Your First Code Contribution
Begin by looking for good first issues tags in the Issues.
Do not work on an issue without expressing interest by commenting on the issue.

## Pull Requests
  - Fork the Repo: Fork the project repository and clone your fork.
  - Create a Branch: Make a new branch for your feature or bugfix.
  - Commit Your Changes: Make sure your code meets the go style guidelines and add tests for new features.
  - Push to Your Fork: And submit a pull request to the main branch.
