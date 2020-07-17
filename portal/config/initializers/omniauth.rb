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

Rails.application.config.middleware.use OmniAuth::Builder do
  provider :openid_connect, {
    name: :keycloak,
    scope: %i[openid email profile],
    response_type: :code,
    issuer: 'http://localhost:8080/auth/realms/portal',
    discovery: true,
    client_options: {
      identifier: 'portal',
      secret: '3e32f359-0e77-46d3-a056-7392b69b104f',
      redirect_uri: 'http://localhost:3000/auth/keycloak/callback'
    }
  }
end
