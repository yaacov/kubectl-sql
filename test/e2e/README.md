# End-to-End Tests for kubectl-sql

This directory contains end-to-end tests for the kubectl-sql tool that validate functionality against a real Kubernetes cluster.

## Prerequisites

- Red Hat CodeReady Containers (CRC) running
- Logged in as admin user to the CRC cluster
- kubectl-sql compiled successfully

## Running the Tests

From the project root, run:

```bash
make e2e-test
```

The test script will:

1. Verify CRC is running
2. Use the locally built kubectl-sql binary
3. Create test resources in a namespace called "kubectl-sql-test"
4. Run a series of example commands based on the project README
5. Validate the output against expected patterns

## Test Results

The script will output a summary of passed and failed tests. If any test fails, the script will exit with a non-zero exit code.

## Cleaning Up

The test creates resources in the "kubectl-sql-test" namespace. To clean up after testing:

```bash
oc delete namespace kubectl-sql-test
```
