import "server-only";

import { proxyGatewayRequest } from "../../../../lib/gateway-proxy";

const gatewayOrigin = process.env.GATEWAY_ORIGIN ?? "http://127.0.0.1:3000";

function handler(request: Request) {
  return proxyGatewayRequest(request, gatewayOrigin);
}

export { handler as GET, handler as POST };
