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

      class OIDC < self
      end
    end

    class Authorization < OpenStruct
      extend BuildSubclass

      class OPA < self
      end

      class JWT < self

      end
    end

    def enabled?
      !!enabled
    end

    def identity
      Array(self[:identity]).map(&Identity.method(:build))
    end

    def authorization
      Array(self[:authorization]).map(&Authorization.method(:build))
    end
  end
end
