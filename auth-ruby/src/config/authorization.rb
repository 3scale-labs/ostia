class Config::Authorization < OpenStruct
  extend Config::BuildSubclass

  class OPA < self
  end

  class JWT < self

  end
end
