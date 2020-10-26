require 'yaml/store'

class Config
  def initialize(file)
    @store = YAML::Store.new(file)
  end


  def for_host(name)
    @store.transaction(true) do
      case (service = @store.fetch(name, Service::NotFound))
      when Service::NotFound
        Service::NotFound
      else
        Service.new(service)
      end
    end
  end

  class Service < OpenStruct
    NotFound = Object.new

    module BuildSubclass
      class AmbiguousKeysError < StandardError; end

      def build(object)
        name = object.keys.tap { |keys| raise AmbiguousKeysError, keys if keys.size > 1  }.first

        const = constants(false).find { |const|  const.to_s.downcase == name.to_s.downcase }

        const_get(const, false).new(object.fetch(name))
      end
    end
    class Identity < OpenStruct
      extend BuildSubclass

      def call(req)
        raise NotImplementedError, __method__
      end

      def initialize(hash = {})
        super
        self.enabled = !!config if self.enabled.nil?
      end

      class OIDC < self
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
    end

    class Authorization < OpenStruct
      extend BuildSubclass

      class OPA < self
      end

      class JWT < self

      end
    end

    def initialize(hash = nil)
      super

      self.identity = Array(self.identity).map(&Identity.method(:build))
      self.authorization = Array(self.authorization).map(&Authorization.method(:build))
    end

    def enabled?
      !!enabled
    end
  end
end

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
