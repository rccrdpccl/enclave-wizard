# Test plugin validation: unknown plugins rejected, valid combinations accepted
set -euo pipefail
source "$(dirname "$0")/helpers.sh"

echo "  Step 1: Single unknown plugin is rejected"
RESULT=$(api_post "/api/v1/plugins/validate" '{"plugins":["bogus-plugin"]}')
assert_contains "valid is false" '"valid":false' echo "${RESULT}"
assert_contains "error message" "unknown plugin: bogus-plugin" echo "${RESULT}"

echo "  Step 2: Known + unknown mix is rejected"
RESULT=$(api_post "/api/v1/plugins/validate" '{"plugins":["lvms","nonexistent","odf"]}')
assert_contains "valid is false" '"valid":false' echo "${RESULT}"
assert_contains "error mentions nonexistent" "unknown plugin: nonexistent" echo "${RESULT}"

echo "  Step 3: Multiple unknowns each get their own error"
RESULT=$(api_post "/api/v1/plugins/validate" '{"plugins":["fake-1","fake-2"]}')
assert_contains "valid is false" '"valid":false' echo "${RESULT}"
assert_contains "error mentions fake-1" "unknown plugin: fake-1" echo "${RESULT}"
assert_contains "error mentions fake-2" "unknown plugin: fake-2" echo "${RESULT}"

echo "  Step 4: All valid plugins pass"
RESULT=$(api_post "/api/v1/plugins/validate" '{"plugins":["lvms","nvidia-gpu","openshift-ai"]}')
assert_contains "valid is true" '"valid":true' echo "${RESULT}"

echo "  Step 5: Single valid plugin passes"
RESULT=$(api_post "/api/v1/plugins/validate" '{"plugins":["odf"]}')
assert_contains "valid is true" '"valid":true' echo "${RESULT}"

echo "  Step 6: Empty plugins array returns 422"
assert_http_code "Empty plugins list" "422" "POST" "/api/v1/plugins/validate" '{"plugins":[]}'

echo "  Step 7: Missing plugins field returns 422"
assert_http_code "Missing plugins field" "422" "POST" "/api/v1/plugins/validate" '{}'

echo "  Step 8: All 4 known plugins pass"
RESULT=$(api_post "/api/v1/plugins/validate" '{"plugins":["lvms","odf","nvidia-gpu","openshift-ai"]}')
assert_contains "valid is true" '"valid":true' echo "${RESULT}"

echo "  All checks passed"
