service {
  name = "auth"
  address = "host.docker.internal"
  port = 3000
  tagged_addresses {
    virtual = {
      address = "host.docker.internal"
      port = 3000
    }
  }
  check {
    id = "auth"
    name = "Authentication API GRPC"
    tcp = "host.docker.internal:3000"
    interval = "10s"
    timeout = "1s"
  }
}

