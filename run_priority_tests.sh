#!/bin/bash

# Priority Queue Test Runner
# This script runs all priority queue tests with different configurations

echo "🚀 Running ORCHWF Priority Queue Tests"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to run tests and report results
run_test_suite() {
    local test_name="$1"
    local test_pattern="$2"
    local description="$3"
    
    echo -e "\n${YELLOW}📋 $test_name${NC}"
    echo "Description: $description"
    echo "Pattern: $test_pattern"
    echo "----------------------------------------"
    
    if go test -v -run "$test_pattern" ./...; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        return 1
    fi
}

# Function to run tests with coverage
run_test_with_coverage() {
    local test_name="$1"
    local test_pattern="$2"
    local coverage_file="$3"
    
    echo -e "\n${YELLOW}📊 $test_name (with coverage)${NC}"
    echo "Pattern: $test_pattern"
    echo "Coverage file: $coverage_file"
    echo "----------------------------------------"
    
    if go test -v -cover -coverprofile="$coverage_file" -run "$test_pattern" ./...; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        echo "Coverage report saved to: $coverage_file"
        return 0
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    echo -e "\n${YELLOW}⚡ Running Priority Queue Benchmarks${NC}"
    echo "----------------------------------------"
    
    if go test -bench=BenchmarkPriority -benchmem ./...; then
        echo -e "${GREEN}✅ Benchmarks completed${NC}"
        return 0
    else
        echo -e "${RED}❌ Benchmarks failed${NC}"
        return 1
    fi
}

# Initialize counters
total_tests=0
passed_tests=0
failed_tests=0

# Test Suite 1: Basic Functionality
echo -e "\n${YELLOW}🧪 Test Suite 1: Basic Functionality${NC}"
echo "============================================="

run_test_suite "Basic Priority Tests" "TestPriorityQueueBasic" "Tests basic priority-based execution order"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Priority with Dependencies" "TestPriorityWithDependencies" "Tests priority with step dependencies"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Priority with Async Steps" "TestPriorityWithAsyncSteps" "Tests priority with async steps"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Priority Default Values" "TestPriorityDefaultValue" "Tests default priority values"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Priority Negative Values" "TestPriorityNegativeValues" "Tests negative priority values"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

# Test Suite 2: Database Persistence
echo -e "\n${YELLOW}🗄️  Test Suite 2: Database Persistence${NC}"
echo "============================================="

run_test_suite "Database Persistence" "TestPriorityDatabasePersistence" "Tests priority values are stored in database"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Database Retrieval" "TestPriorityDatabaseRetrieval" "Tests priority values are correctly retrieved"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Database Index" "TestPriorityDatabaseIndex" "Tests priority-based queries use indexes"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Database Migration" "TestPriorityDatabaseMigration" "Tests database migration to add priority column"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

# Test Suite 3: Edge Cases
echo -e "\n${YELLOW}🔬 Test Suite 3: Edge Cases${NC}"
echo "============================================="

run_test_suite "Edge Cases" "TestPriorityEdgeCases" "Tests extreme priority values"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Identical Priorities" "TestPriorityWithIdenticalPriorities" "Tests steps with identical priorities"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Large Number of Steps" "TestPriorityWithLargeNumberOfSteps" "Tests priority with many steps (50+)"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Concurrent Execution" "TestPriorityWithConcurrentExecution" "Tests priority with high concurrency"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Step Failures" "TestPriorityWithStepFailure" "Tests priority behavior when steps fail"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Step Timeouts" "TestPriorityWithTimeout" "Tests priority behavior with step timeouts"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Retry Logic" "TestPriorityWithRetry" "Tests priority behavior with retry logic"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

run_test_suite "Complex Dependencies" "TestPriorityWithComplexDependencies" "Tests priority with complex dependency chains"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

# Test Suite 4: Coverage Analysis
echo -e "\n${YELLOW}📊 Test Suite 4: Coverage Analysis${NC}"
echo "============================================="

run_test_with_coverage "All Priority Tests" "TestPriority" "priority_coverage.out"
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

# Test Suite 5: Performance Benchmarks
echo -e "\n${YELLOW}⚡ Test Suite 5: Performance Benchmarks${NC}"
echo "============================================="

run_benchmarks
((total_tests++))
if [ $? -eq 0 ]; then ((passed_tests++)); else ((failed_tests++)); fi

# Generate coverage report
echo -e "\n${YELLOW}📈 Generating Coverage Report${NC}"
echo "============================================="

if [ -f "priority_coverage.out" ]; then
    echo "Generating HTML coverage report..."
    go tool cover -html=priority_coverage.out -o priority_coverage.html
    echo -e "${GREEN}✅ Coverage report generated: priority_coverage.html${NC}"
else
    echo -e "${RED}❌ Coverage file not found${NC}"
fi

# Final Results
echo -e "\n${YELLOW}📋 Final Test Results${NC}"
echo "============================================="
echo "Total Tests: $total_tests"
echo -e "Passed: ${GREEN}$passed_tests${NC}"
echo -e "Failed: ${RED}$failed_tests${NC}"

if [ $failed_tests -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All priority queue tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please check the output above.${NC}"
    exit 1
fi
