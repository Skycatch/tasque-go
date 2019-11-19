resource "aws_ecs_task_definition" "docker-task-test" {
  family = "tasque-instance-interruption-test"
  task_role_arn = "${aws_iam_role.dtm-worker-role.arn}"
  execution_role_arn = "${aws_iam_role.dtm-worker-execution-role.arn}"
  container_definitions = <<EOF
[
  {
    "name": "test-tasque-instance-termination",
    "image": "skycatch/tasque:instance-termination-test",
    "cpu": 1024,
    "memory": 256,
    "essential": true,
    "environment":
    [
      {
        "name": "TASK_PAYLOAD",
        "value": "foo"
      },
      {
        "name": "TASK_TIMEOUT",
        "value": "10m"
      },
      {
        "name": "TRACK_INSTANCE_INTERRUPTION",
        "value": "true"
      },
      {
        "name": "AWS_REGION",
        "value": "us-west-2"
      }
    ],
    "dockerLabels": {
    },
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "/ecs/tasque-test-instance-interruption",
        "awslogs-region": "us-west-2",
        "awslogs-stream-prefix": "tasque"
      }
    }
  }
]
EOF
}

resource "aws_cloudwatch_log_group" "log" {
  name = "/ecs/tasque-test-instance-interruption"

  tags = {
    Environment = "development"
    Application = "tasque"
  }
}