# frozen_string_literal: true

require 'net/http'
require 'uri'
require 'json'

class Config::Authorization < OpenStruct
  extend Config::BuildSubclass

  class Response < OpenStruct
    def authorized?
      raise NotImplementedError, __method__
    end
  end

  class OPA < self
    class Response < Config::Authorization::Response
      def authorized?
        result['allow']
      end
    end

    def call(context)
      uri = URI.parse(endpoint)
      auth_request = Net::HTTP::Post.new(uri, 'Content-Type' => 'application/json')
      request, identity, metadata = context.to_h.values_at(:request, :identity, :metadata)
      auth_request.body = { input: request.merge(context: { identity: identity.values.first, metadata: metadata }) }.to_json
      auth_response = Net::HTTP.start(uri.hostname, uri.port) do |http|
        http.request(auth_request)
      end

      case auth_response
      when Net::HTTPOK
        response_json = case auth_response['content-type']
                        when 'application/json'
                          JSON.parse(auth_response.body)
                        else
                          { allowed: true, message: auth_response.body }
                        end
        Response.new(response_json)
      end
    end
  end

  class JWT < self
    class Response < Config::Authorization::Response
      def authorized?
        true
      end
    end

    def call(context)
      Response.new # TODO
    end
  end
end
