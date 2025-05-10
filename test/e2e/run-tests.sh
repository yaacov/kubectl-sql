#!/bin/bash

# E2E test script for kubectl-sql
# Runs examples from README against a CRC cluster and validates output

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

KUBECTL_SQL_BIN="$(pwd)/kubectl-sql"
TOTAL_TESTS=0
PASSED_TESTS=0

echo -e "${YELLOW}Starting kubectl-sql E2E tests${NC}"

# Check if CRC is running
if ! crc status | grep -q "Running"; then
    echo -e "${RED}Error: CRC is not running. Please start CRC before running tests.${NC}"
    exit 1
fi

# Verify admin login
if ! oc whoami | grep -q "admin"; then
    echo -e "${YELLOW}Warning: You may not be logged in as admin. Some tests might fail.${NC}"
    echo -e "${YELLOW}Current user: $(oc whoami)${NC}"
fi

# Ensure kubectl-sql binary exists
if [ ! -f "$KUBECTL_SQL_BIN" ]; then
    echo -e "${RED}kubectl-sql binary not found at $KUBECTL_SQL_BIN${NC}"
    exit 1
fi

echo -e "${YELLOW}Using kubectl-sql binary: $KUBECTL_SQL_BIN${NC}"

# Function to run a test
run_test() {
    local name="$1"
    local command="$2"
    local expected_pattern="$3"
    local should_fail="${4:-false}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}Running test: $name${NC}"
    echo "Command: $command"
    
    if [ "$should_fail" = "true" ]; then
        if eval "$command" 2>&1 | grep -q "$expected_pattern"; then
            echo -e "${GREEN}✓ Test passed (expected error found)${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ Test failed (expected error not found)${NC}"
        fi
    else
        if output=$(eval "$command"); then
            if echo "$output" | grep -q "$expected_pattern"; then
                echo -e "${GREEN}✓ Test passed${NC}"
                echo -e "${YELLOW}First 4 lines of output:${NC}"
                echo "$output" | head -n 4
                PASSED_TESTS=$((PASSED_TESTS + 1))
            else
                echo -e "${RED}✗ Test failed (expected pattern not found in output)${NC}"
                echo "Expected: $expected_pattern"
                echo "Got: $output"
            fi
        else
            echo -e "${RED}✗ Test failed (command execution error)${NC}"
        fi
    fi
    
    echo ""
}

# Create some test resources if needed
echo -e "${YELLOW}Creating test resources...${NC}"

# Check if namespace exists and delete it if it does
if oc get namespace kubectl-sql-test >/dev/null 2>&1; then
    echo -e "${YELLOW}Namespace kubectl-sql-test exists, deleting it...${NC}"
    oc delete namespace kubectl-sql-test
    echo -e "${YELLOW}Waiting for namespace deletion to complete...${NC}"
    while oc get namespace kubectl-sql-test >/dev/null 2>&1; do
        echo -e "${YELLOW}.${NC}"
        sleep 2
    done
    echo -e "${GREEN}Namespace deleted.${NC}"
fi

# Create new project (this will also create namespace and set it as current context)
echo -e "${YELLOW}Creating new project kubectl-sql-test...${NC}"
oc new-project kubectl-sql-test

oc create deployment hello-world --image=quay.io/redhatworkshops/welcome-app:latest -n kubectl-sql-test 2>/dev/null || true
oc scale deployment hello-world --replicas=3 -n kubectl-sql-test 2>/dev/null || true
oc expose deployment hello-world --port=8080 -n kubectl-sql-test 2>/dev/null || true

# Wait for deployment to be ready
echo -e "${YELLOW}Waiting for deployment to be ready...${NC}"
if oc wait --for=condition=Available --timeout=60s deployment/hello-world -n kubectl-sql-test; then
    echo -e "${GREEN}Deployment is ready.${NC}"
else
    echo -e "${RED}Timeout waiting for deployment to be ready.${NC}"
    exit 1
fi

# Run tests based on README examples
echo -e "${YELLOW}Running basic tests:${NC}"

run_test "Basic Pod Query" \
    "$KUBECTL_SQL_BIN 'select name, status.phase from pods where status.phase = \"Running\"'" \
    "Running"

run_test "Deployment Query" \
    "$KUBECTL_SQL_BIN 'select name, spec.replicas from deployments where spec.replicas > 2'" \
    "hello-world"

run_test "Namespace Filter" \
    "$KUBECTL_SQL_BIN -n kubectl-sql-test 'select name from pods'" \
    "hello-world"

run_test "Multiple Resources" \
    "$KUBECTL_SQL_BIN 'select kind, name from services where namespace = \"kubectl-sql-test\"'" \
    "Service"

# Basic Selection & Namespace Filtering Tests
echo -e "${YELLOW}Running Basic Selection & Namespace Filtering tests:${NC}"

run_test "Select All Pods in Namespace" \
    "$KUBECTL_SQL_BIN 'SELECT * FROM kubectl-sql-test/pods'" \
    "hello-world"

run_test "Service Names and Types" \
    "$KUBECTL_SQL_BIN 'SELECT name, spec.type FROM kubectl-sql-test/services'" \
    "ClusterIP"

# Sorting and Limiting Results Tests
echo -e "${YELLOW}Running Sorting and Limiting tests:${NC}"

run_test "Sort Pods by Name" \
    "$KUBECTL_SQL_BIN 'SELECT name, status.phase FROM */pods ORDER BY name LIMIT 5'" \
    "Running"

