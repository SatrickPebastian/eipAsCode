
terraform {
  required_providers {
    rabbitmq = {
      source = "cyrilgdn/rabbitmq"
      version = "1.8.0"
    }
  }
}

provider "rabbitmq" {
  endpoint  = ""
  username  = "admin"
  password  = "password"
}

resource "rabbitmq_queue" "myPipe" {
  name      = "myPipe"
  settings {
    durable    = true
    auto_delete = false
  }
}

