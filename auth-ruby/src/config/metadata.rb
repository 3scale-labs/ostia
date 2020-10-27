class Config::Metadata < OpenStruct
  extend Config::BuildSubclass

  class UserInfo < self

    def call(context)
      finder = Config::Identity::OIDC[self[:oidc]]
      id = context.service.identity.find { |id| finder === id } or return
    end
  end
end
