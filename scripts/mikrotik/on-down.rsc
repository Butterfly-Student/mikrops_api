# On Down Script
# Replace http://your-app-url with your actual application URL
# Replace ROUTER_ID with the ID of this router in your database

:local url "http://your-app-url/webhooks/pppoe/on-down?router_id=ROUTER_ID"

:local user $"user"
:local ip $"local-address"
:local callerId $"caller-id"
:local sessionId $"session-id"
:local interface $"interface"
:local uptime $"uptime"
:local bytesIn $"bytes-in"
:local bytesOut $"bytes-out"
:local packetsIn $"packets-in"
:local packetsOut $"packets-out"

# Construct JSON payload
:local payload ("{\"user\":\"" . $user . "\",\"ip-address\":\"" . $ip . "\",\"caller-id\":\"" . $callerId . "\",\"session-id\":\"" . $sessionId . "\",\"interface\":\"" . $interface . "\",\"uptime\":\"" . $uptime . "\",\"bytes-in\":" . $bytesIn . ",\"bytes-out\":" . $bytesOut . ",\"packets-in\":" . $packetsIn . ",\"packets-out\":" . $packetsOut . "}")

/tool fetch url=$url http-method=post http-header-field="Content-Type: application/json" http-data=$payload keep-result=no
