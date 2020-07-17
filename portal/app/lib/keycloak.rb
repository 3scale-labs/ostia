module Keycloak
  class Client
    attr_reader :url

    def initialize(server: , realm: , credentials: nil)
      @url = (URI(server) + 'auth/realms/' + "#{realm}/").freeze

      @auth = credentials ? { 'Authorization' => "Bearer #{credentials}" } : { }
      @discovery = Concurrent::Promises.future(&method(:fetch_discovery))
    end

    def http_client
      http = Net::HTTP.new(@url.hostname, @url.port)
      http.use_ssl = @url.scheme == 'https'
      http.set_debug_output($stderr)
      http
    end

    def discovery
      @discovery.value
    end

    protected def fetch_discovery
      http_client.start do |http|
        decode http.request_get(url.path + '.well-known/openid-configuration')
      end
    end

    def create_client(client)
      endpoint = URI(discovery.fetch('registration_endpoint'))
      req =  ClientRegistrationRequest.new ActionController::Parameters.new(client.attributes).permit(ClientRegistrationRequest.attribute_names)

      http_client.start do |http|
        parse_client_registration http.request_post(endpoint.path, req.to_json, @auth.merge('Content-Type' => 'application/json'))
      end
    end

    def read_client(client)
      uri = URI(client.registration_client_uri)

      http_client.start do |http|
        parse_client_registration http.request_get(uri.path, @auth.merge('Authorization' => "Bearer #{client.registration_access_token}"))
      end
    end

    protected

    def parse_client_registration(res)
      params = ActionController::Parameters.new(decode res)
      ClientRegistrationResponse.new params.permit(ClientRegistrationResponse.attribute_names)
    end

    class ClientRegistrationRequest
      include ActiveModel::Model
      include ActiveModel::Attributes
      include ActiveModel::Serializers::JSON

      attribute :client_name, :string
    end

    class ClientRegistrationResponse
      include ActiveModel::Model
      include ActiveModel::Attributes

      attribute :client_id, :string
      attribute :client_secret, :string
      attribute :client_name, :string
      attribute :registration_access_token, :string
      attribute :registration_client_uri, :string
      attribute :client_id_issued_at, :integer
      attribute :client_secret_expires_at, :integer
    end

    def decode(response)
      case response
      when Net::HTTPSuccess
        parse response
      when Net::HTTPClientError
        raise response.inspect
      end
    end

    def parse(response)
      mime = Mime::Type.lookup(response['content-type'] || 'text/plain')
      case mime.symbol
      when :json then JSON.parse(response.body)
      else raise mime.inspect
      end
    end
  end
end
