variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "app_name" {
  type    = string
  default = "pathplanner-lab"
}

variable "env" {
  type    = string
  default = "prod"
}

variable "callback_urls" {
  description = "Allowed OAuth callback URLs (e.g. https://app.example.com/auth/callback)"
  type        = list(string)
  default     = ["http://localhost:5173/auth/callback"]
}

variable "logout_urls" {
  description = "Allowed logout URLs"
  type        = list(string)
  default     = ["http://localhost:5173/login"]
}
