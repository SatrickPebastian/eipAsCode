terraform {
required_providers {
  rabbitmq = {
    source  = "cyrilgdn/rabbitmq"
    version = "~> 1.8"
  }
}
}

provider "rabbitmq" {
endpoint = "http://localhost:15672"
username = "admin"
password = "admin"
}

resource "rabbitmq_queue" "PipeA" {
name       = "PipeA"
vhost      = "/"
settings {
  durable   = true
  auto_delete = false
}
}

resource "rabbitmq_queue" "PipeB" {
name       = "PipeB"
vhost      = "/"
settings {
  durable   = true
  auto_delete = false
}
}

resource "rabbitmq_queue" "PipeC" {
name       = "PipeC"
vhost      = "/"
settings {
  durable   = false
  auto_delete = false
}
}

