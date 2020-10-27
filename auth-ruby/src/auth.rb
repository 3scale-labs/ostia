# frozen_string_literal: true

require 'envoy/service/auth/v3/external_auth_services_pb'
require 'envoy/service/auth/v2/external_auth_pb'

require 'logger'
require 'pry'

require 'rack/auth/basic'
require_relative 'config'

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

  class Context
    attr_reader :identity, :metadata
    attr_reader :request, :service

    def initialize(request, service)
      @request = request
      @service = service
      @identity = {}
      @metadata = {}
    end

    def evaluate!
      proc = ->(obj, result) { result[obj] = obj.call(self) }

      service.identity.each_with_object(identity, &proc)
      service.metadata.each_with_object(metadata, &proc)

      @identity.freeze
      @metadata.freeze
    end

    def valid?
      identity.values.any?
    end
  end

  def check(req, rest)
    GRPC.logger.debug(req.class.name) { req.to_json(emit_defaults: true) }
    host = req.attributes.request.http.host

    case service = config.for_host(host)
    when Config::Service
      # Verify all the stuff!
      # 1. verify identity // call OIDC, verify tokens, mTLS, basic auth
      # 2. load metadata // load plans, metadata, jwt, etc.
      # 3. apply authorization // user written REGO, custom yaml, our generated REGO

      context = Context.new(req, service)

      context.evaluate!

      if context.valid?
        return ok_response(req, service)
      else
        return denied_response("Not authorized")
      end
    end

    denied_response("Service not found")
  end

  protected


  def ok_response(req, service)
    Envoy::Service::Auth::V2::CheckResponse.new(
      status: Google::Rpc::Status.new(code: GRPC::Core::StatusCodes::OK),
      ok_response: Envoy::Service::Auth::V2::OkHttpResponse.new(
        headers: [
          # add headers
        ]
      )
    )
  end

  def denied_response(message)
    Envoy::Service::Auth::V2::CheckResponse.new(
      status: Google::Rpc::Status.new(code: GRPC::Core::StatusCodes::NOT_FOUND),
      denied_response: Envoy::Service::Auth::V2::DeniedHttpResponse.new(
        status: Envoy::Type::HttpStatus.new(code: Envoy::Type::StatusCode::NotFound),
        body: message,
        headers: [
          Envoy::Api::V2::Core::HeaderValueOption.new(header: Envoy::Api::V2::Core::HeaderValue.new(key: 'x-ext-auth-reason', value: "not_found")),
        ]
      )
    )
  end
end

class ResponseInterceptor < GRPC::ServerInterceptor
  def request_response(request:, call:, method:)
    GRPC.logger.info("Received request/response call at method #{method}" \
      " with request #{request} for call #{call}")

    GRPC.logger.info("[GRPC::Ok] (#{method.owner.name}.#{method.name})")
    yield
  end
end

def main
  port = "0.0.0.0:#{ENV.fetch('PORT', 50051)}"
  config = Config.new ENV.fetch('CONFIG', 'examples/config.yml')
  s = GRPC::RpcServer.new(interceptors: [ResponseInterceptor.new])
  s.add_http2_port(port, :this_port_is_insecure)
  GRPC.logger.info("... running insecurely on #{port}")
  s.handle(V2AuthorizationService.new(config))

  # Runs the server with SIGHUP, SIGINT and SIGQUIT signal handlers to
  #   gracefully shutdown.
  # User could also choose to run server via call to run_till_terminated
  s.run_till_terminated_or_interrupted([1, +'int', +'SIGQUIT'])
end

main if __FILE__ == $0
