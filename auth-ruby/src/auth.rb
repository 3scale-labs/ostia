# frozen_string_literal: true

require 'envoy/service/auth/v3/external_auth_services_pb'
require 'envoy/service/auth/v2/external_auth_pb'

require 'logger'
require 'pry'

require 'rack/auth/basic'

module RubyLogger
  def logger
    LOGGER
  end

  LOGGER = Logger.new(STDOUT)
end

# GRPC is the general RPC module
module GRPC
  # Inject the noop #logger if no module-level logger method has been injected.
  extend RubyLogger
end

$rate_limits = Hash.new { |hash, key| hash[key] = rand(60) }

Envoy::Service::Auth::V2::AttributeContext::HttpRequest.module_eval do
  def to_env
    headers.to_h.delete_if { |k, _| k.start_with?(':') }.transform_keys { |k| "HTTP_#{k.tr('-', '_').upcase}" }
  end
end

class V2AuthorizationService
  attr_reader :config
  def initialize(config)
    @config = config
  end

  include GRPC::GenericService

  self.marshal_class_method = :encode
  self.unmarshal_class_method = :decode
  self.service_name = 'envoy.service.auth.v2.Authorization'

  # Performs authorization check based on the attributes associated with the
  # incoming request, and returns status `OK` or not `OK`.
  rpc :Check, Envoy::Service::Auth::V2::CheckRequest, Envoy::Service::Auth::V2::CheckResponse

  def check(check, rest)
    puts 'check', check, rest
    puts check.to_json(emit_defaults: true), check.class.name
    auth = Rack::Auth::Basic::Request.new(check.attributes.request.http.to_env)

    Envoy::Service::Auth::V2::CheckResponse.new(
      status: Google::Rpc::Status.new(code: GRPC::Core::StatusCodes::OK),
      ok_response: Envoy::Service::Auth::V2::OkHttpResponse.new(
        headers: [
          Envoy::Api::V2::Core::HeaderValueOption.new(header: Envoy::Api::V2::Core::HeaderValue.new(key: 'x-ext-auth-ratelimit', value: $rate_limits[auth.username].to_s)),
          Envoy::Api::V2::Core::HeaderValueOption.new(header: Envoy::Api::V2::Core::HeaderValue.new(key: 'x-ext-auth-user', value: auth.username))
        ]
      )
    )
  end
end

def main
  port = '0.0.0.0:50051'
  config = Config.new ENV.fetch('CONFIG', 'examples/config.yml')
  s = GRPC::RpcServer.new
  s.add_http2_port(port, :this_port_is_insecure)
  GRPC.logger.info("... running insecurely on #{port}")
  s.handle(V2AuthorizationService.new(config))

  # Runs the server with SIGHUP, SIGINT and SIGQUIT signal handlers to
  #   gracefully shutdown.
  # User could also choose to run server via call to run_till_terminated
  s.run_till_terminated_or_interrupted([1, +'int', +'SIGQUIT'])
end

main if __FILE__ == $0
