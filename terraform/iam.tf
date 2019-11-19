resource "aws_iam_role" "dtm-worker-role" {
  name = "tasque-instance-interruption-test"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "dtm-worker-role-policy" {
  name = "tasque-instance-interruption-test-policy"
  role = "${aws_iam_role.dtm-worker-role.id}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "states:SendTaskSuccess",
          "states:SendTaskFailure",
          "states:SendTaskHeartbeat"
        ],
        "Resource": "*"
      }
    ]
}
EOF
}

resource "aws_iam_role" "dtm-worker-execution-role" {
  name = "tasque-instance-interruption-test-execution"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "dtm-worker-execution-role-policy" {
  name = "tasque-instance-interruption-test-execution-policy"
  role = "${aws_iam_role.dtm-worker-execution-role.id}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        "Resource": "arn:aws:logs:*:*:*"
      }
    ]
}
EOF
}

