# Verify authentication flow: login, password change, token validation
set -euo pipefail

# Skip auto-login — this test manages auth manually
SKIP_AUTH=true
source "$(dirname "$0")/helpers.sh"

echo "  Step 1: Verify unauthenticated request returns 401"
assert_http_code_no_auth "Unauthenticated GET /api/v1/config → 401" "401" GET "/api/v1/config"

echo "  Step 2: Login with initial password"
INIT_PASS=$(vm_exec "cat /tmp/enclave-wizard-init-pass 2>/dev/null")
if [ -z "${INIT_PASS}" ]; then
    echo "    ✗ Could not read initial password"
    exit 1
fi
echo "    ✓ Read initial password"

LOGIN_RESULT=$(vm_exec "curl -sfk -X POST https://localhost:3443/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"${INIT_PASS}\"}'")
FIRST_TOKEN=$(echo "${LOGIN_RESULT}" | jq -r '.token')
MUST_CHANGE=$(echo "${LOGIN_RESULT}" | jq -r '.mustChangePassword')
assert_ok "Got token from login" [ -n "${FIRST_TOKEN}" ]
assert_field "mustChangePassword is true" '.mustChangePassword' "true" "${LOGIN_RESULT}"
echo "    ✓ Login successful, mustChangePassword=${MUST_CHANGE}"

echo "  Step 3: Authenticated request with token succeeds"
assert_http_code_with_token "GET /api/v1/config with token → 200" "200" GET "/api/v1/config" "${FIRST_TOKEN}"

echo "  Step 4: Change password"
NEW_PASS="e2e-auth-test-$(date +%s)"
CHANGE_RESULT=$(vm_exec "curl -sfk -X POST https://localhost:3443/api/v1/auth/password -H 'Content-Type: application/json' -H 'Authorization: Bearer ${FIRST_TOKEN}' -d '{\"currentPassword\":\"${INIT_PASS}\",\"newPassword\":\"${NEW_PASS}\"}'")
NEW_TOKEN=$(echo "${CHANGE_RESULT}" | jq -r '.token')
assert_ok "Got new token from password change" [ -n "${NEW_TOKEN}" ]
echo "    ✓ Password changed"

echo "  Step 5: Old token is rejected"
assert_http_code_with_token "GET /api/v1/config with old token → 401" "401" GET "/api/v1/config" "${FIRST_TOKEN}"

echo "  Step 6: New token works"
assert_http_code_with_token "GET /api/v1/config with new token → 200" "200" GET "/api/v1/config" "${NEW_TOKEN}"

echo "  Step 7: Login with old password fails"
OLD_LOGIN_STATUS=$(vm_exec "curl -sk -o /dev/null -w '%{http_code}' -X POST https://localhost:3443/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"${INIT_PASS}\"}'")
if [ "${OLD_LOGIN_STATUS}" = "401" ]; then
    echo "    ✓ Login with old password rejected (401)"
else
    echo "    ✗ Login with old password should fail (expected 401, got ${OLD_LOGIN_STATUS})"
    exit 1
fi

echo "  Step 8: Login with new password succeeds"
NEW_LOGIN_RESULT=$(vm_exec "curl -sfk -X POST https://localhost:3443/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"${NEW_PASS}\"}'")
FINAL_TOKEN=$(echo "${NEW_LOGIN_RESULT}" | jq -r '.token')
assert_ok "Got token from new password login" [ -n "${FINAL_TOKEN}" ]
assert_http_code_with_token "GET /api/v1/config with final token → 200" "200" GET "/api/v1/config" "${FINAL_TOKEN}"

# Persist the changed password so subsequent tests can auto-login
vm_exec "echo '${NEW_PASS}' > /tmp/enclave-wizard-current-pass"
echo "    ✓ Persisted new password for subsequent tests"

echo "  All auth checks passed"
