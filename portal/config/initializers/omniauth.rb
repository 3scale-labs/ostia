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
    scope: %i[openid email profile clients],
    response_type: :code,
    issuer: 'http://localhost:8080/auth/realms/portal',
    discovery: true,
    client_options: {
      identifier: 'portal',
      secret: '69acf2ed-d28a-4e5b-8fe7-0083829d0e0f',
      redirect_uri: 'http://localhost:3000/auth/keycloak/callback'
    }
  }
end
