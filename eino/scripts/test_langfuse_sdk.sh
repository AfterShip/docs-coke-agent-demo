#!/bin/bash

# Test runner for Langfuse SDK unit tests
# This script runs tests independently to avoid module dependency issues

set -e

echo "🧪 Running Langfuse SDK Unit Tests"
echo "=================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Track test results
TOTAL_TESTS=0
PASSED_TESTS=0

run_test() {
    local test_name="$1"
    local test_files="$2"
    
    echo -e "${BLUE}Running: $test_name${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_files"; then
        echo -e "${GREEN}✅ $test_name PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $test_name FAILED${NC}"
    fi
    echo ""
}

echo "Testing Internal Utilities..."
echo "-----------------------------"

# ID generation tests
run_test "ID Generation Tests" \
    "go test -run TestGenerateNanoid pkg/langfuse/internal/utils/ids_test.go pkg/langfuse/internal/utils/ids.go"

run_test "ID Validation Tests" \
    "go test -run TestIsValidNanoid pkg/langfuse/internal/utils/ids_test.go pkg/langfuse/internal/utils/ids.go"

run_test "ID Utility Tests" \
    "go test -run TestGenerateTraceID pkg/langfuse/internal/utils/ids_test.go pkg/langfuse/internal/utils/ids.go"

# Validation tests
run_test "Validation Error Tests" \
    "go test -run TestValidationError pkg/langfuse/internal/utils/validation_test.go pkg/langfuse/internal/utils/validation.go pkg/langfuse/internal/utils/errors.go pkg/langfuse/internal/utils/conversions.go pkg/langfuse/internal/utils/ids.go"

run_test "String Validation Tests" \
    "go test -run TestValidateString pkg/langfuse/internal/utils/validation_test.go pkg/langfuse/internal/utils/validation.go pkg/langfuse/internal/utils/errors.go pkg/langfuse/internal/utils/conversions.go pkg/langfuse/internal/utils/ids.go"

run_test "Email Validation Tests" \
    "go test -run TestValidateEmail pkg/langfuse/internal/utils/validation_test.go pkg/langfuse/internal/utils/validation.go pkg/langfuse/internal/utils/errors.go pkg/langfuse/internal/utils/conversions.go pkg/langfuse/internal/utils/ids.go"

run_test "Numeric Validation Tests" \
    "go test -run TestValidateNumericRange pkg/langfuse/internal/utils/validation_test.go pkg/langfuse/internal/utils/validation.go pkg/langfuse/internal/utils/errors.go pkg/langfuse/internal/utils/conversions.go pkg/langfuse/internal/utils/ids.go"

run_test "Metadata Validation Tests" \
    "go test -run TestValidateMetadata pkg/langfuse/internal/utils/validation_test.go pkg/langfuse/internal/utils/validation.go pkg/langfuse/internal/utils/errors.go pkg/langfuse/internal/utils/conversions.go pkg/langfuse/internal/utils/ids.go"

# Conversion tests
run_test "Type Conversion Tests" \
    "go test -run TestToFloat64 pkg/langfuse/internal/utils/conversions_test.go pkg/langfuse/internal/utils/conversions.go"

run_test "Time Conversion Tests" \
    "go test -run TestParseTime pkg/langfuse/internal/utils/conversions_test.go pkg/langfuse/internal/utils/conversions.go"

run_test "String Helper Tests" \
    "go test -run TestToString pkg/langfuse/internal/utils/conversions_test.go pkg/langfuse/internal/utils/conversions.go"

run_test "Metadata Helper Tests" \
    "go test -run TestMergeMetadata pkg/langfuse/internal/utils/conversions_test.go pkg/langfuse/internal/utils/conversions.go"

echo "Testing Core Data Models..."
echo "---------------------------"

# Usage model tests
run_test "Usage JSON Serialization Tests" \
    "go test -run TestUsage_JSONSerialization pkg/langfuse/api/resources/commons/types/usage_test.go pkg/langfuse/api/resources/commons/types/usage.go"

run_test "Usage Calculation Tests" \
    "go test -run TestUsage_CalculateTotalTokens pkg/langfuse/api/resources/commons/types/usage_test.go pkg/langfuse/api/resources/commons/types/usage.go"

run_test "Usage Constructor Tests" \
    "go test -run TestNewUsage pkg/langfuse/api/resources/commons/types/usage_test.go pkg/langfuse/api/resources/commons/types/usage.go"

# Trace model tests
run_test "Trace JSON Serialization Tests" \
    "go test -run TestTrace_JSONSerialization pkg/langfuse/api/resources/commons/types/trace_test.go pkg/langfuse/api/resources/commons/types/trace.go"

run_test "Trace Request Tests" \
    "go test -run TestTraceCreateRequest pkg/langfuse/api/resources/commons/types/trace_test.go pkg/langfuse/api/resources/commons/types/trace.go"

# Observation model tests
run_test "Observation Constants Tests" \
    "go test -run TestObservationType_Constants pkg/langfuse/api/resources/commons/types/observation_test.go pkg/langfuse/api/resources/commons/types/observation.go pkg/langfuse/api/resources/commons/types/usage.go"

run_test "Observation JSON Serialization Tests" \
    "go test -run TestObservation_JSONSerialization pkg/langfuse/api/resources/commons/types/observation_test.go pkg/langfuse/api/resources/commons/types/observation.go pkg/langfuse/api/resources/commons/types/usage.go"

# Score model tests
run_test "Score Constants Tests" \
    "go test -run TestScoreDataType_Constants pkg/langfuse/api/resources/commons/types/score_test.go pkg/langfuse/api/resources/commons/types/score.go"

run_test "Score JSON Serialization Tests" \
    "go test -run TestScore_JSONSerialization pkg/langfuse/api/resources/commons/types/score_test.go pkg/langfuse/api/resources/commons/types/score.go"

run_test "Score Helper Functions Tests" \
    "go test -run TestNumericScore pkg/langfuse/api/resources/commons/types/score_test.go pkg/langfuse/api/resources/commons/types/score.go"

echo "=================================="
echo -e "${BLUE}Test Summary${NC}"
echo "=================================="

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "${GREEN}🎉 All tests passed! ($PASSED_TESTS/$TOTAL_TESTS)${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. ($PASSED_TESTS/$TOTAL_TESTS passed)${NC}"
    exit 1
fi