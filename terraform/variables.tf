variable "cloudwatch_log_group_name" {
  type        = string
  description = "Name of the CloudWatch Log Group for logs"
}

variable "ecr_repository_url" {
  type        = string
  description = "URL of the search service ECR Repository"
}

variable "ecs_cluster_id" {
  type        = string
  description = "The ID of the ECS Cluster to deploy the service into"
}

variable "elasticsearch_domain_arn" {
  type        = string
  description = "ARN for the elasticsearch domain"
}

variable "execution_role_arn" {
  type        = string
  description = "ARN of the ECS Execution Role"
}

variable "task_role_assume_policy" {
  description = "Assume Role Policy Object"
}

variable "tags" {
  type        = map
  description = "List of tags to apply to all resources"
}

variable "task_count" {
  type        = number
  description = "The number of search service tasks to run"
  default     = 1
}

variable "vpc_id" {
  type        = string
  description = "ID of the VPC the service will be deployed into"
}

variable "vpc_private_subnet_ids" {
  type        = list(string)
  description = "List of the VPC Private Subnets"
}
