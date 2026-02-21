#!/bin/bash

BASEURL="http://localhost:8000"
RESULTS_FILE="/home/butterfly_student/code/Mikrotik/mikrops_no-tenant/test_curl_results.json"
RESULTS=()

add_result() {
  local method="$1" path="$2" status="$3" success="$4" note="$5"
  local entry="{\"method\":\"${method}\",\"path\":\"${path}\",\"status\":${status},\"success\":${success},\"note\":\"${note}\"}"
  RESULTS+=("$entry")
  echo "  [${method}] ${path} => ${status} (${note})"
}

do_request() {
  local method="$1" path="$2" auth="$3" body="$4" note_prefix="$5"
  local curl_args=(-s -o /tmp/curl_body.txt -w '%{http_code}' -X "$method" "${BASEURL}${path}")
  curl_args+=(-H "Content-Type: application/json")
  if [ -n "$auth" ]; then
    curl_args+=(-H "Authorization: Bearer ${auth}")
  fi
  if [ -n "$body" ]; then
    curl_args+=(-d "$body")
  fi
  local status
  status=$(curl --max-time 15 "${curl_args[@]}" 2>/dev/null)
  if [ $? -ne 0 ] || [ -z "$status" ]; then
    status=0
  fi
  local resp
  resp=$(cat /tmp/curl_body.txt 2>/dev/null)
  local success="false"
  if [ "$status" -ge 200 ] 2>/dev/null && [ "$status" -lt 300 ] 2>/dev/null; then
    success="true"
  fi
  local note="${note_prefix:-ok}"
  # Escape double quotes and backslashes in note for JSON safety
  local short_resp
  short_resp=$(echo "$resp" | head -c 200 | tr -d '\n' | sed 's/\\/\\\\/g; s/"/\\"/g')
  if [ "$success" = "false" ] && [ "$status" -ne 0 ]; then
    note="${note} | ${short_resp}"
  fi
  add_result "$method" "$path" "$status" "$success" "$note"
  echo "$resp" > /tmp/curl_last_resp.txt
}

echo "============================================"
echo "  API Route Testing - $(date)"
echo "============================================"
echo ""

# ──────────────────────────────────────────────
# 1. AUTH - Login
# ──────────────────────────────────────────────
echo "=== 1. AUTH ==="

# Login
curl_args=(-s -o /tmp/curl_body.txt -w '%{http_code}' -X POST "${BASEURL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@mikrotik.local","password":"Admin@123"}')
LOGIN_STATUS=$(curl --max-time 15 "${curl_args[@]}" 2>/dev/null)
LOGIN_RESP=$(cat /tmp/curl_body.txt)

ACCESS_TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null)
REFRESH_TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('refresh_token',''))" 2>/dev/null)

if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "" ]; then
  add_result "POST" "/auth/login" "$LOGIN_STATUS" "true" "Login successful, got tokens"
  echo "    ACCESS_TOKEN=${ACCESS_TOKEN:0:40}..."
else
  add_result "POST" "/auth/login" "$LOGIN_STATUS" "false" "Login failed"
  echo "    WARN: No access token obtained. Authenticated routes will fail."
  echo "    Response: $LOGIN_RESP"
fi
TOKEN="$ACCESS_TOKEN"

# Register
RAND_EMAIL="postman-test-${RANDOM}@example.com"
do_request "POST" "/auth/register" "" \
  "{\"name\":\"Postman Test\",\"email\":\"${RAND_EMAIL}\",\"password\":\"Test@123\",\"role\":\"user\"}" \
  "Register new user"

# Refresh
if [ -n "$REFRESH_TOKEN" ]; then
  do_request "POST" "/auth/refresh" "" \
    "{\"refresh_token\":\"${REFRESH_TOKEN}\"}" \
    "Refresh token"
else
  add_result "POST" "/auth/refresh" 0 "false" "Skipped - no refresh token"
fi

# ──────────────────────────────────────────────
# 2. USER
# ──────────────────────────────────────────────
echo ""
echo "=== 2. USER ==="

do_request "GET" "/user/profile" "$TOKEN" "" "Get profile"

do_request "PUT" "/user/profile" "$TOKEN" \
  '{"name":"Admin","email":"admin@mikrotik.local","role":"admin","status":"active"}' \
  "Update profile"

do_request "POST" "/user/change-password" "$TOKEN" \
  '{"old_password":"Admin@123","new_password":"Admin@123"}' \
  "Change password"

