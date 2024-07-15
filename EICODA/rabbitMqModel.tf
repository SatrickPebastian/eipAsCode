
terraform {
  required_providers {
    rabbitmq = {
      source = "cyrilgdn/rabbitmq"
      version = "1.8.0"
    }
  }
}

provider "rabbitmq" {
  endpoint  = "http://localhost:15672"
  username  = "admin"
  password  = "password"
}

resource "rabbitmq_queue" "myPipe" {
  name      = "myPipe"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_exchange" "myPipe" {
  name  = "myPipe"
  vhost = "/"
  settings {
    type        = "topic"
    durable     = true
    auto_delete = false
  }
}

