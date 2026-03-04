output "aws_region" {
  description = "AWS region"
  value       = var.aws_region
}

output "user_pool_id" {
  description = "Cognito User Pool ID  → COGNITO_USER_POOL_ID env var"
  value       = aws_cognito_user_pool.main.id
}

output "client_id" {
  description = "Cognito App Client ID  → COGNITO_CLIENT_ID / VITE_COGNITO_CLIENT_ID env var"
  value       = aws_cognito_user_pool_client.main.id
}

output "cognito_domain" {
  description = "Cognito Hosted UI domain  → VITE_COGNITO_DOMAIN env var"
  value       = "${aws_cognito_user_pool_domain.main.domain}.auth.${var.aws_region}.amazoncognito.com"
}

output "jwks_url" {
  description = "JWKS endpoint used by the backend to verify ID tokens"
  value       = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.main.id}/.well-known/jwks.json"
}
