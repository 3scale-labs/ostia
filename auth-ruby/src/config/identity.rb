class Config::Identity < OpenStruct
  extend Config::BuildSubclass

  def call(req)
    raise NotImplementedError, __method__
  end

  def initialize(hash = {})
    super
    self.enabled = !!config if self.enabled.nil?
  end
end

require_relative 'identity/oidc'
