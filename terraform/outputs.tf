output "service_name" {
  value       = aws_ecs_service.search_service.name : ""
  description = "Name of the service for use in task stabilizer"
}

output "task_count" {
  value       = aws_ecs_service.search_service.desired_count : 0
  description = "Number of tasks expected to be running in the ECS Service"
}
