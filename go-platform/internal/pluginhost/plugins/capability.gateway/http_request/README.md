# capability.gateway/http_request

以 HTTP/HTTPS 對外呼叫 API（GET/POST/PUT/DELETE）。
- 入參（payload）：`method`、`url`、`headers`（map）、`body`（string）
- 回傳：`status`、`headers`、`body`（string）
- 觀測：`capability.request`（由宿主框架統一打點）
- Profiles：請於 `contracts/profiles/*` 設定 `egress.allowlist`、mTLS 憑證

對應 Tool：`detectviz.tools.http_request`（Python/ADK Remote Tool）