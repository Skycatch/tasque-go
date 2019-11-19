# data "aws_batch_compute_environment" "generic" {
#   compute_environment_name = "generic-cluster-${var.environment}"
# }

# resource "aws_batch_job_queue" "publish-queue" {
#   name                 = "publish-queue-${var.environment}"
#   state                = "ENABLED"
#   priority             = 1
#   compute_environments = [
#     data.aws_batch_compute_environment.generic.arn
#   ]
# }