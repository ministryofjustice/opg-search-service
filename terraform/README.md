# terraform-search-service

Terraform module for deploying the search service into an ECS Cluster

### Usage

```hcl
module "search_service" {
  source                    = "github.com/ministryofjustice/opg-search-service//terraform"
  cloudwatch_log_group_name = aws_cloudwatch_log_group.sirius.name
  ecr_repository_url        = data.aws_ecr_repository.search_service.repository_url
  ecs_cluster_id            = aws_ecs_cluster.main.id
  execution_role_arn        = aws_iam_role.execution_role.arn
  tags                      = local.default_tags
  task_role_assume_policy   = data.aws_iam_policy_document.task_role_assume_policy
  vpc_id                    = data.aws_vpc.sirius.id
  vpc_private_subnet_ids    = data.aws_subnet_ids.private.ids
}
```
