
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

resource "rabbitmq_queue" "reutlingenPipe" {
  name      = "reutlingenPipe"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_queue" "messagePipe" {
  name      = "messagePipe"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}


resource "rabbitmq_exchange" "worldPipe" {
  name  = "worldPipe"
  vhost = "/"
  settings {
    type        = "topic"
    durable     = true
    auto_delete = false
  }
}

