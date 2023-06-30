watches = [
  {
    type = "keyprefix"
    prefix = "auth/"
    handler_type = "http"
    http_handler_config {
      path = "http://host.docker.internal:4000/watch"
      method = "POST"
      timeout = "10s"
      tls_skip_verify = false
    }
  }
]