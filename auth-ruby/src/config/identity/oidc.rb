require 'openid_connect'

Module.new do
  ### Monkey patch to keep the desired scheme of the issuer instead forcing it into https

  attr_reader :scheme

  def initialize(uri)
    @scheme = uri.scheme
    super
  end

  def endpoint
    URI::Generic.build(scheme: scheme, host: host, port: port, path: path)
  rescue URI::Error => e
    raise SWD::Exception.new(e.message)
  end

  prepend_features(::OpenIDConnect::Discovery::Provider::Config::Resource)
end

# not in the RFC, but keycloak has it
OpenIDConnect::Discovery::Provider::Config::Response.attr_optional :token_introspection_endpoint

class Config::Identity::OIDC < Config::Identity
  def config
    case config = self[:config]
    when nil
      discover!
    when Hash
      self[:config] = OpenStruct.new(config)
    else
      config
    end
  end

  def discover!
    raise OpenIDConnect::Discovery::DiscoveryFailed unless endpoint

    self[:config] ||= ::OpenIDConnect::Discovery::Provider::Config.discover!(endpoint)
  rescue OpenIDConnect::Discovery::DiscoveryFailed
    self.enabled = false
    nil
  end

  class Token
    def initialize(token)
      @token = token
    end

    def decode!(keys)
      @decoded = ::OpenIDConnect::ResponseObject::IdToken.decode(@token, keys)
    end

    def to_s
      @token
    end

    delegate :as_json, to: :@decoded, allow_nil: true
    alias to_h as_json

    private def method_missing(symbol, *args, &block)
      return super unless @decoded
      @decoded.public_send(symbol, *args, &block)
    end
  end

  def call(context)
    id_token = decode_id_token(context.request)
  rescue JSON::JWK::Set::KidNotFound, JSON::JWS::VerificationFailed => err
    false
  end

  def decode_id_token(req)
    auth = Rack::Auth::AbstractRequest.new(req.attributes.request.http.to_env)

    case auth.scheme
    when 'bearer'
      Token.new(auth.params).tap { |t| t.decode!(public_keys) }
    end
  end

  def public_keys
    config&.jwks
  end

end
