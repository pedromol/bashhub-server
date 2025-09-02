#!/bin/bash

# Test runner script for bashhub-server
# This script provides comprehensive test execution with different modes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TEST_TYPE="all"
COVERAGE=false
VERBOSE=true
POSTGRES=false

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

# Function to show usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Run tests for bashhub-server"
    echo
    echo "Options:"
    echo "  -t, --type TYPE      Test type: unit, integration, all (default: all)"
    echo "  -c, --coverage       Generate coverage report"
    echo "  -v, --verbose        Verbose output (default: true)"
    echo "  -q, --quiet          Quiet output"
    echo "  -p, --postgres       Run PostgreSQL integration tests"
    echo "  -h, --help           Show this help message"
    echo
    echo "Examples:"
    echo "  $0                    # Run all tests"
    echo "  $0 -t unit           # Run only unit tests"
    echo "  $0 -c                # Run tests with coverage"
    echo "  $0 -p                # Run PostgreSQL integration tests"
    echo
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quiet)
            VERBOSE=false
            shift
            ;;
        -p|--postgres)
            POSTGRES=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -d "internal" ]]; then
    print_error "Please run this script from the project root directory"
    exit 1
fi

# Function to run unit tests
run_unit_tests() {
    print_header "Running unit tests..."

    local test_cmd="go test"
    if [[ "$VERBOSE" == "true" ]]; then
        test_cmd="$test_cmd -v"
    fi

    if [[ "$COVERAGE" == "true" ]]; then
        test_cmd="$test_cmd -coverprofile=coverage_unit.out"
    fi

    test_cmd="$test_cmd ./internal/..."

    if eval "$test_cmd"; then
        print_status "Unit tests passed!"
        return 0
    else
        print_error "Unit tests failed!"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_header "Running integration tests..."

    local test_cmd="go test -tags=integration"
    if [[ "$VERBOSE" == "true" ]]; then
        test_cmd="$test_cmd -v"
    fi

    if [[ "$COVERAGE" == "true" ]]; then
        test_cmd="$test_cmd -coverprofile=coverage_integration.out"
    fi

    test_cmd="$test_cmd ./..."

    if eval "$test_cmd"; then
        print_status "Integration tests passed!"
        return 0
    else
        print_error "Integration tests failed!"
        return 1
    fi
}

# Function to run PostgreSQL integration tests
run_postgres_tests() {
    print_header "Running PostgreSQL integration tests..."

    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Cannot run PostgreSQL tests."
        return 1
    fi

    # Start PostgreSQL container
    print_status "Starting PostgreSQL container..."
    docker run --name bashhub-test-postgres \
        -e POSTGRES_PASSWORD=password \
        -e POSTGRES_DB=bashhub_test \
        -p 5433:5432 \
        -d postgres:13-alpine

    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL to be ready..."
    sleep 10

    # Set test database URL
    export TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/bashhub_test?sslmode=disable"

    # Run tests
    local test_cmd="go test -tags=integration"
    if [[ "$VERBOSE" == "true" ]]; then
        test_cmd="$test_cmd -v"
    fi

    if [[ "$COVERAGE" == "true" ]]; then
        test_cmd="$test_cmd -coverprofile=coverage_postgres.out"
    fi

    test_cmd="$test_cmd ./internal/db/"

    local result=0
    if eval "$test_cmd"; then
        print_status "PostgreSQL integration tests passed!"
    else
        print_error "PostgreSQL integration tests failed!"
        result=1
    fi

    # Clean up
    print_status "Cleaning up PostgreSQL container..."
    docker stop bashhub-test-postgres >/dev/null 2>&1 || true
    docker rm bashhub-test-postgres >/dev/null 2>&1 || true

    return $result
}

# Function to generate coverage report
generate_coverage_report() {
    print_header "Generating coverage report..."

    if [[ -f "coverage_unit.out" ]] || [[ -f "coverage_integration.out" ]] || [[ -f "coverage_postgres.out" ]]; then
        # Merge coverage files if multiple exist
        if [[ $(ls coverage_*.out 2>/dev/null | wc -l) -gt 1 ]]; then
            print_status "Merging coverage files..."
            go tool cover -html=coverage_unit.out -o coverage_unit.html
            go tool cover -html=coverage_integration.out -o coverage_integration.html 2>/dev/null || true
            go tool cover -html=coverage_postgres.out -o coverage_postgres.html 2>/dev/null || true
        else
            # Single coverage file
            local coverage_file=$(ls coverage_*.out | head -1)
            go tool cover -html="$coverage_file" -o coverage.html
            go tool cover -func="$coverage_file"
        fi

        print_status "Coverage report generated!"
        print_status "Open coverage.html in your browser to view the report"
    else
        print_warning "No coverage files found"
    fi
}

# Main execution
main() {
    print_header "Starting test suite for bashhub-server"
    print_status "Test type: $TEST_TYPE"
    if [[ "$COVERAGE" == "true" ]]; then
        print_status "Coverage reporting enabled"
    fi

    local exit_code=0

    case $TEST_TYPE in
        "unit")
            if ! run_unit_tests; then
                exit_code=1
            fi
            ;;
        "integration")
            if ! run_integration_tests; then
                exit_code=1
            fi
            ;;
        "all")
            if ! run_unit_tests; then
                exit_code=1
            fi

            if [[ $exit_code -eq 0 ]]; then
                if ! run_integration_tests; then
                    exit_code=1
                fi
            fi
            ;;
        *)
            print_error "Invalid test type: $TEST_TYPE"
            usage
            exit 1
            ;;
    esac

    # Run PostgreSQL tests if requested
    if [[ "$POSTGRES" == "true" ]]; then
        if ! run_postgres_tests; then
            exit_code=1
        fi
    fi

    # Generate coverage report if requested
    if [[ "$COVERAGE" == "true" ]]; then
        generate_coverage_report
    fi

    if [[ $exit_code -eq 0 ]]; then
        print_status "All tests completed successfully!"
    else
        print_error "Some tests failed!"
    fi

    return $exit_code
}

# Run main function
main "$@"
