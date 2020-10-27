package ostia.authz.localhost

import input.attributes.request.http as http_request
import input.context.identity
import input.context.metadata

resource = object.get(input.context, "resource", {})
path_arr = split(trim_left(http_request.path, "/"), "/")

default allow = false

allow {
  http_request.method == "GET"
  path_arr = ["pets"]
}

allow {
  http_request.method == "POST"
  path_arr = ["pets"]
}

allow {
  http_request.method == "GET"
  own_resource
}

allow {
  http_request.method == "PUT"
  own_resource
}

allow {
  http_request.method == "DELETE"
  own_resource
}

allow {
  http_request.method == "GET"
  path_arr = ["pets", "stats"]
  is_admin
}

own_resource {
  some petid
  path_arr = ["pets", petid]
  subject := object.get(identity, "sub", object.get(identity, "username", ""))
  subject == object.get(resource, "owner", "")
}

is_admin {
  metadata.user_info.roles[_] == "admin"
}