# Logout - do this last in user section, then re-login
do_request "POST" "/user/logout" "$TOKEN" "" "Logout"

# Re-login to get fresh token
curl_args=(-s -o /tmp/curl_body.txt -w '%{http_code}' -X POST "${BASEURL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@mikrotik.local","password":"Admin@123"}')
RELOGIN_STATUS=$(curl --max-time 15 "${curl_args[@]}" 2>/dev/null)
RELOGIN_RESP=$(cat /tmp/curl_body.txt)
TOKEN=$(echo "$RELOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('access_token',''))" 2>/dev/null)
echo "  Re-logged in, new token: ${TOKEN:0:40}..."

# ──────────────────────────────────────────────
# 3. MIKROTIK ROUTER CRUD
# ──────────────────────────────────────────────
echo ""
echo "=== 3. MIKROTIK ROUTER CRUD ==="

ROUTER_ID_KNOWN="550e8400-e29b-41d4-a716-446655440001"

# Create router
do_request "POST" "/mikrotik" "$TOKEN" \
  '{"name":"Curl Test Router","address":"10.0.0.99","api_port":8728,"username":"admin","password":"test","use_ssl":false,"is_active":true}' \
  "Create router"
CREATED_ROUTER_RESP=$(cat /tmp/curl_last_resp.txt)
CREATED_ROUTER_ID=$(echo "$CREATED_ROUTER_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
echo "    Created router ID: $CREATED_ROUTER_ID"

# List routers
do_request "GET" "/mikrotik" "$TOKEN" "" "List routers"

# Get specific router
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}" "$TOKEN" "" "Get router by ID"

# Update router
do_request "PUT" "/mikrotik/${ROUTER_ID_KNOWN}" "$TOKEN" \
  '{"name":"Router Updated","address":"192.168.1.1","api_port":8728,"username":"admin","password":"test-password","use_ssl":false,"is_active":true}' \
  "Update router"

# Test connection
do_request "POST" "/mikrotik/${ROUTER_ID_KNOWN}/test" "$TOKEN" "" "Test router connection"

# Delete created router
if [ -n "$CREATED_ROUTER_ID" ] && [ "$CREATED_ROUTER_ID" != "" ]; then
  do_request "DELETE" "/mikrotik/${CREATED_ROUTER_ID}" "$TOKEN" "" "Delete created router"
else
  add_result "DELETE" "/mikrotik/{created_id}" 0 "false" "Skipped - no router was created"
fi

# ──────────────────────────────────────────────
# 4. BANDWIDTH PROFILES
# ──────────────────────────────────────────────
echo ""
echo "=== 4. BANDWIDTH PROFILES ==="

do_request "POST" "/bandwidth-profiles" "$TOKEN" \
  '{"profile_code":"CURL-TEST-10M","profile_name":"Curl Test 10Mbps","speed_up":"10M","speed_down":"10M","price":100000,"description":"Test profile from curl"}' \
  "Create bandwidth profile"