run_test "Multiple Column Sorting" \
    "$KUBECTL_SQL_BIN 'SELECT kind, name, namespace FROM */services ORDER BY namespace ASC, name DESC'" \
    "Service"

# WHERE Clause Filtering Tests
echo -e "${YELLOW}Running WHERE Clause Filtering tests:${NC}"

run_test "Filter by Pod Phase" \
    "$KUBECTL_SQL_BIN 'SELECT name, status.phase FROM */pods WHERE status.phase = \"Running\"'" \
    "Running"

run_test "Filter by Container Name" \
    "$KUBECTL_SQL_BIN 'SELECT name from */pods where spec.containers[0].name is not null'" \
    "hello-world"

# Aliasing Tests
echo -e "${YELLOW}Running Aliasing tests:${NC}"

run_test "Alias Field Names" \
    "$KUBECTL_SQL_BIN 'SELECT name, status.phase AS pod_phase FROM */pods'" \
    "pod_phase"

# RegExp Filtering Tests
echo -e "${YELLOW}Running RegExp tests:${NC}"
run_test "Filter Pods by Name Pattern" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */pods WHERE name ~= \"hello-.*\"'" \
    "hello-world"

# Array Operations Tests
echo -e "${YELLOW}Running Array Operations tests:${NC}"

run_test "Test Array Length Function" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */pods WHERE len(spec.containers) > 0'" \
    "hello-world"

# Add new Array tests
echo -e "${YELLOW}Running Additional Array Tests:${NC}"

run_test "Array Element Access" \
    "$KUBECTL_SQL_BIN 'SELECT name, spec.containers[0].name FROM */pods WHERE spec.containers[0].name is not null'" \
    "hello-world"

run_test "Array Wildcard Access" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */pods WHERE any (spec.containers[*].name is not null)'" \
    "hello-world"

run_test "Array Named Key Access" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */deployments WHERE spec.selector.matchLabels[app] = \"hello-world\"'" \
    "hello-world"

# Function Tests
echo -e "${YELLOW}Running Function Tests:${NC}"

run_test "Length Function" \
    "$KUBECTL_SQL_BIN 'SELECT name, len(spec.containers) FROM */pods'" \
    "hello-world"

run_test "Any Function" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */pods WHERE any(spec.containers[*].image ~= \"welcome-app\")'" \
    "hello-world"

run_test "All Function" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */pods WHERE all(spec.containers[*].imagePullPolicy = \"Always\")'" \
    "hello-world"

run_test "Sum Function" \
    "$KUBECTL_SQL_BIN 'SELECT name, sum(spec.containers[*].resources.limits.memory) as memory FROM */pods WHERE sum(spec.containers[*].resources.limits.memory) != 0'" \
    "memory"

# Alias Tests
echo -e "${YELLOW}Running Alias Tests:${NC}"

run_test "Basic Field Alias" \
    "$KUBECTL_SQL_BIN 'SELECT name, status.phase AS pod_status FROM */pods'" \
    "pod_status"

run_test "Multiple Aliases" \
    "$KUBECTL_SQL_BIN 'SELECT name AS pod_name, namespace AS ns, status.phase AS status FROM */pods'" \
    "pod_name"

run_test "Alias with Function" \
    "$KUBECTL_SQL_BIN 'SELECT name, len(spec.containers) AS container_count FROM */pods'" \
    "container_count"

# Complex Alias with Filter Tests
echo -e "${YELLOW}Running Complex Alias with Filter Tests:${NC}"

run_test "Filter on Aliased Field" \
    "$KUBECTL_SQL_BIN 'SELECT name, status.phase AS phase FROM */pods WHERE phase = \"Running\"'" \
    "Running"

run_test "Multiple Filters with Aliases" \
    "$KUBECTL_SQL_BIN 'SELECT name, namespace AS ns, status.phase AS phase FROM */pods WHERE ns = \"kubectl-sql-test\" AND phase = \"Running\"'" \
    "kubectl-sql-test"

run_test "Order By Aliased Field" \
    "$KUBECTL_SQL_BIN 'SELECT name, status.phase AS phase FROM */pods ORDER BY phase'" \
    "Running"

run_test "Complex Function with Alias in Filter" \
    "$KUBECTL_SQL_BIN 'SELECT name, len(spec.containers) AS containers FROM */pods WHERE containers > 0'" \
    "hello-world"

# Field Comparison Tests
echo -e "${YELLOW}Running Field Comparison tests:${NC}"

run_test "Compare Fields in Resources" \
    "$KUBECTL_SQL_BIN 'SELECT name FROM */deployments WHERE status.availableReplicas <= spec.replicas'" \
    "hello-world"

# Error Tests
echo -e "${YELLOW}Running Error tests:${NC}"

run_test "Error Test - Invalid SQL" \
    "$KUBECTL_SQL_BIN 'select invalid syntax from pods'" \
    "Error" \
    "true"

run_test "Error Test - Non-existent Resource" \
    "$KUBECTL_SQL_BIN 'select name from nonexistentresource'" \
    "Error" \
    "true"

# Generate summary
echo -e "${YELLOW}Test Summary:${NC}"
echo -e "${YELLOW}Total Tests: $TOTAL_TESTS${NC}"
echo -e "${GREEN}Passed Tests: $PASSED_TESTS${NC}"
if [ "$TOTAL_TESTS" -ne "$PASSED_TESTS" ]; then
    echo -e "${RED}Failed Tests: $(($TOTAL_TESTS - $PASSED_TESTS))${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
