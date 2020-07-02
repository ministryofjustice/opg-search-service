resource "aws_iam_role" "search_service" {
  name_prefix        = "search-service-${terraform.workspace}-"
  assume_role_policy = var.task_role_assume_policy.json
  tags               = var.tags
}

resource "aws_iam_role_policy" "search_service_elasticsearch_access" {
  name   = "search-service-elasticsearch-access.${terraform.workspace}"
  role   = aws_iam_role.search_service.name
  policy = data.aws_iam_policy_document.search_service_elasticsearch_access.json
}

data "aws_iam_policy_document" "search_service_elasticsearch_access" {

  statement {
    effect = "Allow"
    sid    = "SearchServiceElasticSearchAccess${replace(terraform.workspace, "-", "")}"
    resources = [
      var.elasticsearch_domain_arn,
      "${var.elasticsearch_domain_arn}/*"
    ]

    actions = [
      "es:DescribeElasticsearchDomain",
      "es:ESHttpDelete",
      "es:ESHttpGet",
      "es:ESHttpHead",
      "es:ESHttpPatch",
      "es:ESHttpPost",
      "es:ESHttpPut",
    ]
  }
}
