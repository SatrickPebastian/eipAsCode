
terraform {
  required_providers {
    rabbitmq = {
      source = "0UserName/rabbitmq"
      version = "1.9.1"
    }
  }
}

provider "rabbitmq" {
  endpoint  = "http://localhost:5672"
  username  = "patrick"
  password  = "admin"
}

resource "rabbitmq_queue" "myPipe" {
  name      = "myPipe"
  settings {
    durable    = true
    auto_delete = false
  }
}

