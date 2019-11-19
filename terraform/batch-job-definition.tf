resource "aws_batch_job_definition" "publish" {
  name = "tasque-batch-job"
  type = "container"

  retry_strategy {
    attempts = 2
  }

  timeout {
    attempt_duration_seconds = 18000
  }

  container_properties = <<CONTAINER_PROPERTIES
{
  "image": "skycatch/tasque:instance-termination-test",
  "jobRoleArn": "${aws_iam_role.dtm-worker-role.arn}",
  "memory": 1024,
  "vcpus": 1
}
CONTAINER_PROPERTIES
}
