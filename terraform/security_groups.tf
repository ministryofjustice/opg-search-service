resource "aws_security_group" "search_service" {
  name_prefix            = "search-service-${terraform.workspace}-"
  revoke_rules_on_delete = true
  vpc_id                 = var.vpc_id
  description            = "Search Service ECS Service"

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(
    var.tags,
    map("Name", "search-service-${terraform.workspace}")
  )
}

resource "aws_security_group_rule" "search_service_ecs_egress" {
  count             = local.enabled
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.search_service.id
  description       = "Outbound Search Service"
}

resource "aws_security_group_rule" "search_service_ecs_api" {
  count                    = local.enabled
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 80
  to_port                  = 80
  source_security_group_id = var.api_security_group
  security_group_id        = aws_security_group.search_service.id
}