BP_RESP=$(cat /tmp/curl_last_resp.txt)
BP_ID=$(echo "$BP_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
echo "    Created bandwidth profile ID: $BP_ID"

do_request "GET" "/bandwidth-profiles" "$TOKEN" "" "List bandwidth profiles"

do_request "GET" "/bandwidth-profiles/code/CURL-TEST-10M" "$TOKEN" "" "Get profile by code"

if [ -n "$BP_ID" ] && [ "$BP_ID" != "" ]; then
  do_request "GET" "/bandwidth-profiles/${BP_ID}" "$TOKEN" "" "Get profile by ID"

  do_request "PUT" "/bandwidth-profiles/${BP_ID}" "$TOKEN" \
    '{"profile_code":"CURL-TEST-10M","profile_name":"Curl Test 10Mbps Updated","speed_up":"20M","speed_down":"20M","price":150000,"description":"Updated test profile"}' \
    "Update bandwidth profile"

  do_request "DELETE" "/bandwidth-profiles/${BP_ID}" "$TOKEN" "" "Delete bandwidth profile"
else
  add_result "GET" "/bandwidth-profiles/{id}" 0 "false" "Skipped - no profile created"
  add_result "PUT" "/bandwidth-profiles/{id}" 0 "false" "Skipped - no profile created"
  add_result "DELETE" "/bandwidth-profiles/{id}" 0 "false" "Skipped - no profile created"
fi

# ──────────────────────────────────────────────
# 5. CUSTOMERS
# ──────────────────────────────────────────────
echo ""
echo "=== 5. CUSTOMERS ==="

do_request "POST" "/customers" "$TOKEN" \
  '{"customer_code":"CUST-CURL-001","full_name":"Curl Test Customer","email":"curl-test@example.com","phone":"08123456789","address":"Test Address 123","bandwidth_profile_code":"10M-BASIC","pppoe_username":"curl_test_user","pppoe_password":"test123","router_id":"550e8400-e29b-41d4-a716-446655440001","status":"active"}' \
  "Create customer"
CUST_RESP=$(cat /tmp/curl_last_resp.txt)
CUST_ID=$(echo "$CUST_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
echo "    Created customer ID: $CUST_ID"

do_request "GET" "/customers" "$TOKEN" "" "List customers"

do_request "GET" "/customers/code/CUST-CURL-001" "$TOKEN" "" "Get customer by code"

if [ -n "$CUST_ID" ] && [ "$CUST_ID" != "" ]; then
  do_request "GET" "/customers/${CUST_ID}" "$TOKEN" "" "Get customer by ID"

  do_request "PUT" "/customers/${CUST_ID}" "$TOKEN" \
    '{"customer_code":"CUST-CURL-001","full_name":"Curl Test Customer Updated","email":"curl-test-updated@example.com","phone":"08123456789","address":"Updated Address","bandwidth_profile_code":"10M-BASIC","pppoe_username":"curl_test_user","pppoe_password":"test123","router_id":"550e8400-e29b-41d4-a716-446655440001","status":"active"}' \
    "Update customer"

  do_request "POST" "/customers/${CUST_ID}/status" "$TOKEN" \
    '{"status":"suspended"}' \
    "Change customer status"

  do_request "DELETE" "/customers/${CUST_ID}" "$TOKEN" "" "Delete customer"
else
  add_result "GET" "/customers/{id}" 0 "false" "Skipped - no customer created"
  add_result "PUT" "/customers/{id}" 0 "false" "Skipped - no customer created"
  add_result "POST" "/customers/{id}/status" 0 "false" "Skipped - no customer created"
  add_result "DELETE" "/customers/{id}" 0 "false" "Skipped - no customer created"
fi

# ──────────────────────────────────────────────
# 6. INVOICES
# ──────────────────────────────────────────────
echo ""
echo "=== 6. INVOICES ==="

# Create a temp customer for invoice testing
do_request "POST" "/customers" "$TOKEN" \
  '{"customer_code":"CUST-INV-TEMP","full_name":"Invoice Test Customer","email":"invoice-test@example.com","phone":"08123456799","address":"Invoice Test Address","bandwidth_profile_code":"10M-BASIC","pppoe_username":"inv_test_user","pppoe_password":"test123","router_id":"550e8400-e29b-41d4-a716-446655440001","status":"active"}' \
  "Create temp customer for invoices"
INV_CUST_RESP=$(cat /tmp/curl_last_resp.txt)
INV_CUST_ID=$(echo "$INV_CUST_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)

if [ -n "$INV_CUST_ID" ] && [ "$INV_CUST_ID" != "" ]; then
  do_request "POST" "/invoices" "$TOKEN" \
    "{\"customer_id\":\"${INV_CUST_ID}\",\"amount\":100000,\"due_date\":\"2026-03-01\",\"description\":\"Test invoice from curl\"}" \
    "Create invoice"
  INV_RESP=$(cat /tmp/curl_last_resp.txt)
  INV_ID=$(echo "$INV_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
  echo "    Created invoice ID: $INV_ID"
else
  do_request "POST" "/invoices" "$TOKEN" \
    '{"customer_id":"00000000-0000-0000-0000-000000000001","amount":100000,"due_date":"2026-03-01","description":"Test invoice"}' \
    "Create invoice (fallback customer)"
  INV_RESP=$(cat /tmp/curl_last_resp.txt)
  INV_ID=$(echo "$INV_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
fi

do_request "GET" "/invoices" "$TOKEN" "" "List invoices"

if [ -n "$INV_ID" ] && [ "$INV_ID" != "" ]; then
  do_request "GET" "/invoices/${INV_ID}" "$TOKEN" "" "Get invoice by ID"
  do_request "DELETE" "/invoices/${INV_ID}" "$TOKEN" "" "Delete invoice"
else
  add_result "GET" "/invoices/{id}" 0 "false" "Skipped - no invoice created"
  add_result "DELETE" "/invoices/{id}" 0 "false" "Skipped - no invoice created"
fi

# Cleanup temp customer
if [ -n "$INV_CUST_ID" ] && [ "$INV_CUST_ID" != "" ]; then
  do_request "DELETE" "/customers/${INV_CUST_ID}" "$TOKEN" "" "Delete temp invoice customer"
fi

# ──────────────────────────────────────────────
# 7. PAYMENTS
# ──────────────────────────────────────────────
echo ""
echo "=== 7. PAYMENTS ==="

# Create temp customer + invoice for payment
do_request "POST" "/customers" "$TOKEN" \
  '{"customer_code":"CUST-PAY-TEMP","full_name":"Payment Test Customer","email":"pay-test@example.com","phone":"08123456800","address":"Payment Test Address","bandwidth_profile_code":"10M-BASIC","pppoe_username":"pay_test_user","pppoe_password":"test123","router_id":"550e8400-e29b-41d4-a716-446655440001","status":"active"}' \
  "Create temp customer for payments"
PAY_CUST_RESP=$(cat /tmp/curl_last_resp.txt)
PAY_CUST_ID=$(echo "$PAY_CUST_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)

PAY_INV_ID=""
if [ -n "$PAY_CUST_ID" ] && [ "$PAY_CUST_ID" != "" ]; then
  do_request "POST" "/invoices" "$TOKEN" \
    "{\"customer_id\":\"${PAY_CUST_ID}\",\"amount\":100000,\"due_date\":\"2026-03-01\",\"description\":\"Payment test invoice\"}" \
    "Create temp invoice for payments"
  PAY_INV_RESP=$(cat /tmp/curl_last_resp.txt)
  PAY_INV_ID=$(echo "$PAY_INV_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
fi

if [ -n "$PAY_INV_ID" ] && [ "$PAY_INV_ID" != "" ]; then
  do_request "POST" "/payments" "$TOKEN" \
    "{\"invoice_id\":\"${PAY_INV_ID}\",\"amount\":100000,\"payment_method\":\"cash\",\"payment_date\":\"2026-02-21\",\"notes\":\"Test payment from curl\"}" \
    "Create payment"
  PAY_RESP=$(cat /tmp/curl_last_resp.txt)
  PAY_ID=$(echo "$PAY_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
  echo "    Created payment ID: $PAY_ID"
else
  do_request "POST" "/payments" "$TOKEN" \
    '{"invoice_id":"00000000-0000-0000-0000-000000000001","amount":100000,"payment_method":"cash","payment_date":"2026-02-21","notes":"Test"}' \
    "Create payment (fallback)"
  PAY_RESP=$(cat /tmp/curl_last_resp.txt)
  PAY_ID=$(echo "$PAY_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
fi

do_request "GET" "/payments" "$TOKEN" "" "List payments"

if [ -n "$PAY_ID" ] && [ "$PAY_ID" != "" ]; then
  do_request "GET" "/payments/${PAY_ID}" "$TOKEN" "" "Get payment by ID"
  do_request "DELETE" "/payments/${PAY_ID}" "$TOKEN" "" "Delete payment"
else
  add_result "GET" "/payments/{id}" 0 "false" "Skipped - no payment created"
  add_result "DELETE" "/payments/{id}" 0 "false" "Skipped - no payment created"
fi

# Cleanup
if [ -n "$PAY_INV_ID" ] && [ "$PAY_INV_ID" != "" ]; then
  do_request "DELETE" "/invoices/${PAY_INV_ID}" "$TOKEN" "" "Delete temp payment invoice"
fi
if [ -n "$PAY_CUST_ID" ] && [ "$PAY_CUST_ID" != "" ]; then
  do_request "DELETE" "/customers/${PAY_CUST_ID}" "$TOKEN" "" "Delete temp payment customer"
fi

# ──────────────────────────────────────────────
# 8. PPPOE LEGACY
# ──────────────────────────────────────────────
echo ""
echo "=== 8. PPPOE LEGACY ==="

do_request "GET" "/pppoe/secrets?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" "" "List PPPoE secrets"
do_request "GET" "/pppoe/profiles?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" "" "List PPPoE profiles"
do_request "GET" "/pppoe/sessions/active?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" "" "List active PPPoE sessions"
do_request "GET" "/pppoe/sessions/inactive?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" "" "List inactive PPPoE sessions"

# ──────────────────────────────────────────────
# 9. QUEUES LEGACY
# ──────────────────────────────────────────────
echo ""
echo "=== 9. QUEUES LEGACY ==="

do_request "GET" "/queues?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" "" "List queues"

# ──────────────────────────────────────────────
# 10. IP POOLS
# ──────────────────────────────────────────────
echo ""
echo "=== 10. IP POOLS ==="

do_request "GET" "/ip-pools?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" "" "List IP pools"

# ──────────────────────────────────────────────
# 11. MIKROTIK OPERATIONS
# ──────────────────────────────────────────────
echo ""
echo "=== 11. MIKROTIK OPERATIONS ==="

do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/pppoe/secrets" "$TOKEN" "" "MikroTik PPPoE secrets"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/pppoe/profiles" "$TOKEN" "" "MikroTik PPPoE profiles"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/pppoe/sessions/active" "$TOKEN" "" "MikroTik active PPPoE sessions"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/queues" "$TOKEN" "" "MikroTik queues"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/ip-pools" "$TOKEN" "" "MikroTik IP pools"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/hotspot/profiles" "$TOKEN" "" "MikroTik hotspot profiles"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/hotspot/users" "$TOKEN" "" "MikroTik hotspot users"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/hotspot/sessions" "$TOKEN" "" "MikroTik hotspot sessions"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/hotspot/sessions/stats" "$TOKEN" "" "MikroTik hotspot session stats"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/hotspot/sales" "$TOKEN" "" "MikroTik hotspot sales"
do_request "GET" "/mikrotik/${ROUTER_ID_KNOWN}/hotspot/sales/revenue" "$TOKEN" "" "MikroTik hotspot sales revenue"

# ──────────────────────────────────────────────
# 12. WEBHOOKS
# ──────────────────────────────────────────────
echo ""
echo "=== 12. WEBHOOKS ==="

do_request "POST" "/webhooks/pppoe/on-up" "" \
  '{"user":"test","ip-address":"10.0.0.1","caller-id":"AA:BB:CC:DD:EE:FF","session-id":"123"}' \
  "Webhook PPPoE on-up"

do_request "POST" "/webhooks/pppoe/on-down" "" \
  '{"user":"test","ip-address":"10.0.0.1","caller-id":"AA:BB:CC:DD:EE:FF","session-id":"123"}' \
  "Webhook PPPoE on-down"

# ──────────────────────────────────────────────
# 13. PING
# ──────────────────────────────────────────────
echo ""
echo "=== 13. PING ==="

do_request "POST" "/ping?router_id=${ROUTER_ID_KNOWN}" "$TOKEN" \
  '{"address":"8.8.8.8","count":1}' \
  "Ping via router"

# ──────────────────────────────────────────────
# BUILD FINAL JSON
# ──────────────────────────────────────────────
echo ""
echo "============================================"
echo "  Building results JSON..."
echo "============================================"

# Count stats
TOTAL=${#RESULTS[@]}
SUCCESS_COUNT=0
FAIL_COUNT=0

{
  echo "{"
  echo "  \"test_date\": \"$(date -Iseconds)\","
  echo "  \"base_url\": \"${BASEURL}\","
  echo "  \"total_tests\": ${TOTAL},"

  # Build array
  echo "  \"results\": ["
  for i in "${!RESULTS[@]}"; do
    if [ $i -lt $((TOTAL - 1)) ]; then
      echo "    ${RESULTS[$i]},"
    else
      echo "    ${RESULTS[$i]}"
    fi
    # Count success/fail
    if echo "${RESULTS[$i]}" | grep -q '"success":true'; then
      SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
      FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
  done
  echo "  ],"
  echo "  \"summary\": {"
  echo "    \"total\": ${TOTAL},"
  echo "    \"success\": ${SUCCESS_COUNT},"
  echo "    \"failed\": ${FAIL_COUNT}"
  echo "  }"
  echo "}"
} > "$RESULTS_FILE"

echo ""
echo "============================================"
echo "  SUMMARY"
echo "============================================"
echo "  Total tests:  ${TOTAL}"
echo "  Successful:   ${SUCCESS_COUNT}"
echo "  Failed:       ${FAIL_COUNT}"
echo "  Results file: ${RESULTS_FILE}"
echo "============================================"
