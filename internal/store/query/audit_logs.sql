-- name: CreateAuditLog :exec
INSERT INTO audit.audit_logs (
  admin_id, action, resource, resource_id,
  details, status, ip_address, user_agent, request_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
