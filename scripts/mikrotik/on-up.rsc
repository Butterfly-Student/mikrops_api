# On Up Script
# Replace http://your-app-url with your actual application URL
# Replace ROUTER_ID with the ID of this router in your database

:local url "http://your-app-url/webhooks/pppoe/on-up?router_id=ROUTER_ID"

:local user $"user"
:local ip $"local-address"
:local callerId $"caller-id"
:local sessionId $"session-id"
:local interface $"interface"

# Construct JSON payload
# Note: RouterOS scripting JSON support is limited, constructing string manually
:local payload ("{\"user\":\"" . $user . "\",\"ip-address\":\"" . $ip . "\",\"caller-id\":\"" . $callerId . "\",\"session-id\":\"" . $sessionId . "\",\"interface\":\"" . $interface . "\"}")

/tool fetch url=$url http-method=post http-header-field="Content-Type: application/json" http-data=$payload keep-result=no
