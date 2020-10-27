# frozen_string_literal: true

require 'net/http'

class Config::Metadata < OpenStruct
  extend Config::BuildSubclass

  class UserInfo < self
    def call(context)
      finder = Config::Identity::OIDC[self[:oidc]]
      oidc = context.service.identity.find { |id| finder === id } or return
      id = context.identity.fetch(oidc) { return }
      puts id

      uri = URI(oidc.config.token_introspection_endpoint)
      uri.user = client_id
      uri.password = client_secret

      res = Net::HTTP.post_form uri,
                                { 'token' => id, 'token_type_hint' => 'requesting_party_token' }

      case res
      when Net::HTTPOK

        case res['content-type']
        when 'application/json'
          JSON.parse(res.body)
        end
      end
    end
  end
end
