class SessionsController < ApplicationController
  def new

  end

  def create
    @auth_hash = auth_hash

    @identity = Identity.find_or_create_by!(uid: auth_hash['uid'], provider: auth_hash['provider']) do |identity|
      identity.build_user
    end

    @user = @identity.user

    @identity.update(credentials: @auth_hash.dig('credentials'))

    session[:identity_gid] = @identity.to_gid_param
  end

  def destroy
    reset_session
  end

  protected

  def auth_hash
    request.env['omniauth.auth']
  end
end
