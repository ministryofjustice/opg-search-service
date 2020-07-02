resource "aws_ecs_service" "search_service" {
  name            = "search-service"
  cluster         = var.ecs_cluster_id
  task_definition = aws_ecs_task_definition.search_service.arn
  desired_count   = var.task_count
  launch_type     = "FARGATE"
  tags = merge(var.tags,
    map("Name", "search-service-ecs-service-${terraform.workspace}")
  )


  deployment_controller {
    type = "ECS"
  }

  network_configuration {
    security_groups  = [aws_security_group.search_service.id]
    subnets          = var.vpc_private_subnet_ids
    assign_public_ip = false
  }

  service_registries {
    registry_arn = aws_service_discovery_service.search_service.arn
  }
}

resource "aws_service_discovery_service" "search_service" {
  name = "search-service"

  dns_config {
    namespace_id   = var.service_discovery_namespace
    routing_policy = "MULTIVALUE"
    dns_records {
      ttl  = 10
      type = "A"
    }
  }
}

resource "aws_ecs_task_definition" "search_service" {
  family                   = "search-service-${terraform.workspace}"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 1024
  memory                   = 2048
  container_definitions    = "[${local.search_service}]"
  task_role_arn            = aws_iam_role.search_service.arn
  execution_role_arn       = var.execution_role
  tags = merge(var.default_tags,
    { "Role" = "search-service-ecs-task" },
  )
}

variable "search_service_tag" {
  type    = string
  default = "master-dba41e7"
}

locals {
  search_service = jsonencode({
    name      = "search-service",
    cpu       = 0,
    essential = true,
    image     = "${var.ecr_repository_url}:${var.search_service_tag`}",
    portMappings = [{
      containerPort = 80,
      hostPort      = 80,
      protocol      = "tcp"
    }],
    healthCheck = {
      command = [
        "CMD-SHELL",
        "curl -f http://localhost/services/search-service/health-check || exit 1"
      ],
      startPeriod = 30,
      interval    = 15,
      timeout     = 10,
      retries     = 3
    },
    logConfiguration = {
      logDriver = "awslogs",
      options = {
        awslogs-group         = var.cloudwatch_log_group_name,
        awslogs-region        = "eu-west-1",
        awslogs-stream-prefix = "search-service"
      }
    },
    environment = [
      {
        name  = "PATH_PREFIX",
        value = "/services/search-service"
      },
    ],
    secrets     = [],
    mountPoints = [],
    volumesFrom = [],
  })
}

