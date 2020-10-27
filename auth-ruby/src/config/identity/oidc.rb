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

class Config::Identity::OIDC < Config::Identity
  def config
    self[:config] || discover!
  end

  def discover!
    self[:config] ||= ::OpenIDConnect::Discovery::Provider::Config.discover!(endpoint)
  rescue OpenIDConnect::Discovery::DiscoveryFailed
    self.enabled = false
    nil
  end

  def call(req)
    id_token = decode_id_token(req)
  rescue JSON::JWK::Set::KidNotFound
    false
  end

  def decode_id_token(req)
    auth = Rack::Auth::AbstractRequest.new(req.attributes.request.http.to_env)

    case auth.scheme
    when 'bearer'
      ::OpenIDConnect::ResponseObject::IdToken.decode(auth.params, public_keys)
    end
  end

  def public_keys
    config.jwks
  end

end
